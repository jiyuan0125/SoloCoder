package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-aggregator/internal/db"
	"api-aggregator/internal/server"
)

const (
	defaultDBPath = "./data/aggregator.db"
	defaultPort   = "8080"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	if err := db.Init(dbPath); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	srv := server.New()

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: srv,
	}

	go func() {
		log.Printf("starting server on port %s", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	fmt.Println("shutdown complete")
}
