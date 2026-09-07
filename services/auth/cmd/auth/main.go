package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"orderflow/auth/internal"
	"orderflow/auth/internal/cache"
	"orderflow/auth/internal/database"
	"orderflow/platform/httpx"
	"orderflow/platform/identity"
)

func main() {
	ctx := context.Background()
	jwtSecret := os.Getenv("JWT_SECRET")
	db, err := database.Open(ctx, "orderflow_auth")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := database.Migrate(ctx, db); err != nil {
		log.Fatal(err)
	}

	redisClient := cache.New(os.Getenv("REDIS_ADDR"))
	defer redisClient.Close()
	if err := redisClient.Ping(ctx); err != nil {
		log.Printf("redis unavailable: %v", err)
	}
	tokens := auth.NewTokenService(jwtSecret, time.Hour, 7*24*time.Hour)
	service := auth.NewService(auth.NewRepository(db), redisClient, tokens)
	initialAdminEmail := os.Getenv("INITIAL_ADMIN_EMAIL")
	initialAdminPassword := os.Getenv("INITIAL_ADMIN_PASSWORD")
	if err := service.EnsureInitialAdmin(ctx, initialAdminEmail, initialAdminPassword); err != nil {
		log.Fatal(err)
	}
	handler := auth.NewHandler(service)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/auth/register", handler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", handler.Login)
	mux.Handle("GET /api/v1/auth/me", identity.Authenticate(jwtSecret, http.HandlerFunc(handler.Me)))
	mux.HandleFunc("POST /api/v1/auth/refresh", handler.Refresh)
	mux.Handle("POST /api/v1/auth/logout", identity.Authenticate(jwtSecret, http.HandlerFunc(handler.Logout)))
	mux.HandleFunc("GET /health", health)
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func health(w http.ResponseWriter, _ *http.Request) {
	httpx.Success(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
