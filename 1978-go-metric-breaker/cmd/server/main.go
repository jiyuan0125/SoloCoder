package main

import (
	"breaker/internal/api"
	"breaker/internal/breaker"
	"breaker/internal/middleware"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	manager := breaker.NewManager()
	handler := api.NewHandler(manager)
	cbMiddleware := middleware.NewCircuitBreakerMiddleware(manager)

	r := gin.Default()

	_ = manager.GetOrCreate("default", &breaker.CircuitBreakerConfig{
		WindowDuration: time.Second * 10,
		CooldownPeriod: time.Second * 30,
		HalfOpenLimit:  5,
	})

	r.GET("/breakers", handler.ListBreakers)
	r.GET("/breakers/:name/status", handler.GetStatus)
	r.PUT("/breakers/:name/metrics", handler.UpdateMetrics)
	r.POST("/breakers/:name/metrics/custom", handler.ReportCustomMetric)

	r.GET("/ping", cbMiddleware.Middleware("default"), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	r.GET("/slow", cbMiddleware.Middleware("default"), func(c *gin.Context) {
		time.Sleep(time.Millisecond * 100)
		c.JSON(200, gin.H{"message": "slow response"})
	})

	r.GET("/error", cbMiddleware.Middleware("default"), func(c *gin.Context) {
		c.JSON(500, gin.H{"error": "internal error"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "9805"
	}

	r.Run(":" + port)
}
