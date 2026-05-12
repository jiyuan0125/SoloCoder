package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type ctxKey string

const claimsCtxKey ctxKey = "claims"

type Gateway struct {
	keys       *KeyManager
	routes     *RouteManager
	backends   *BackendManager
	audit      *AuditLogManager
	adminToken string
}

func NewGateway() *Gateway {
	rotateSeconds := int64(3600)
	if v := os.Getenv("KEY_ROTATE_SECONDS"); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil && i > 0 {
			rotateSeconds = i
		}
	}

	gw := &Gateway{
		keys:     NewKeyManager(rotateSeconds),
		routes:   NewRouteManager(),
		backends: NewBackendManager(nil),
		audit:    NewAuditLogManager(),
	}

	adminKey := gw.keys.GetActiveKey()
	adminToken, _ := SignJWT(
		JWTHeader{Kid: adminKey.ID},
		JWTClaims{
			Sub:  "admin-gateway",
			Role: "admin",
			Iat:  NowUnix(),
			Exp:  NowUnix() + 86400*365,
		},
		adminKey.Secret,
	)
	gw.adminToken = adminToken

	return gw
}

func (gw *Gateway) extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return auth
}

func (gw *Gateway) authenticate(r *http.Request) (*JWTClaims, AuditResult) {
	token := gw.extractToken(r)
	if token == "" {
		return nil, ResultUnauthorized
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ResultUnauthorized
	}

	tokenData, signingInput, err := ParseJWT(token)
	if err != nil {
		return nil, ResultUnauthorized
	}

	if tokenData.Header.Alg != "HS256" {
		return nil, ResultUnauthorized
	}

	key := gw.keys.GetKey(tokenData.Header.Kid)
	if key == nil {
		return nil, ResultUnauthorized
	}

	if !VerifySignature(signingInput, parts[2], key.Secret) {
		return nil, ResultUnauthorized
	}

	now := NowUnix()
	if tokenData.Claims.Exp > 0 && tokenData.Claims.Exp < now {
		return nil, ResultUnauthorized
	}

	return tokenData.Claims, ResultSuccess
}

func (gw *Gateway) authorize(claims *JWTClaims, path string) AuditResult {
	if strings.HasPrefix(path, "/admin/") {
		if claims.Role != "admin" {
			return ResultForbidden
		}
	}
	return ResultSuccess
}

func (gw *Gateway) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, result := gw.authenticate(r)
		clientID := ""
		if claims != nil {
			clientID = claims.Sub
		}

		if result != ResultSuccess {
			gw.audit.Log(clientID, r.URL.Path, result)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}

		result = gw.authorize(claims, r.URL.Path)
		if result != ResultSuccess {
			gw.audit.Log(clientID, r.URL.Path, result)
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "forbidden"})
			return
		}

		gw.audit.Log(clientID, r.URL.Path, ResultSuccess)

		ctx := context.WithValue(r.Context(), claimsCtxKey, claims)
		next(w, r.WithContext(ctx))
	}
}

func getClaims(r *http.Request) *JWTClaims {
	if v := r.Context().Value(claimsCtxKey); v != nil {
		return v.(*JWTClaims)
	}
	return nil
}

type AddRouteRequest struct {
	Path     string   `json:"path"`
	Backends []string `json:"backends"`
}

func (gw *Gateway) handleAddRoute(w http.ResponseWriter, r *http.Request) {
	var req AddRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	if req.Path == "" || len(req.Backends) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "path and backends required"})
		return
	}

	route := gw.routes.Add(req.Path, req.Backends)
	gw.backends.AddBackends(req.Backends)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(route)
}

func (gw *Gateway) handleDeleteRoute(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/admin/routes/")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "route id required"})
		return
	}

	if gw.routes.Delete(id) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
	} else {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "route not found"})
	}
}

func (gw *Gateway) handleListRoutes(w http.ResponseWriter, r *http.Request) {
	routes := gw.routes.List()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(routes)
}

func (gw *Gateway) handleRotateKey(w http.ResponseWriter, r *http.Request) {
	newKey, err := gw.keys.Rotate()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"key_id": newKey.ID, "status": "rotated"})
}

func (gw *Gateway) handleListAudit(w http.ResponseWriter, r *http.Request) {
	logs := gw.audit.List()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(logs)
}

func (gw *Gateway) handleGetAdminToken(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": gw.adminToken})
}

func (gw *Gateway) handleProxy(w http.ResponseWriter, r *http.Request) {
	route := gw.routes.Match(r.URL.Path)
	if route == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "no route found"})
		return
	}

	backend := gw.backends.GetNext()
	if backend == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "no backend available"})
		return
	}

	backend.proxy.ServeHTTP(w, r)
}

func (gw *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "POST" && r.URL.Path == "/admin/routes":
		gw.authMiddleware(gw.handleAddRoute)(w, r)
	case r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, "/admin/routes/"):
		gw.authMiddleware(gw.handleDeleteRoute)(w, r)
	case r.Method == "GET" && r.URL.Path == "/admin/routes":
		gw.authMiddleware(gw.handleListRoutes)(w, r)
	case r.Method == "POST" && r.URL.Path == "/admin/keys/rotate":
		gw.authMiddleware(gw.handleRotateKey)(w, r)
	case r.Method == "GET" && r.URL.Path == "/admin/audit":
		gw.authMiddleware(gw.handleListAudit)(w, r)
	case r.Method == "GET" && r.URL.Path == "/admin/token":
		http.HandlerFunc(gw.handleGetAdminToken)(w, r)
	default:
		gw.authMiddleware(gw.handleProxy)(w, r)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	gw := NewGateway()

	server := &http.Server{
		Addr:    ":" + port,
		Handler: gw,
	}

	fmt.Printf("Auth Gateway listening on port %s\n", port)
	fmt.Printf("Admin token endpoint: GET /admin/token\n")

	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
