package main

import (
	"fmt"
	"log"
	"net/http"
	"orderflow/internal/auth"
	"orderflow/internal/config"
	"orderflow/internal/database"
	"orderflow/internal/order"
	"orderflow/internal/product"
)

func main() {
	cfg := config.Load()

	// Connect DB
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("cannot connect to database: %v", err)
	}
	log.Println("Connected to database")
	defer db.Close()

	// Migrate DB
	if err := database.RunMigrations(db, "internal/database/migrations"); err != nil {
		log.Fatalf("cannot run migrations: %v", err)
	}

	// Auth
	tokenSvc := auth.NewTokenService(cfg.JWTSecret, cfg.JWTExpiration, cfg.JWTRefreshExpiration)
	authRepo := auth.NewRepository(db)
	authSvc := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authSvc, tokenSvc)

	// Product
	productRepo := product.NewRepository(db)
	productSvc := product.NewService(productRepo)
	productHandler := product.NewHandler(productSvc)

	// Order
	orderRepo := order.NewRepository(db)
	orderSvc := order.NewService(orderRepo, productRepo)
	orderHandler := order.NewHandler(orderSvc)

	mux := http.NewServeMux()

	// Auth
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.Handle("GET /api/v1/auth/me", tokenSvc.AuthMiddleware(http.HandlerFunc(authHandler.Me)))
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.Refresh)
	mux.Handle("POST /api/v1/auth/logout", tokenSvc.AuthMiddleware(http.HandlerFunc(authHandler.Logout)))

	// Product
	mux.HandleFunc("POST /api/v1/products", productHandler.Create)
	mux.HandleFunc("GET /api/v1/products", productHandler.List)

	// Order
	mux.Handle("POST /api/v1/orders", tokenSvc.AuthMiddleware(http.HandlerFunc(orderHandler.Create)))
	mux.Handle("GET /api/v1/orders", tokenSvc.AuthMiddleware(http.HandlerFunc(orderHandler.ListByUser)))

	// Health
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	addr := ":" + cfg.Port
	log.Printf("OrderFlow server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
