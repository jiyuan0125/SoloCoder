package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"registry/pkg/handler"
	"registry/pkg/healthcheck"
	"registry/pkg/manager"
	"registry/pkg/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	instanceStore := store.NewInstanceStore()
	healthChecker := healthcheck.NewHealthChecker(instanceStore)
	instanceManager := manager.NewInstanceManager(instanceStore, healthChecker)
	httpHandler := handler.NewHandler(instanceStore, healthChecker)

	instanceManager.Start()
	defer instanceManager.Stop()
	defer healthChecker.StopAll()

	server := &http.Server{
		Addr:    ":" + port,
		Handler: httpHandler,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("registry server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-stop
	log.Println("shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}

	log.Println("registry server stopped")
}
