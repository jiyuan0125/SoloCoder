package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"circuit-breaker/circuit"
	"circuit-breaker/handlers"
	"circuit-breaker/subscribers"
)

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return ":" + port
}

func main() {
	breakerManager := circuit.NewBreakerManager()
	subscriberManager := subscribers.NewManager()
	
	server := handlers.NewServer(breakerManager, subscriberManager)
	
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	
	server.RegisterRoutes(r)
	
	addr := getPort()
	log.Printf("Circuit Breaker API server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
