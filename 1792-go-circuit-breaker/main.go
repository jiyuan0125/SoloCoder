package main

import (
	"log"
	"os"
	"time"

	"circuit-breaker/circuitbreaker"
	"circuit-breaker/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	manager := circuitbreaker.NewManager()

	setupDemoServices(manager)

	h := handler.NewHandler(manager)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := gin.Default()

	r.GET("/health", h.HealthCheck)
	r.GET("/status", h.GetStatus)
	r.POST("/services/:service/reset", h.ResetService)
	r.POST("/simulate", h.SimulateCall)

	log.Printf("[INFO] Starting server on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("[ERROR] Failed to start server: %v", err)
	}
}

func setupDemoServices(manager *circuitbreaker.Manager) {
	manager.Register(&circuitbreaker.Config{
		ServiceName:      "payment-service",
		FailureThreshold: 5,
		OpenDuration:     30 * time.Second,
		HalfOpenTimeout:  5 * time.Second,
		OnStateChange: func(serviceName string, from, to circuitbreaker.State) {
			log.Printf("[ALERT] Service '%s' state changed: %s -> %s", serviceName, from, to)
		},
	})

	manager.Register(&circuitbreaker.Config{
		ServiceName:      "order-service",
		FailureThreshold: 3,
		OpenDuration:     10 * time.Second,
		HalfOpenTimeout:  5 * time.Second,
		OnStateChange: func(serviceName string, from, to circuitbreaker.State) {
			log.Printf("[ALERT] Service '%s' state changed: %s -> %s", serviceName, from, to)
		},
	})

	manager.Register(&circuitbreaker.Config{
		ServiceName:      "user-service",
		FailureThreshold: 5,
		OpenDuration:     30 * time.Second,
		HalfOpenTimeout:  5 * time.Second,
		OnStateChange: func(serviceName string, from, to circuitbreaker.State) {
			log.Printf("[ALERT] Service '%s' state changed: %s -> %s", serviceName, from, to)
		},
	})
}
