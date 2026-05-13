package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/example/apigateway/internal/model"
	"github.com/example/apigateway/internal/storage"
)

const rateLimitWindow = time.Minute

type RateLimitResponse struct {
	Error     string `json:"error"`
	RetryAfter int   `json:"retry_after"`
}

func writeRateLimitError(w http.ResponseWriter, retryAfter int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
	w.WriteHeader(http.StatusTooManyRequests)
	json.NewEncoder(w).Encode(RateLimitResponse{
		Error:     "rate limit exceeded",
		RetryAfter: retryAfter,
	})
}

func getCurrentWindow() time.Time {
	return time.Now().Truncate(rateLimitWindow)
}

func secondsUntilNextWindow() int {
	now := time.Now()
	nextWindow := now.Truncate(rateLimitWindow).Add(rateLimitWindow)
	return int(nextWindow.Sub(now).Seconds()) + 1
}

func RateLimitMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			route, ok := ctx.Value("route").(*model.RouteRule)
			if !ok || route == nil {
				next.ServeHTTP(w, r)
				return
			}

			store := storage.GetStorage()
			window := getCurrentWindow()

			if route.RateLimitIP > 0 {
				clientIP := GetClientIP(r)
				ipKey := "ip:" + clientIP
				_, allowed, err := store.GetAndIncrementRateLimit(ipKey, window, route.RateLimitIP)
				if err != nil {
					next.ServeHTTP(w, r)
					return
				}
				if !allowed {
					writeRateLimitError(w, secondsUntilNextWindow())
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
						writeRateLimitError(w, secondsUntilNextWindow())
						return
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func StartRateLimitCleanup(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	store := storage.GetStorage()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cutoff := time.Now().Add(-2 * rateLimitWindow)
			store.CleanupOldRateLimits(cutoff)
		}
	}
}
