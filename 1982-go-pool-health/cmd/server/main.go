package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"pool-health/internal/api"
	"pool-health/internal/pool"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	manager := pool.NewManager()
	handler := api.NewHandler(manager)
	router := api.SetupRouter(handler)

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Starting server on port %s...", port)
		if err := router.Run(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-stopChan
	log.Println("Shutting down server...")
	manager.CloseAll()
	log.Println("Server shutdown complete")
}
