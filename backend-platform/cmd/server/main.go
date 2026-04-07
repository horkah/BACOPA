package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/horkah/bacopa/backend-platform/internal/db"
	"github.com/horkah/bacopa/backend-platform/internal/handler"
	"github.com/horkah/bacopa/backend-platform/internal/rlgb"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "bacopa.db"
	}

	if err := db.Initialize(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create RLGB client and perform health check with retries
	rlgbClient := rlgb.NewClient()
	for i := 0; i < 10; i++ {
		if err := rlgbClient.Health(); err != nil {
			log.Printf("RLGB health check attempt %d/10 failed: %v", i+1, err)
			if i < 9 {
				time.Sleep(2 * time.Second)
				continue
			}
			log.Printf("WARNING: RLGB service not reachable after 10 attempts, starting anyway")
		} else {
			log.Println("RLGB service is healthy")
			break
		}
	}
	handler.SetRLGBClient(rlgbClient)

	r := mux.NewRouter()

	// CORS middleware
	r.Use(corsMiddleware)

	// Auth routes (public)
	r.HandleFunc("/api/auth/register", handler.RegisterHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/auth/login", handler.LoginHandler).Methods("POST", "OPTIONS")

	// Auth routes (protected)
	authRouter := r.PathPrefix("/api/auth").Subrouter()
	authRouter.Use(handler.AuthMiddleware)
	authRouter.HandleFunc("/me", handler.MeHandler).Methods("GET", "OPTIONS")

	// Game routes (public)
	r.HandleFunc("/api/games/types", handler.GetGameTypes).Methods("GET", "OPTIONS")

	// Game routes (protected)
	gameRouter := r.PathPrefix("/api/games").Subrouter()
	gameRouter.Use(handler.AuthMiddleware)
	gameRouter.HandleFunc("", handler.CreateGame).Methods("POST", "OPTIONS")
	gameRouter.HandleFunc("/lobby", handler.GetLobby).Methods("GET", "OPTIONS")
	gameRouter.HandleFunc("/history", handler.GetGameHistory).Methods("GET", "OPTIONS")
	gameRouter.HandleFunc("/{id}/join", handler.JoinGame).Methods("POST", "OPTIONS")

	// WebSocket (auth via query param)
	r.HandleFunc("/ws", handler.WebSocketHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("Server starting on port %s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
	log.Println("Server stopped")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
