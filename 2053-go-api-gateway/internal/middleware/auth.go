package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/example/apigateway/internal/config"
	"github.com/example/apigateway/internal/model"
)

type AuthResponse struct {
	Error string `json:"error"`
}

func writeAuthError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(AuthResponse{Error: message})
}

func AuthMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			case model.AuthTypeJWT:
				handleJWTAuth(w, r, next)

			default:
				next.ServeHTTP(w, r)
			}
		})
	}
}

func handleAPIKeyAuth(w http.ResponseWriter, r *http.Request, next http.Handler) {
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		apiKey = r.URL.Query().Get("api_key")
	}

	if apiKey == "" {
		writeAuthError(w, http.StatusUnauthorized, "missing API key")
		return
	}

	mgr := config.GetConfigManager()
	key, err := mgr.GetAPIKey(apiKey)
	if err != nil {
		writeAuthError(w, http.StatusUnauthorized, "invalid API key")
		return
	}
	if key == nil {
		writeAuthError(w, http.StatusUnauthorized, "invalid API key")
		return
	}

	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		writeAuthError(w, http.StatusUnauthorized, "API key expired")
		return
	}

	ctx := context.WithValue(r.Context(), "user_id", key.UserID)
	ctx = context.WithValue(ctx, "api_key", apiKey)
	next.ServeHTTP(w, r.WithContext(ctx))
}

func handleJWTAuth(w http.ResponseWriter, r *http.Request, next http.Handler) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeAuthError(w, http.StatusUnauthorized, "missing authorization header")
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		writeAuthError(w, http.StatusUnauthorized, "invalid authorization format")
		return
	}

	tokenStr := parts[1]
	mgr := config.GetConfigManager()
	secretCfg, err := mgr.GetJWTSecret()
	if err != nil || secretCfg == nil {
		writeAuthError(w, http.StatusUnauthorized, "JWT not configured")
		return
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secretCfg.Secret), nil
	})

	if err != nil {
		switch {
		case strings.Contains(err.Error(), "token is expired"):
			writeAuthError(w, http.StatusUnauthorized, "JWT token expired")
		case strings.Contains(err.Error(), "signature is invalid"):
			writeAuthError(w, http.StatusUnauthorized, "invalid JWT signature")
		default:
			writeAuthError(w, http.StatusUnauthorized, "invalid JWT token: "+err.Error())
		}
		return
	}

	if !token.Valid {
		writeAuthError(w, http.StatusUnauthorized, "invalid JWT token")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok {
		userID, _ := claims["sub"].(string)
		if userID == "" {
			userID, _ = claims["user_id"].(string)
		}
		if userID != "" {
			ctx := context.WithValue(r.Context(), "user_id", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
	}

	next.ServeHTTP(w, r)
}

func GenerateAPIKey() string {
	return uuid.New().String()
}
