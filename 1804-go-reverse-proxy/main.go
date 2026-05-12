package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type Backend struct {
	URL      *url.URL
	Proxy    *httputil.ReverseProxy
	Stats    *BackendStats
	Weight   int
	BaseWeight int
}

type BackendStats struct {
	Requests  []RequestRecord
	mu        sync.RWMutex
}

type RequestRecord struct {
	Timestamp time.Time
	Duration  time.Duration
	Success   bool
}

type Route struct {
	Prefix   string
	Backends []*Backend
	mu       sync.RWMutex
}

type Router struct {
	routes map[string]*Route
	mu     sync.RWMutex
}

type JWTError struct {
	Reason string
}

func (e *JWTError) Error() string {
	return e.Reason
}

func NewBackend(targetURL string, baseWeight int) (*Backend, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(parsedURL)
	
	backend := &Backend{
		URL:        parsedURL,
		Proxy:      proxy,
		Stats:      NewBackendStats(),
		Weight:     baseWeight,
		BaseWeight: baseWeight,
	}

	return backend, nil
}

func NewBackendStats() *BackendStats {
	return &BackendStats{
		Requests: make([]RequestRecord, 0),
	}
}

func (bs *BackendStats) Record(duration time.Duration, success bool) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	
	bs.Requests = append(bs.Requests, RequestRecord{
		Timestamp: time.Now(),
		Duration:  duration,
		Success:   success,
	})
	
	bs.cleanup()
}

func (bs *BackendStats) cleanup() {
	cutoff := time.Now().Add(-1 * time.Minute)
	filtered := bs.Requests[:0]
	for _, r := range bs.Requests {
		if r.Timestamp.After(cutoff) {
			filtered = append(filtered, r)
		}
	}
	bs.Requests = filtered
}

func (bs *BackendStats) GetStats() (total, success, failure int, avgDuration time.Duration, failureRate float64) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	
	bs.cleanup()
	
	total = len(bs.Requests)
	if total == 0 {
		return 0, 0, 0, 0, 0
	}
	
	var totalDuration time.Duration
	for _, r := range bs.Requests {
		if r.Success {
			success++
		} else {
			failure++
		}
		totalDuration += r.Duration
	}
	
	avgDuration = totalDuration / time.Duration(total)
	failureRate = float64(failure) / float64(total)
	
	return total, success, failure, avgDuration, failureRate
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]*Route),
	}
}

func (r *Router) AddRoute(prefix, targetURL string, baseWeight int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	backend, err := NewBackend(targetURL, baseWeight)
	if err != nil {
		return err
	}
	
	if route, exists := r.routes[prefix]; exists {
		route.mu.Lock()
		defer route.mu.Unlock()
		route.Backends = append(route.Backends, backend)
	} else {
		r.routes[prefix] = &Route{
			Prefix:   prefix,
			Backends: []*Backend{backend},
		}
	}
	
	return nil
}

func (r *Router) RemoveRoute(prefix string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.routes, prefix)
}

func (r *Router) Match(path string) *Route {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var bestMatch *Route
	var longestPrefix string
	
	for prefix, route := range r.routes {
		if strings.HasPrefix(path, prefix) {
			if len(prefix) > len(longestPrefix) {
				longestPrefix = prefix
				bestMatch = route
			}
		}
	}
	
	return bestMatch
}

func (r *Router) GetAllRoutes() map[string]*Route {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make(map[string]*Route)
	for k, v := range r.routes {
		result[k] = v
	}
	return result
}

func (r *Router) GetSortedPrefixes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	prefixes := make([]string, 0, len(r.routes))
	for prefix := range r.routes {
		prefixes = append(prefixes, prefix)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(prefixes)))
	return prefixes
}

