package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/example/apigateway/internal/config"
	"github.com/example/apigateway/internal/middleware"
	"github.com/example/apigateway/internal/proxy"
)

type NotFoundResponse struct {
	Error string `json:"error"`
}

func isManagementPath(path string) bool {
	return strings.HasPrefix(path, "/api/")
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

func GatewayHandler() http.Handler {
	mgr := config.GetConfigManager()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isManagementPath(r.URL.Path) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(NotFoundResponse{Error: "not found"})
			return
		}

		headers := headersToMap(r.Header)
		route, matched := mgr.MatchRoute(r.URL.Path, headers)

		if !matched {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(NotFoundResponse{Error: "route not found"})
			return
		}

		ctx := context.WithValue(r.Context(), "route", route)
		proxy.Forward(w, r.WithContext(ctx), route)
	})
}

func NewGateway() http.Handler {
	handler := GatewayHandler()
	return middleware.Chain(
		handler,
		middleware.AccessLogMiddleware(),
		middleware.AuthMiddleware(),
		middleware.RateLimitMiddleware(),
	)
}
