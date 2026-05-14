package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/example/apigateway/internal/config"
	"github.com/example/apigateway/internal/middleware"
	"github.com/example/apigateway/internal/model"
	"github.com/example/apigateway/internal/proxy"
	"github.com/example/apigateway/internal/storage"
)

type NotFoundResponse struct {
	Error string `json:"error"`
}

var managementPrefixes = []string{
	"/api/routes",
	"/api/changes",
	"/api/api-keys",
	"/api/jwt-secret",
	"/api/jwt-token",
}

func isManagementAPIPath(path string) bool {
	for _, prefix := range managementPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func headersToMap(h http.Header) map[string]string {
	result := make(map[string]string)
	for k, v := range h {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func routeMatcher(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isManagementAPIPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		mgr := config.GetConfigManager()
		headers := headersToMap(r.Header)
		route, matched := mgr.MatchRoute(r.URL.Path, headers)

		if !matched {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(NotFoundResponse{Error: "route not found"})
			return
		}

		ctx := context.WithValue(r.Context(), "route", route)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func proxyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isManagementAPIPath(r.URL.Path) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(NotFoundResponse{Error: "not found"})
			return
		}

		route, ok := r.Context().Value("route").(*model.RouteRule)
		if !ok || route == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(NotFoundResponse{Error: "route not found"})
			return
		}

		proxy.Forward(w, r, route)
	})
}

func authMiddleware() middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isManagementAPIPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			route, ok := ctx.Value("route").(*model.RouteRule)
			if !ok || route == nil {
				next.ServeHTTP(w, r)
				return
			}

			switch route.AuthType {
			case model.AuthTypeNone:
				next.ServeHTTP(w, r)
				return

			case model.AuthTypeAPIKey:
				handleAPIKeyAuth(w, r, next)
				return

			case model.AuthTypeJWT:
				handleJWTAuth(w, r, next)
				return

			default:
				next.ServeHTTP(w, r)
			}
		})
	}
}

func rateLimitMiddleware() middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isManagementAPIPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			route, ok := ctx.Value("route").(*model.RouteRule)
			if !ok || route == nil {
				next.ServeHTTP(w, r)
				return
			}

			store := storage.GetStorage()
			window := time.Now().Truncate(time.Minute)

			if route.RateLimitIP > 0 {
				clientIP := middleware.GetClientIP(r)
				ipKey := "ip:" + clientIP
				_, allowed, err := store.GetAndIncrementRateLimit(ipKey, window, route.RateLimitIP)
				if err != nil {
					next.ServeHTTP(w, r)
					return
				}
				if !allowed {
					writeRateLimitError(w)
					return
				}
			}

			if route.RateLimitKey > 0 {
				apiKey, ok := ctx.Value("api_key").(string)
				if !ok || apiKey == "" {
					apiKey = r.Header.Get("X-API-Key")
				}
				if apiKey != "" {
					keyKey := "key:" + apiKey
					_, allowed, err := store.GetAndIncrementRateLimit(keyKey, window, route.RateLimitKey)
					if err != nil {
						next.ServeHTTP(w, r)
						return
					}
					if !allowed {
						writeRateLimitError(w)
						return
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writeRateLimitError(w http.ResponseWriter) {
	now := time.Now()
	nextWindow := now.Truncate(time.Minute).Add(time.Minute)
	retryAfter := int(nextWindow.Sub(now).Seconds()) + 1

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
	w.WriteHeader(http.StatusTooManyRequests)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":       "rate limit exceeded",
		"retry_after": retryAfter,
	})
}

func accessLogMiddleware() middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: 200}

			next.ServeHTTP(rec, r)

			duration := time.Since(start)
			clientID := middleware.GetClientID(r)
			clientIP := middleware.GetClientIP(r)

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

func buildChain(handler http.Handler, middlewares ...middleware.Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

func NewGateway(mux *http.ServeMux) http.Handler {
	handler := proxyHandler()

	handler = buildChain(
		handler,
		authMiddleware(),
		rateLimitMiddleware(),
	)

	handler = routeMatcher(handler)
	handler = accessLogMiddleware()(handler)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isManagementAPIPath(r.URL.Path) {
			mux.ServeHTTP(w, r)
			return
		}
		handler.ServeHTTP(w, r)
	})
}
