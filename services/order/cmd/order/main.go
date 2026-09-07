package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"orderflow/order/internal"
	"orderflow/order/internal/database"
	"orderflow/order/internal/messaging"
	"orderflow/platform/httpx"
	"orderflow/platform/identity"
)

func main() {
	ctx := context.Background()
	jwtSecret := os.Getenv("JWT_SECRET")
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	topic := os.Getenv("KAFKA_ORDER_TOPIC")
	productURL := os.Getenv("PRODUCT_SERVICE_URL")
	db, err := database.Open(ctx, "orderflow_order")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err = database.Migrate(ctx, db); err != nil {
		log.Fatal(err)
	}
	producer := messaging.NewProducer(brokers)
	defer producer.Close()
	orderRepository := order.NewRepository(db, topic)
	productClient := order.NewProductClient(productURL)
	orderService := order.NewService(orderRepository, productClient)
	go func() {
		if err := order.NewOutboxWorker(orderRepository, producer).Run(ctx); err != nil {
			log.Printf("outbox worker stopped: %v", err)
		}
	}()
	h := order.NewHandler(orderService)
	auth := func(next http.Handler) http.Handler {
		return identity.Authenticate(jwtSecret, next)
	}
	mux := http.NewServeMux()
	mux.Handle("POST /api/v1/orders", auth(http.HandlerFunc(h.Create)))
	mux.Handle("GET /api/v1/orders", auth(http.HandlerFunc(h.ListByUser)))
	mux.Handle("GET /api/v1/orders/{id}", auth(http.HandlerFunc(h.GetByID)))
	mux.Handle("DELETE /api/v1/orders/{id}", auth(http.HandlerFunc(h.Cancel)))
	mux.HandleFunc("GET /health", health)
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func health(w http.ResponseWriter, _ *http.Request) {
	httpx.Success(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
