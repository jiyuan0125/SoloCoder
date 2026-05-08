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

	"jwt-service/pkg/jwtcore"
)

func main() {
	serverConfig, err := LoadServerConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	blacklist := jwtcore.NewBlacklist(5 * time.Minute)
	defer blacklist.Stop()

	service := jwtcore.NewService(serverConfig.JWTConfig, blacklist)
	handler := NewHTTPHandler(service, blacklist)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", serverConfig.Port),
		Handler: mux,
	}

	go func() {
		log.Printf("JWT service starting on port %d (algorithm: %s)", serverConfig.Port, serverConfig.JWTConfig.Algorithm)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
