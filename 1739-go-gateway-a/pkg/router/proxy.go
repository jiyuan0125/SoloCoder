package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (m *Manager) ProxyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		entry, params := m.Match(c.Request.Method, c.Request.URL.Path)
		if entry == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
			return
		}

		for k, v := range params {
			c.Set("route_param_"+k, v)
		}

		start := time.Now()
		isError := false

		for _, p := range entry.Plugins {
			if !p.Execute(c) {
				isError = c.Writer.Status() >= 400
				entry.Stats.Record(time.Since(start), isError)
				return
			}
		}

		if c.Writer.Written() {
			isError = c.Writer.Status() >= 400
			entry.Stats.Record(time.Since(start), isError)
			return
		}

		statusCapturer := &statusWriter{ResponseWriter: c.Writer, status: http.StatusOK}
		c.Writer = statusCapturer

		entry.Proxy.ServeHTTP(statusCapturer, c.Request)

		if statusCapturer.status >= 400 {
			isError = true
		}
		entry.Stats.Record(time.Since(start), isError)
	}
}

type statusWriter struct {
	gin.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	return w.ResponseWriter.Write(b)
}
