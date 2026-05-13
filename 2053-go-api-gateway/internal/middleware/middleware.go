package middleware

import (
	"net/http"
	"time"

	"github.com/example/apigateway/internal/model"
	"github.com/example/apigateway/internal/storage"
)

type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func GetClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

func GetClientID(r *http.Request) string {
	if id := r.Header.Get("X-User-ID"); id != "" {
		return id
	}
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return apiKey
	}
	return GetClientIP(r)
}

func AccessLogMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: 200}

			next.ServeHTTP(rec, r)

			duration := time.Since(start)
			clientID := GetClientID(r)
			clientIP := GetClientIP(r)

			go func() {
				storage.GetStorage().InsertAccessLog(&model.AccessLog{
					RequestPath: r.URL.Path,
					StatusCode:  rec.status,
					Duration:    duration,
					ClientID:    clientID,
					ClientIP:    clientIP,
					Timestamp:   time.Now(),
				})
			}()
		})
	}
}
