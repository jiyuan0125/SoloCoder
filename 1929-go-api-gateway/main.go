package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	store := NewStore()
	executor := NewExecutor()
	handler := NewHandler(store, executor)

	r := gin.Default()

	r.POST("/backends", handler.CreateBackend)
	r.GET("/backends", handler.ListBackends)

	r.POST("/plans", handler.CreatePlan)
	r.GET("/plans", handler.ListPlans)
	r.GET("/plans/:id", handler.GetPlan)
	r.POST("/plans/:id/execute", handler.ExecutePlan)

	log.Printf("API Gateway starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
