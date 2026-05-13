package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"gateway/internal/buffer"
	"gateway/internal/handler"
	"gateway/internal/middleware"
	"gateway/internal/store"
)

const logBufferCapacity = 10000

func main() {
	keyStore := store.NewKeyStore()
	rateLimiter := middleware.NewRateLimiter()
	logBuffer := buffer.NewRingBuffer(logBufferCapacity)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(gin.Recovery())

	admin := r.Group("/admin")
	{
		admin.POST("/keys", handler.CreateKey(keyStore))
		admin.DELETE("/keys/:kid", handler.RevokeKey(keyStore))
		admin.GET("/keys", handler.ListKeys(keyStore))
	}

	apiHandlers := gin.HandlersChain{
		middleware.BodyLimit(),
		middleware.Auth(keyStore, nil),
		middleware.RateLimit(rateLimiter),
		middleware.Logger(logBuffer),
		gin.HandlerFunc(func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "request passed through gateway",
				"path":    c.Request.URL.Path,
			})
		}),
	}

	r.NoRoute(apiHandlers...)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8400"
	}

	log.Printf("Gateway starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
