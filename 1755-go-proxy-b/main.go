package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"proxy-service/config"
	"proxy-service/proxy"
)

func main() {
	cfgPath := "config.json"
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		cfgPath = path
	}

	cfg := config.LoadConfig(cfgPath)
	log.Printf("Loaded %d backends from config", len(cfg.Backends))

	bm := proxy.NewBackendManager(cfg)
	bm.HealthCheck.Start()
	log.Println("Health check started")

	ps := proxy.NewProxyService(bm)

	r := gin.Default()

	adminGroup := r.Group(cfg.ManagementPath)
	{
		adminGroup.GET("/status", ps.GetStatus)
	}

	r.NoRoute(ps.Forward)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting proxy server on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
