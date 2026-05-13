package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const maxBodySize = 1 * 1024 * 1024

func BodyLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodySize)
		c.Next()
	}
}
