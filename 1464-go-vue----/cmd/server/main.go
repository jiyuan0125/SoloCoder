package main

import (
	"context"
	"flag"
	"firemanagement/internal/core"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		flagPort := flag.String("port", "8080", "Server port")
		flag.Parse()
		port = *flagPort
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	return port
}

func main() {
	port := getPort()

	service := core.NewService()
	service.Start()
	defer service.Stop()

	handler := NewHandler(service)
	router := handler.SetupRoutes()

	server := &http.Server{
		Addr:    port,
		Handler: router,
	}

	go func() {
		log.Printf("Fire Management Server starting on %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
