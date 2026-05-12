package main

import (
	"net/http"
	"os"

	"rate-limiter/internal/handler"
	"rate-limiter/internal/limiter"

	"github.com/gin-gonic/gin"
)

func main() {
	mgr := limiter.NewManager()
	defer mgr.Close()

	r := gin.Default()

	r.Use(limiter.Middleware(mgr))

	handler.SetupAdmin(r, mgr)

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	r.GET("/api/*any", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
