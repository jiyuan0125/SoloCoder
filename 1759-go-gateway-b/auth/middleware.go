package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"gateway/config"
)

type Middleware struct {
	cfg *config.Manager
}

func NewMiddleware(cfg *config.Manager) *Middleware {
	return &Middleware{cfg: cfg}
}

func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg := m.cfg.Get()
		route := m.matchRoute(cfg.Routes, r.URL.Path)

		if route == nil || !route.AuthRequired {
			next.ServeHTTP(w, r)
			return
		}

		if cfg.Auth.Mode == "jwt" {
			if m.validateJWT(r, cfg.Auth.JWT.Secret) {
				next.ServeHTTP(w, r)
				return
			}
		} else if cfg.Auth.Mode == "api_key" {
			if m.validateAPIKey(r, &cfg.Auth.APIKey) {
				next.ServeHTTP(w, r)
				return
			}
		}

		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func (m *Middleware) matchRoute(routes []config.Route, path string) *config.Route {
	for i := range routes {
		if strings.HasPrefix(path, routes[i].Path) {
			return &routes[i]
		}
	}
	return nil
}

func (m *Middleware) validateJWT(r *http.Request, secret string) bool {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return false
	}

	token := strings.TrimPrefix(auth, "Bearer ")
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}

	payload, err := decodeSegment(parts[1])
	if err != nil {
		return false
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return false
	}

	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return false
		}
	}

	signingInput := parts[0] + "." + parts[1]
	expectedSig := hmacSHA256(signingInput, secret)
	sig, err := decodeSegment(parts[2])
	if err != nil {
		return false
	}

	return hmac.Equal(sig, expectedSig)
}

func (m *Middleware) validateAPIKey(r *http.Request, cfg *config.APIKeyConfig) bool {
	header := cfg.Header
	if header == "" {
		header = "X-API-Key"
	}
	key := r.Header.Get(header)
	if key == "" {
		return false
	}
	for _, valid := range cfg.Keys {
		if key == valid {
			return true
		}
	}
	return false
}

func decodeSegment(seg string) ([]byte, error) {
	if l := len(seg) % 4; l != 0 {
		seg += strings.Repeat("=", 4-l)
	}
	return base64.URLEncoding.DecodeString(seg)
}

func hmacSHA256(data, secret string) []byte {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return h.Sum(nil)
}
