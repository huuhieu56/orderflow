package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"orderflow/notification/internal"
	"orderflow/notification/internal/database"
	"orderflow/notification/internal/messaging"
	"orderflow/platform/httpx"
	"orderflow/platform/identity"
)

func main() {
	ctx := context.Background()
	jwtSecret := os.Getenv("JWT_SECRET")
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	topic := os.Getenv("KAFKA_ORDER_TOPIC")
	consumerGroup := os.Getenv("KAFKA_CONSUMER_GROUP")
	db, err := database.Open(ctx, "orderflow_notification")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err = database.Migrate(ctx, db); err != nil {
		log.Fatal(err)
	}
	svc := notification.NewService(notification.NewRepository(db))
	consumer := messaging.NewConsumer(brokers, topic, consumerGroup)
	defer consumer.Close()
	go func() {
		if err := notification.NewEventConsumer(svc, consumer).Run(ctx); err != nil {
			log.Printf("consumer stopped: %v", err)
		}
	}()
	h := notification.NewHandler(svc)
	auth := func(next http.Handler) http.Handler {
		return identity.Authenticate(jwtSecret, next)
	}
	mux := http.NewServeMux()
	mux.Handle("GET /api/v1/notifications", auth(http.HandlerFunc(h.List)))
	mux.Handle("PATCH /api/v1/notifications/{id}/read", auth(http.HandlerFunc(h.MarkRead)))
	mux.HandleFunc("GET /health", health)
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func health(w http.ResponseWriter, _ *http.Request) {
	httpx.Success(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
