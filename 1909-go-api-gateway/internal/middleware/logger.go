package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"gateway/internal/buffer"
)

func Logger(logBuffer *buffer.RingBuffer) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		entry := buffer.LogEntry{
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			Status:    c.Writer.Status(),
			Duration:  duration.Milliseconds(),
			Timestamp: start,
		}
		logBuffer.Push(entry)
	}
}
