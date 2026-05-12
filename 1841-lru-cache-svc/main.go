package main

import (
	"net/http"
	"os"
	"strconv"

	"lru-cache-svc/cache"

	"github.com/gin-gonic/gin"
)

type SetRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
	TTL   *int   `json:"ttl"`
}

type ConfigRequest struct {
	Capacity   int `json:"capacity" binding:"required"`
	DefaultTTL int `json:"default_ttl" binding:"required"`
}

type StatsResponse struct {
	HitRate       float64 `json:"hit_rate"`
	CurrentSize   int     `json:"current_size"`
	TotalQueries  uint64  `json:"total_queries"`
}

func main() {
	c := cache.New(1000, 60)

	r := gin.Default()

	r.POST("/cache", func(ctx *gin.Context) {
		var req SetRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Set(req.Key, req.Value, req.TTL)
		ctx.Status(http.StatusNoContent)
	})

	r.GET("/cache/:key", func(ctx *gin.Context) {
		key := ctx.Param("key")
		value, ok := c.Get(key)
		if !ok {
			ctx.Status(http.StatusNotFound)
			return
		}
		ctx.String(http.StatusOK, value)
	})

	r.DELETE("/cache/:key", func(ctx *gin.Context) {
		key := ctx.Param("key")
		c.Delete(key)
		ctx.Status(http.StatusNoContent)
	})

	r.DELETE("/cache", func(ctx *gin.Context) {
		c.Flush()
		ctx.Status(http.StatusNoContent)
	})

	r.GET("/cache/stats", func(ctx *gin.Context) {
		stats := c.Stats()
		resp := StatsResponse{
			CurrentSize:  stats.CurrentSize,
			TotalQueries: stats.TotalCount,
		}
		if stats.TotalCount > 0 {
			resp.HitRate = float64(stats.HitCount) / float64(stats.TotalCount)
		} else {
			resp.HitRate = 0.0
		}
		ctx.JSON(http.StatusOK, resp)
	})

	r.PUT("/cache/config", func(ctx *gin.Context) {
		var req ConfigRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.SetConfig(req.Capacity, req.DefaultTTL)
		ctx.Status(http.StatusNoContent)
	})

	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		if _, err := strconv.Atoi(envPort); err == nil {
			port = envPort
		}
	}

	r.Run(":" + port)
}
