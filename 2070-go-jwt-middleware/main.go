package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"jwt-middleware/config"
	"jwt-middleware/jwt"
	"jwt-middleware/keys"
	"jwt-middleware/service"
	"jwt-middleware/store"
)

type LoginRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type RotateKeyRequest struct {
	Transition bool `json:"transition"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	Token string `json:"token"`
}

type StatsResponse struct {
	TotalTokensIssued  int64 `json:"total_tokens_issued"`
	TotalTokensRevoked int64 `json:"total_tokens_revoked"`
	TotalRefreshes     int64 `json:"total_refreshes"`
	TotalUsers         int64 `json:"total_users"`
	ActiveTokens       int64 `json:"active_tokens"`
}

var tokenService *service.TokenService

func main() {
	ks := keys.NewKeyStore()
	_, err := ks.AddHMACKey("HS256", false)
	if err != nil {
		log.Fatalf("Failed to initialize key: %v", err)
	}
	tokenService = service.NewTokenService(ks)

	mux := http.NewServeMux()
	mux.HandleFunc("/login", handleLogin)
	mux.HandleFunc("/refresh", handleRefresh)
	mux.HandleFunc("/logout", handleLogout)
	mux.HandleFunc("/validate", authMiddleware(handleValidate))
	mux.HandleFunc("/protected", authMiddleware(handleProtected))
	mux.HandleFunc("/rotate-key", handleRotateKey)
	mux.HandleFunc("/stats", handleStats)
	mux.HandleFunc("/", handleRoot)

	addr := fmt.Sprintf(":%d", config.ServerPort)
	log.Printf("Server starting on %s", addr)
	go cleanupExpired()
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func cleanupExpired() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		db, _ := store.NewStore(config.DBPath)
		if db != nil {
			db.Cleanup()
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string, details ...string) {
	response := map[string]interface{}{
		"error":   true,
		"message": message,
	}
	if len(details) > 0 {
		response["details"] = details[0]
	}
	writeJSON(w, status, response)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "JWT Authentication Server",
		"version": "1.0",
		"endpoints": strings.Join([]string{
			"POST /login - Login and get tokens",
			"POST /refresh - Refresh access token",
			"POST /logout - Logout (revoke token)",
			"GET  /validate - Validate token",
			"GET  /protected - Protected resource",
			"POST /rotate-key - Rotate signing key",
			"GET  /stats - Get statistics",
		}, "\n"),
	})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.UserID == "" {
		writeError(w, http.StatusBadRequest, "User ID is required")
		return
	}
	if req.Role == "" {
		req.Role = "user"
	}
	pair, _, _, err := tokenService.IssueTokens(req.UserID, req.Role)
	if err != nil {
		log.Printf("Login error: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to generate tokens")
		return
	}
	writeJSON(w, http.StatusOK, TokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(config.AccessTokenDuration.Seconds()),
	})
}

func handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}
	pair, err := tokenService.RefreshTokens(req.RefreshToken)
	if err != nil {
		switch e := err.(type) {
		case *jwt.ExpiredTokenError:
			writeError(w, http.StatusUnauthorized, "Refresh token expired", e.ExpiresAt.Format(time.RFC3339))
		case *jwt.InvalidSignatureError:
			writeError(w, http.StatusUnauthorized, "Invalid signature")
		case *jwt.InvalidIssuerError:
			writeError(w, http.StatusUnauthorized, "Invalid issuer")
		case *jwt.InvalidAudienceError:
			writeError(w, http.StatusUnauthorized, "Invalid audience")
		default:
			if err == service.ErrRefreshTokenReused {
				writeError(w, http.StatusUnauthorized, "Refresh token already used - all tokens revoked")
			} else if err == service.ErrTokenBlacklisted {
				writeError(w, http.StatusUnauthorized, "Token has been revoked")
			} else {
				log.Printf("Refresh error: %v", err)
				writeError(w, http.StatusUnauthorized, "Invalid refresh token")
			}
		}
		return
	}
	writeJSON(w, http.StatusOK, TokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(config.AccessTokenDuration.Seconds()),
	})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Token == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			req.Token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}
	if req.Token == "" {
		writeError(w, http.StatusBadRequest, "Token is required")
		return
	}
	if err := tokenService.RevokeToken(req.Token); err != nil {
		log.Printf("Logout error: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to revoke token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Token revoked successfully",
	})
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	claims, ok := getClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "Failed to get claims")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"valid":   true,
		"user_id": claims.UserID,
		"role":    claims.Role,
		"jti":     claims.Jti,
		"type":    claims.Type,
		"expires": claims.ExpiresAt().Format(time.RFC3339),
	})
}

func handleProtected(w http.ResponseWriter, r *http.Request) {
	claims, ok := getClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "Failed to get claims")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "This is a protected resource",
		"user": map[string]string{
			"id":   claims.UserID,
			"role": claims.Role,
		},
		"access": true,
	})
}

func handleRotateKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var req RotateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	kid, err := tokenService.RotateKey(req.Transition)
	if err != nil {
		log.Printf("Key rotation error: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to rotate key")
		return
	}
	response := map[string]interface{}{
		"success": true,
		"new_key_id": kid,
	}
	if req.Transition {
		response["transition_period_minutes"] = config.KeyTransitionPeriod.Minutes()
		response["message"] = fmt.Sprintf("Key rotated. Transition period: %.0f minutes", config.KeyTransitionPeriod.Minutes())
	} else {
		response["message"] = "Key rotated immediately"
	}
	writeJSON(w, http.StatusOK, response)
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	stats, err := tokenService.GetStats()
	if err != nil {
		log.Printf("Stats error: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to get statistics")
		return
	}
	writeJSON(w, http.StatusOK, StatsResponse{
		TotalTokensIssued:  stats["total_tokens_issued"],
		TotalTokensRevoked: stats["total_tokens_revoked"],
		TotalRefreshes:     stats["total_refreshes"],
		TotalUsers:         stats["total_users"],
		ActiveTokens:       stats["active_tokens"],
	})
}

type contextKey string

const claimsKey contextKey = "claims"

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		_, claims, err := tokenService.ValidateToken(tokenString)
		if err != nil {
			switch e := err.(type) {
			case *jwt.ExpiredTokenError:
				writeError(w, http.StatusUnauthorized, "Token expired", e.ExpiresAt.Format(time.RFC3339))
			case *jwt.InvalidSignatureError:
				writeError(w, http.StatusUnauthorized, "Invalid signature")
			case *jwt.InvalidIssuerError:
				writeError(w, http.StatusUnauthorized, "Invalid issuer")
			case *jwt.InvalidAudienceError:
				writeError(w, http.StatusUnauthorized, "Invalid audience")
			default:
				if err == service.ErrTokenBlacklisted {
					writeError(w, http.StatusUnauthorized, "Token has been revoked")
				} else {
					log.Printf("Auth error: %v", err)
					writeError(w, http.StatusUnauthorized, "Invalid token")
				}
			}
			return
		}
		if claims.Type != "access" {
			writeError(w, http.StatusUnauthorized, "Access token required")
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, claimsKey, claims)
		next(w, r.WithContext(ctx))
	}
}

func getClaimsFromContext(ctx context.Context) (jwt.Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	return claims, ok
}
