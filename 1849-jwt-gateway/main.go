package main

import (
	"log"
	"net/http"
	"os"

	"jwt-gateway/handlers"
	"jwt-gateway/middleware"
	"jwt-gateway/proxy"
	"jwt-gateway/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	backends := os.Getenv("BACKENDS")
	if backends == "" {
		backends = "http://localhost:8081,http://localhost:8082"
	}

	rotationPeriod := os.Getenv("ROTATION_PERIOD")
	if rotationPeriod == "" {
		rotationPeriod = "3600"
	}

	keyManager := storage.NewKeyManager(storage.MustParseDuration(rotationPeriod))
	auditLog := storage.NewAuditLog()
	proxyManager := proxy.NewProxyManager(storage.ParseBackends(backends))

	r := gin.Default()

	r.POST("/keys/generate", handlers.GenerateKey(keyManager))
	r.POST("/keys/rotate", handlers.RotateKey(keyManager))
	r.GET("/keys", handlers.ListKeys(keyManager))

	adminGroup := r.Group("/admin")
	adminGroup.Use(middleware.Auth(keyManager, auditLog), middleware.RequireAdmin(auditLog))
	{
		adminGroup.GET("/audit", handlers.GetAuditLogs(auditLog))
	}

	r.NoRoute(middleware.Auth(keyManager, auditLog), proxyManager.Proxy)

	log.Printf("JWT Gateway starting on port %s", port)
	log.Printf("Backends: %v", storage.ParseBackends(backends))

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
