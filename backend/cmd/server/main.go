package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devmuggs/youtube-music-continuity/backend/internal/auth"
	"github.com/devmuggs/youtube-music-continuity/backend/internal/database"
	"github.com/devmuggs/youtube-music-continuity/backend/internal/websocket"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// Initialize database
	db, err := database.NewDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Create router
	r := mux.NewRouter()

	// CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"chrome-extension://*", "http://localhost:*", "https://music.youtube.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Auth routes
	authHandler := auth.NewHandler(db)
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST")

	// Session routes (protected)
	api.HandleFunc("/session", auth.AuthMiddleware(authHandler.GetSession)).Methods("GET")
	api.HandleFunc("/session", auth.AuthMiddleware(authHandler.UpdateSession)).Methods("PUT")

	// WebSocket endpoint
	api.HandleFunc("/ws", websocket.HandleWebSocket(hub, db))

	// Apply CORS
	handler := c.Handler(r)

	// Server configuration
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Println("Server starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
