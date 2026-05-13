package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"leaderboard/internal/api"
	"leaderboard/internal/database"
	"leaderboard/internal/leaderboard"
	"leaderboard/internal/planner"
)

func main() {
	dbPath := "./leaderboard.db"
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := database.Init(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.DB.Close()

	leaderboardService := leaderboard.NewService()
	plannerService := planner.NewService()
	handler := api.NewHandler(leaderboardService, plannerService)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	stopChan := make(chan struct{})
	go startPeriodicCleanup(leaderboardService, stopChan)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Server starting on :8080...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-sigChan
	log.Println("Shutting down...")
	close(stopChan)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	
	log.Println("Server stopped gracefully")
}

func startPeriodicCleanup(service *leaderboard.Service, stopChan <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	cleanupTicker := time.NewTicker(24 * time.Hour)
	defer cleanupTicker.Stop()

	log.Println("Periodic cleanup scheduler started")

	for {
		select {
		case <-stopChan:
			log.Println("Periodic cleanup scheduler stopped")
			return
		case <-cleanupTicker.C:
			log.Println("Running periodic cleanup...")
			if err := service.CleanupExpired(context.Background()); err != nil {
				log.Printf("Cleanup failed: %v", err)
			} else {
				log.Println("Cleanup completed successfully")
			}
		}
	}
}
