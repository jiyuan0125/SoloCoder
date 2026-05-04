package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"luggage-tracking/server"
)

const (
	port         = ":8080"
	anomalyCheckInterval = 5 * time.Minute
)

func main() {
	store := server.NewStore()
	service := server.NewService(store)
	handler := server.NewHandler(service)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go startAnomalyChecker(ctx, service)

	go func() {
		log.Printf("Server starting on port %s...", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server stopped gracefully")
}

func startAnomalyChecker(ctx context.Context, service *server.Service) {
	ticker := time.NewTicker(anomalyCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			service.CheckAnomalies()
		case <-ctx.Done():
			return
		}
	}
}
