package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rest_go_toko/config"
	"rest_go_toko/database"
	"rest_go_toko/routes"
)

func main() {
	log.Println("Starting REST API Toko Pintar...")

	// 1. Load configurations from .env/environment
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Critical Configuration Error: %v", err)
	}

	// 2. Initialize Firebird connection pool
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Critical Database Connection Error: %v", err)
	}
	defer func() {
		log.Println("Closing database connection pool...")
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection pool: %v", err)
		} else {
			log.Println("Database connection pool closed successfully")
		}
	}()

	// 3. Set up routes
	router := routes.SetupRouter(db)

	// 4. Configure HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	// 5. Start HTTP server in a goroutine
	go func() {
		log.Printf("Server is running on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server ListenAndServe failed: %v", err)
		}
	}()

	// 6. Graceful shutdown handler listening for SIGINT and SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutdown signal received, shutting down server gracefully...")

	// Define timeout duration for waiting outstanding requests to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly")
}
