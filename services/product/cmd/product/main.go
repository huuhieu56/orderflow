package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"orderflow/platform/httpx"
	"orderflow/platform/identity"
	"orderflow/product/internal"
	"orderflow/product/internal/cache"
	"orderflow/product/internal/database"
)

func main() {
	ctx := context.Background()
	jwtSecret := os.Getenv("JWT_SECRET")
	db, err := database.Open(ctx, "orderflow_product")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err = database.Migrate(ctx, db); err != nil {
		log.Fatal(err)
	}
	c := cache.New(os.Getenv("REDIS_ADDR"))
	defer c.Close()
	if err = c.Ping(ctx); err != nil {
		log.Printf("redis unavailable: %v", err)
	}
	productRepository := product.NewRepository(db)
	productService := product.NewService(productRepository, c)
	h := product.NewHandler(productService)
	mux := http.NewServeMux()
	adminOnly := func(next http.Handler) http.Handler {
		return identity.Authenticate(
			jwtSecret,
			identity.RequireRole("admin", next),
		)
	}
	mux.Handle("POST /api/v1/products", adminOnly(http.HandlerFunc(h.Create)))
	mux.HandleFunc("GET /api/v1/products", h.List)
	mux.HandleFunc("GET /api/v1/products/{id}", h.GetByID)
	mux.Handle(
		"PATCH /api/v1/products/{id}",
		adminOnly(http.HandlerFunc(h.Update)),
	)
	mux.HandleFunc("GET /health", health)
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func health(w http.ResponseWriter, _ *http.Request) {
	httpx.Success(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
