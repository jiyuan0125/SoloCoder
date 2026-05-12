package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL          *url.URL
	Healthy      bool
	ActiveConns  int64
	MaxConns     int64
	LastCheck    time.Time
	mu           sync.RWMutex
}

type BackendManager struct {
	backends map[string]*Backend
	routes   map[string]string
	mu       sync.RWMutex
}

type RequestIDGenerator struct {
	prefix string
	counter uint64
}

var (
	backendManager *BackendManager
	idGenerator    *RequestIDGenerator
	httpClient     *http.Client
)

func NewRequestIDGenerator() *RequestIDGenerator {
	buf := make([]byte, 6)
	rand.Read(buf)
	return &RequestIDGenerator{
		prefix: hex.EncodeToString(buf),
		counter: 0,
	}
}

func (g *RequestIDGenerator) Generate() string {
	count := atomic.AddUint64(&g.counter, 1)
	return fmt.Sprintf("%s-%016x-%d", g.prefix, time.Now().UnixNano(), count)
}

func NewBackendManager() *BackendManager {
	return &BackendManager{
		backends: make(map[string]*Backend),
		routes:   make(map[string]string),
	}
}

func (bm *BackendManager) RegisterBackend(name string, targetURL *url.URL, maxConns int64) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.backends[name] = &Backend{
		URL:      targetURL,
		Healthy:  true,
		MaxConns: maxConns,
		LastCheck: time.Now(),
	}
}

func (bm *BackendManager) AddRoute(prefix, backendName string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.routes[prefix] = backendName
}

func (bm *BackendManager) RemoveRoute(prefix string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	delete(bm.routes, prefix)
}

func (bm *BackendManager) GetBackendByPath(path string) (*Backend, bool) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	
	var matchedPrefix string
	for prefix := range bm.routes {
		if strings.HasPrefix(path, prefix) {
			if len(prefix) > len(matchedPrefix) {
				matchedPrefix = prefix
			}
		}
	}
	
	if matchedPrefix == "" {
		return nil, false
	}
	
	backendName := bm.routes[matchedPrefix]
	backend, exists := bm.backends[backendName]
	return backend, exists
}

func (bm *BackendManager) GetBackend(name string) (*Backend, bool) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	backend, exists := bm.backends[name]
	return backend, exists
}

func (bm *BackendManager) ListBackends() map[string]interface{} {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	
	result := make(map[string]interface{})
	backendsInfo := make(map[string]interface{})
	
	for name, backend := range bm.backends {
		backend.mu.RLock()
		backendsInfo[name] = map[string]interface{}{
			"url":         backend.URL.String(),
			"healthy":     backend.Healthy,
			"active":      atomic.LoadInt64(&backend.ActiveConns),
			"max":         backend.MaxConns,
			"last_check":  backend.LastCheck,
		}
		backend.mu.RUnlock()
	}
	
	result["backends"] = backendsInfo
	result["routes"] = bm.routes
	return result
}

func (bm *BackendManager) CheckHealth() {
	bm.mu.RLock()
	backends := make([]*Backend, 0, len(bm.backends))
	for _, backend := range bm.backends {
		backends = append(backends, backend)
	}
	bm.mu.RUnlock()
	
	for _, backend := range backends {
		checkBackendHealth(backend)
	}
}

func checkBackendHealth(backend *Backend) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "GET", backend.URL.String(), nil)
	if err != nil {
		backend.mu.Lock()
		backend.Healthy = false
		backend.LastCheck = time.Now()
		backend.mu.Unlock()
		return
	}
	
	resp, err := httpClient.Do(req)
	if err != nil {
		backend.mu.Lock()
		backend.Healthy = false
		backend.LastCheck = time.Now()
		backend.mu.Unlock()
		return
	}
	defer resp.Body.Close()
	
	backend.mu.Lock()
	backend.Healthy = resp.StatusCode >= 200 && resp.StatusCode < 500
	backend.LastCheck = time.Now()
	backend.mu.Unlock()
}

func incrementActive(backend *Backend) bool {
	active := atomic.AddInt64(&backend.ActiveConns, 1)
	if active > backend.MaxConns {
		atomic.AddInt64(&backend.ActiveConns, -1)
		return false
	}
	return true
}

func decrementActive(backend *Backend) {
	atomic.AddInt64(&backend.ActiveConns, -1)
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = idGenerator.Generate()
	}
	r.Header.Set("X-Request-ID", requestID)
	w.Header().Set("X-Request-ID", requestID)
	
	backend, exists := backendManager.GetBackendByPath(r.URL.Path)
	if !exists {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}
	
	backend.mu.RLock()
	isHealthy := backend.Healthy
	targetURL := backend.URL
	backend.mu.RUnlock()
	
	if !isHealthy {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	
	if !incrementActive(backend) {
		http.Error(w, "Service overloaded", http.StatusServiceUnavailable)
		return
	}
	defer decrementActive(backend)
	
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Header.Set("X-Request-ID", requestID)
		req.Host = targetURL.Host
	}
	
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		http.Error(rw, "Proxy error", http.StatusBadGateway)
	}
	
	proxy.ServeHTTP(w, r)
}

func healthCheckLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		backendManager.CheckHealth()
	}
}

type AddRouteRequest struct {
	Prefix      string `json:"prefix"`
	BackendName string `json:"backend_name"`
}

type RegisterBackendRequest struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	MaxConns int64  `json:"max_conns"`
}

func handleListBackends(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(backendManager.ListBackends())
}

func handleRegisterBackend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req RegisterBackendRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if req.Name == "" || req.URL == "" {
		http.Error(w, "Name and URL are required", http.StatusBadRequest)
		return
	}
	
	targetURL, err := url.Parse(req.URL)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	
	if req.MaxConns <= 0 {
		req.MaxConns = 100
	}
	
	backendManager.RegisterBackend(req.Name, targetURL, req.MaxConns)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleAddRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req AddRouteRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if req.Prefix == "" || req.BackendName == "" {
		http.Error(w, "Prefix and backend_name are required", http.StatusBadRequest)
		return
	}
	
	_, exists := backendManager.GetBackend(req.BackendName)
	if !exists {
		http.Error(w, "Backend not found", http.StatusNotFound)
		return
	}
	
	backendManager.AddRoute(req.Prefix, req.BackendName)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleRemoveRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	prefix := r.URL.Query().Get("prefix")
	if prefix == "" {
		http.Error(w, "prefix query parameter is required", http.StatusBadRequest)
		return
	}
	
	backendManager.RemoveRoute(prefix)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	backendManager.CheckHealth()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	idGenerator = NewRequestIDGenerator()
	backendManager = NewBackendManager()
	
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	httpClient = &http.Client{Transport: transport}
	
	http.HandleFunc("/admin/backends", handleListBackends)
	http.HandleFunc("/admin/backends/register", handleRegisterBackend)
	http.HandleFunc("/admin/routes/add", handleAddRoute)
	http.HandleFunc("/admin/routes/remove", handleRemoveRoute)
	http.HandleFunc("/admin/health", handleHealthCheck)
	http.HandleFunc("/", proxyHandler)
	
	go healthCheckLoop()
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	addr := ":" + port
	fmt.Printf("Proxy server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