func ParseJWT(authHeader string) (string, error) {
	if authHeader == "" {
		return "", &JWTError{Reason: "Missing Authorization header"}
	}
	
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", &JWTError{Reason: "Invalid Authorization header format"}
	}
	
	token := strings.TrimPrefix(authHeader, "Bearer ")
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", &JWTError{Reason: "Invalid JWT format"}
	}
	
	payload, err := decodeJWTBase64(parts[1])
	if err != nil {
		return "", &JWTError{Reason: "Invalid JWT payload encoding"}
	}
	
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", &JWTError{Reason: "Invalid JWT payload JSON"}
	}
	
	userID, ok := claims["sub"]
	if !ok {
		userID, ok = claims["user_id"]
		if !ok {
			userID, ok = claims["userId"]
			if !ok {
				return "", &JWTError{Reason: "JWT missing user identifier (sub/user_id/userId)"}
			}
		}
	}
	
	return fmt.Sprintf("%v", userID), nil
}

func decodeJWTBase64(s string) ([]byte, error) {
	if l := len(s) % 4; l != 0 {
		s += strings.Repeat("=", 4-l)
	}
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	
	return base64Decode(s)
}

func base64Decode(s string) ([]byte, error) {
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	
	result := make([]byte, 0, len(s)*3/4)
	var buffer uint32
	var bits int
	
	for i := 0; i < len(s); i++ {
		char := s[i]
		if char == '=' {
			break
		}
		
		idx := strings.IndexByte(base64Chars, char)
		if idx == -1 {
			return nil, fmt.Errorf("invalid base64 character")
		}
		
		buffer = (buffer << 6) | uint32(idx)
		bits += 6
		
		if bits >= 8 {
			bits -= 8
			result = append(result, byte(buffer>>bits))
		}
	}
	
	return result, nil
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/admin/") {
			next.ServeHTTP(w, r)
			return
		}
		
		userID, err := ParseJWT(r.Header.Get("Authorization"))
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error":  "Unauthorized",
				"reason": err.Error(),
			})
			return
		}
		
		r.Header.Set("X-User-Id", userID)
		next.ServeHTTP(w, r)
	})
}

func SelectBackend(route *Route) *Backend {
	route.mu.RLock()
	defer route.mu.RUnlock()
	
	if len(route.Backends) == 0 {
		return nil
	}
	
	for _, backend := range route.Backends {
		_, _, _, _, failureRate := backend.Stats.GetStats()
		if failureRate > 0.30 {
			backend.Weight = backend.BaseWeight / 2
			if backend.Weight < 1 {
				backend.Weight = 1
			}
		} else if failureRate < 0.10 {
			backend.Weight = backend.BaseWeight
		}
	}
	
	totalWeight := 0
	for _, backend := range route.Backends {
		totalWeight += backend.Weight
	}
	
	if totalWeight == 0 {
		return route.Backends[0]
	}
	
	randVal := int(time.Now().UnixNano()) % totalWeight
	for _, backend := range route.Backends {
		randVal -= backend.Weight
		if randVal < 0 {
			return backend
		}
	}
	
	return route.Backends[0]
}

func ProxyHandler(router *Router) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := router.Match(r.URL.Path)
		if route == nil {
			handleNotFound(w, router)
			return
		}
		
		backend := SelectBackend(route)
		if backend == nil {
			handleNotFound(w, router)
			return
		}
		
		startTime := time.Now()
		success := true
		
		defer func() {
			duration := time.Since(startTime)
			backend.Stats.Record(duration, success)
		}()
		
		originalPath := r.URL.Path
		r.URL.Path = strings.TrimPrefix(r.URL.Path, route.Prefix)
		if !strings.HasPrefix(r.URL.Path, "/") {
			r.URL.Path = "/" + r.URL.Path
		}
		
		proxy := httputil.NewSingleHostReverseProxy(backend.URL)
		
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			success = false
			log.Printf("Proxy error: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Bad Gateway",
			})
		}
		
		proxy.ServeHTTP(w, r)
		
		r.URL.Path = originalPath
	})
}

