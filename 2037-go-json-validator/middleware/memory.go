package middleware

import (
	"net/http"

	"json-validator/memory"

	"github.com/gin-gonic/gin"
)

func MemoryLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		monitor := memory.GetMonitor()
		if monitor.IsOverLimit() {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "内存使用超过限制",
				"code":  413,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
