package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"health-checker/pkg/handler"
	"health-checker/pkg/manager"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	m := manager.New()
	h := handler.New(m)

	r := gin.Default()

	r.GET("/health", h.GetHealth)
	r.POST("/components", h.RegisterComponent)
	r.DELETE("/components/:name", h.DeregisterComponent)
	r.GET("/components/:name", h.GetComponent)

	log.Printf("Health checker server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