func handleNotFound(w http.ResponseWriter, router *Router) {
	prefixes := router.GetSortedPrefixes()
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>404 - Not Found</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #dc3545; }
        .routes { margin-top: 30px; }
        .route-item { padding: 10px; border-bottom: 1px solid #eee; }
        .prefix { font-weight: bold; color: #007bff; }
    </style>
</head>
<body>
    <h1>404 - Not Found</h1>
    <p>The requested path does not match any registered route.</p>
    
    <div class="routes">
        <h2>Available Routes:</h2>
`
	
	if len(prefixes) == 0 {
		html += "        <p>No routes registered yet.</p>\n"
	} else {
		for _, prefix := range prefixes {
			html += fmt.Sprintf("        <div class=\"route-item\"><span class=\"prefix\">%s</span></div>\n", prefix)
		}
	}
	
	html += `
    </div>
</body>
</html>`
	
	fmt.Fprint(w, html)
}

type AdminHandler struct {
	router *Router
}

func NewAdminHandler(router *Router) *AdminHandler {
	return &AdminHandler{router: router}
}

func (h *AdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/admin/routes" && r.Method == http.MethodGet:
		h.ListRoutes(w, r)
	case r.URL.Path == "/admin/routes" && r.Method == http.MethodPost:
		h.AddRoute(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/routes/") && r.Method == http.MethodDelete:
		h.DeleteRoute(w, r)
	case r.URL.Path == "/admin/stats" && r.Method == http.MethodGet:
		h.GetStats(w, r)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Endpoint not found"})
	}
}

type AddRouteRequest struct {
	Prefix     string `json:"prefix"`
	TargetURL  string `json:"target_url"`
	BaseWeight int    `json:"base_weight"`
}

func (h *AdminHandler) AddRoute(w http.ResponseWriter, r *http.Request) {
	var req AddRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}
	
	if req.Prefix == "" || req.TargetURL == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "prefix and target_url are required"})
		return
	}
	
	if req.BaseWeight <= 0 {
		req.BaseWeight = 1
	}
	
	if err := h.router.AddRoute(req.Prefix, req.TargetURL, req.BaseWeight); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *AdminHandler) DeleteRoute(w http.ResponseWriter, r *http.Request) {
	prefix := strings.TrimPrefix(r.URL.Path, "/admin/routes/")
	prefix, _ = url.PathUnescape(prefix)
	
	if prefix == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "prefix is required"})
		return
	}
	
	h.router.RemoveRoute(prefix)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *AdminHandler) ListRoutes(w http.ResponseWriter, r *http.Request) {
	routes := h.router.GetAllRoutes()
	result := make([]map[string]interface{}, 0)
	
	for prefix, route := range routes {
		route.mu.RLock()
		backends := make([]string, 0, len(route.Backends))
		for _, b := range route.Backends {
			backends = append(backends, b.URL.String())
		}
		route.mu.RUnlock()
		
		result = append(result, map[string]interface{}{
			"prefix":   prefix,
			"backends": backends,
		})
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	routes := h.router.GetAllRoutes()
	result := make([]map[string]interface{}, 0)
	
	for prefix, route := range routes {
		route.mu.RLock()
		backends := make([]map[string]interface{}, 0, len(route.Backends))
		
		for _, b := range route.Backends {
			total, success, failure, avgDuration, failureRate := b.Stats.GetStats()
			
			backends = append(backends, map[string]interface{}{
				"url":         b.URL.String(),
				"total":       total,
				"success":     success,
				"failure":     failure,
				"success_rate": fmt.Sprintf("%.2f%%", (1-failureRate)*100),
				"failure_rate": fmt.Sprintf("%.2f%%", failureRate*100),
				"avg_duration_ms": avgDuration.Milliseconds(),
				"current_weight": b.Weight,
				"base_weight":    b.BaseWeight,
			})
		}
		route.mu.RUnlock()
		
		result = append(result, map[string]interface{}{
			"prefix":   prefix,
			"backends": backends,
		})
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	router := NewRouter()
	adminHandler := NewAdminHandler(router)
	
	mux := http.NewServeMux()
	
	mux.Handle("/admin/", adminHandler)
	mux.Handle("/", AuthMiddleware(ProxyHandler(router)))
	
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	
	log.Printf("Reverse proxy server starting on port %s...", port)
	log.Printf("Admin endpoints:")
	log.Printf("  GET    /admin/routes    - List all routes")
	log.Printf("  POST   /admin/routes    - Add new route")
	log.Printf("  DELETE /admin/routes/{prefix} - Remove route")
	log.Printf("  GET    /admin/stats     - View statistics")
	
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
