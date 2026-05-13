package main

import (
	"net/http"
	"os"
	"strconv"

	"cache-proxy/cache"

	"github.com/gin-gonic/gin"
)

type setRequest struct {
	Key   string `json:"key" form:"key"`
	Value string `json:"value" form:"value"`
	TTL   *int   `json:"ttl,omitempty" form:"ttl"`
}

type configRequest struct {
	Capacity   *int `json:"capacity,omitempty"`
	DefaultTTL *int `json:"default_ttl,omitempty"`
}

func main() {
	r := gin.Default()

	defaultTTL := 3600
	if ttlEnv := os.Getenv("DEFAULT_TTL"); ttlEnv != "" {
		if t, err := strconv.Atoi(ttlEnv); err == nil && t > 0 {
			defaultTTL = t
		}
	}

	capacity := 10000
	if capEnv := os.Getenv("CAPACITY"); capEnv != "" {
		if c, err := strconv.Atoi(capEnv); err == nil && c > 0 {
			capacity = c
		}
	}

	c := cache.New(capacity, defaultTTL)

	r.POST("/cache/set", func(ctx *gin.Context) {
		var req setRequest
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.Key == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
			return
		}
		c.Set(req.Key, req.Value, req.TTL)
		ctx.JSON(http.StatusOK, gin.H{"ok": true})
	})

	r.GET("/cache/get", func(ctx *gin.Context) {
		key := ctx.Query("key")
		if key == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
			return
		}
		value, ok := c.Get(key)
		if !ok {
			ctx.JSON(http.StatusOK, gin.H{"value": nil})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"value": value})
	})

	r.DELETE("/cache/delete", func(ctx *gin.Context) {
		key := ctx.Query("key")
		if key == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
			return
		}
		deleted := c.Delete(key)
		ctx.JSON(http.StatusOK, gin.H{"deleted": deleted})
	})

	r.POST("/cache/flush", func(ctx *gin.Context) {
		c.Flush()
		ctx.JSON(http.StatusOK, gin.H{"ok": true})
	})

	r.GET("/cache/stats", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, c.Stats())
	})

	r.PUT("/cache/config", func(ctx *gin.Context) {
		var req configRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.Capacity != nil {
			if *req.Capacity < 0 {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "capacity must be non-negative"})
				return
			}
			c.SetCapacity(*req.Capacity)
		}
		if req.DefaultTTL != nil {
			if *req.DefaultTTL < 0 {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "default_ttl must be non-negative"})
				return
			}
			c.SetDefaultTTL(*req.DefaultTTL)
		}
		ctx.JSON(http.StatusOK, c.Stats())
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8801"
	}

	r.Run(":" + port)
}
