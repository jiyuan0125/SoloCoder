package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type bucket struct {
	count     int
	windowStart time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*bucket),
	}
}

func (rl *RateLimiter) Allow(keyID string, limit int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.buckets[keyID]

	if !exists {
		rl.buckets[keyID] = &bucket{
			count:       1,
			windowStart: now,
		}
		return true
	}

	if now.Sub(b.windowStart) >= time.Minute {
		b.count = 1
		b.windowStart = now
		return true
	}

	if b.count >= limit {
		return false
	}

	b.count++
	return true
}

func RateLimit(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		keyID, ok := GetKeyID(c)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
			return
		}

		limit, ok := GetRateLimit(c)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"error": "rate limit not found"})
			return
		}

		if !limiter.Allow(keyID, limit) {
			c.AbortWithStatusJSON(429, gin.H{"error": "rate limit exceeded"})
			return
		}

		c.Next()
	}
}
