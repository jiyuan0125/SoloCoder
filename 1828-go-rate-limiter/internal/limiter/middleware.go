package limiter

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func Middleware(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		bucket := mgr.GetBucket(path)
		if bucket == nil {
			c.Next()
			return
		}

		allowed, _ := bucket.Consume()
		if !allowed {
			stats := bucket.Stats(path)
			retryAfter := calcRetryAfter(stats)
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":      "too many requests",
				"path":       path,
				"retryAfter": retryAfter,
			})
			return
		}

		c.Next()
	}
}

func calcRetryAfter(stats *BucketStats) int {
	needed := 1.0 - stats.Remaining
	if needed <= 0 {
		return 1
	}
	seconds := int(needed/stats.Rate) + 1
	if seconds < 1 {
		seconds = 1
	}
	if stats.WaitTimeout > 0 {
		fromTimeout := int(stats.WaitTimeout.Seconds()) + 1
		if fromTimeout > seconds {
			return fromTimeout
		}
	}
	return seconds
}

const DefaultWaitTimeout = 5 * time.Second
