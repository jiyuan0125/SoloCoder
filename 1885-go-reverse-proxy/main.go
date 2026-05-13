package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Backend struct {
	Prefix string `json:"prefix"`
	URL    string `json:"url"`
}

type BackendManager struct {
	mu       sync.RWMutex
	backends map[string]string
}

func NewBackendManager() *BackendManager {
	return &BackendManager{
		backends: make(map[string]string),
	}
}

func (bm *BackendManager) Add(prefix, urlStr string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.backends[prefix] = urlStr
}

func (bm *BackendManager) Remove(prefix string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	delete(bm.backends, prefix)
}

func (bm *BackendManager) List() []Backend {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	list := make([]Backend, 0, len(bm.backends))
	for p, u := range bm.backends {
		list = append(list, Backend{Prefix: p, URL: u})
	}
	sort.Slice(list, func(i, j int) bool {
		return len(list[i].Prefix) > len(list[j].Prefix)
	})
	return list
}

func (bm *BackendManager) Match(path string) (string, string, bool) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	prefixes := make([]string, 0, len(bm.backends))
	for p := range bm.backends {
		prefixes = append(prefixes, p)
	}
	sort.Slice(prefixes, func(i, j int) bool {
		return len(prefixes[i]) > len(prefixes[j])
	})

	for _, p := range prefixes {
		if strings.HasPrefix(path, p) {
			return p, bm.backends[p], true
		}
	}
	return "", "", false
}

func isValidURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if u.Host == "" {
		return false
	}
	return true
}

func extractClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}
	return r.RemoteAddr
}

type ProxyHandler struct {
	bm *BackendManager
}

func NewProxyHandler(bm *BackendManager) *ProxyHandler {
	return &ProxyHandler{bm: bm}
}

func (ph *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := uuid.New().String()
	r.Header.Set("X-Request-ID", requestID)

	clientIP := extractClientIP(r)
	if existing := r.Header.Get("X-Forwarded-For"); existing != "" {
		r.Header.Set("X-Forwarded-For", existing+", "+clientIP)
	} else {
		r.Header.Set("X-Forwarded-For", clientIP)
	}

	prefix, backendURL, ok := ph.bm.Match(r.URL.Path)
	if !ok {
		log.Printf("[404] %s %s - No backend registered for path", r.Method, r.URL.Path)
		http.NotFound(w, r)
		return
	}

	target, err := url.Parse(backendURL)
	if err != nil {
		log.Printf("[ERROR] Invalid backend URL %s: %v", backendURL, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = singleJoiningSlash(target.Path, strings.TrimPrefix(r.URL.Path, prefix))
		if r.URL.RawQuery != "" {
			req.URL.RawQuery = r.URL.RawQuery
		}
		req.Host = target.Host
		req.Header = r.Header.Clone()
	}

	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, e error) {
		duration := time.Since(start)
		log.Printf("[502] %s %s -> %s (%v) - Error: %v",
			req.Method, req.URL.Path, backendURL, duration, e)
		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		rw.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(rw, "Bad Gateway: backend %s is unavailable", backendURL)
	}

	rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
	proxy.ServeHTTP(rec, r)

	duration := time.Since(start)
	log.Printf("[%d] %s %s -> %s (%v)", rec.status, r.Method, r.URL.Path, backendURL, duration)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		if a == "" {
			return b
		}
		return a + "/" + b
	}
	return a + b
}

type AdminHandler struct {
	bm *BackendManager
}

func NewAdminHandler(bm *BackendManager) *AdminHandler {
	return &AdminHandler{bm: bm}
}

func (ah *AdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ah.listBackends(w, r)
	case http.MethodPost:
		ah.addBackend(w, r)
	case http.MethodDelete:
		ah.removeBackend(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (ah *AdminHandler) listBackends(w http.ResponseWriter, r *http.Request) {
	backends := ah.bm.List()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(backends)
}

func (ah *AdminHandler) addBackend(w http.ResponseWriter, r *http.Request) {
	var b Backend
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &b); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if b.Prefix == "" {
		http.Error(w, "Prefix is required", http.StatusBadRequest)
		return
	}
	if !strings.HasPrefix(b.Prefix, "/") {
		http.Error(w, "Prefix must start with /", http.StatusBadRequest)
		return
	}
	if !isValidURL(b.URL) {
		http.Error(w, "Invalid URL: must start with http:// or https:// and have a host", http.StatusBadRequest)
		return
	}

	ah.bm.Add(b.Prefix, b.URL)
	log.Printf("[ADMIN] Registered backend: %s -> %s", b.Prefix, b.URL)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(b)
}

func (ah *AdminHandler) removeBackend(w http.ResponseWriter, r *http.Request) {
	var b Backend
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(body) > 0 {
		if err := json.Unmarshal(body, &b); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
	}

	if b.Prefix == "" {
		q := r.URL.Query()
		b.Prefix = q.Get("prefix")
	}

	if b.Prefix == "" {
		http.Error(w, "Prefix is required", http.StatusBadRequest)
		return
	}

	ah.bm.Remove(b.Prefix)
	log.Printf("[ADMIN] Removed backend: %s", b.Prefix)
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	bm := NewBackendManager()

	proxyHandler := NewProxyHandler(bm)
	adminHandler := NewAdminHandler(bm)

	mux := http.NewServeMux()
	mux.Handle("/admin/backends", adminHandler)
	mux.Handle("/", proxyHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8805"
	}

	addr := ":" + port
	log.Printf("Proxy server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
