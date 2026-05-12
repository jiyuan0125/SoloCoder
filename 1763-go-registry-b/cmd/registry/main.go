package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	httpinternal "registry/internal/http"
	"registry/internal/registry"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	reg := registry.New()
	handler := httpinternal.NewHandler(reg)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api")
	{
		api.POST("/register", handler.Register)
		api.POST("/heartbeat", handler.Heartbeat)
		api.POST("/unregister", handler.Unregister)
		api.GET("/services", handler.GetAllServices)
		api.GET("/services/:service_name", handler.GetService)
		api.GET("/subscribe", handler.Subscribe)
	}

	log.Printf("Service Registry listening on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
