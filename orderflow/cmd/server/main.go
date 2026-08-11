package main
import (
	"fmt"
	"log"
	"net/http"
	"orderflow/internal/config"
	"orderflow/internal/database"
	"orderflow/internal/auth"
)

func main() {
	cfg := config.Load()

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
	tokenSvc := auth.NewTokenService(cfg.JWTSecret, cfg.JWTExpiration)
	authRepo := auth.NewRepository(db)
	authSvc := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authSvc, tokenSvc)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.Handle("GET /api/v1/auth/me", tokenSvc.AuthMiddleware(http.HandlerFunc(authHandler.Me)))

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