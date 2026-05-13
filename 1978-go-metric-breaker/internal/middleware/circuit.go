package middleware

import (
	"breaker/internal/breaker"
	"time"

	"github.com/gin-gonic/gin"
)

type CircuitBreakerMiddleware struct {
	manager *breaker.Manager
}

func NewCircuitBreakerMiddleware(manager *breaker.Manager) *CircuitBreakerMiddleware {
	return &CircuitBreakerMiddleware{
		manager: manager,
	}
}

func (m *CircuitBreakerMiddleware) Middleware(breakerName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cb := m.manager.GetOrCreate(breakerName, nil)

		if !cb.Allow() {
			code, body := cb.GetFallback()
			c.AbortWithStatusJSON(code, body)
			return
		}

		start := time.Now()
		c.Next()
		end := time.Now()

		statusCode := c.Writer.Status()
		cb.Record(start, end, statusCode)
	}
}
