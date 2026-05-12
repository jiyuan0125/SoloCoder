package handler

import (
	"net/http"
	"time"

	"rate-limiter/internal/limiter"

	"github.com/gin-gonic/gin"
)

type RuleRequest struct {
	Path        string  `json:"path" binding:"required"`
	Capacity    float64 `json:"capacity" binding:"required"`
	Rate        float64 `json:"rate" binding:"required"`
	WaitTimeout *int    `json:"waitTimeoutSeconds"`
}

func SetupAdmin(r *gin.Engine, mgr *limiter.Manager) {
	admin := r.Group("/admin/limiter")
	{
		admin.POST("/rules", SetRule(mgr))
		admin.DELETE("/rules/*path", DeleteRule(mgr))
		admin.GET("/rules", ListRules(mgr))
	}
}

func SetRule(mgr *limiter.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RuleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request body: " + err.Error(),
			})
			return
		}

		timeout := limiter.DefaultWaitTimeout
		if req.WaitTimeout != nil {
			timeout = time.Duration(*req.WaitTimeout) * time.Second
		}

		cfg := &limiter.BucketConfig{
			Path:        req.Path,
			Capacity:    req.Capacity,
			Rate:        req.Rate,
			WaitTimeout: timeout,
		}

		if err := mgr.SetRule(cfg); err != nil {
			if ce, ok := err.(*limiter.ConfigError); ok {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":  ce.Error(),
					"field":  ce.Field,
					"reason": ce.Reason,
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "rule set",
			"path":    req.Path,
		})
	}
}

func DeleteRule(mgr *limiter.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Param("path")
		if path == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "path is required",
			})
			return
		}

		if mgr.DeleteRule(path) {
			c.JSON(http.StatusOK, gin.H{
				"message": "rule deleted",
				"path":    path,
			})
			return
		}

		c.JSON(http.StatusNotFound, gin.H{
			"error": "rule not found",
			"path":  path,
		})
	}
}

func ListRules(mgr *limiter.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := mgr.ListStats()
		rules := make([]gin.H, 0, len(stats))
		for _, s := range stats {
			rules = append(rules, gin.H{
				"path":           s.Path,
				"capacity":       s.Capacity,
				"rate":           s.Rate,
				"waitTimeout":    s.WaitTimeout.Seconds(),
				"remaining":      s.Remaining,
				"lastRefill":     s.LastRefill.Format(time.RFC3339Nano),
				"currentQueue":   s.CurrentQueue,
				"totalQueued":    s.TotalQueued,
				"totalRejected":  s.TotalRejected,
			})
		}
		c.JSON(http.StatusOK, gin.H{
			"rules": rules,
		})
	}
}
