package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/apigateway/internal/api"
	"github.com/example/apigateway/internal/config"
	"github.com/example/apigateway/internal/gateway"
	"github.com/example/apigateway/internal/middleware"
	"github.com/example/apigateway/internal/storage"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := storage.GetStorage()
	defer store.Close()

	mgr := config.GetConfigManager()
	_ = mgr

	go middleware.StartRateLimitCleanup(ctx)

	mux := http.NewServeMux()

	api.RegisterRoutes(mux)

	gatewayHandler := gateway.NewGateway()
	mux.Handle("/", gatewayHandler)

	server := &http.Server{
		Addr:         ":9102",
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("API Gateway starting on port 9102...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down API Gateway...")

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}

	log.Println("API Gateway stopped gracefully")
}
