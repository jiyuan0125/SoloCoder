package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				debug.PrintStack()
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "服务器内部错误",
					"code":  500,
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
