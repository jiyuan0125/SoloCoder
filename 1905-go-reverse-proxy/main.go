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
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Backend struct {
	ID      string `json:"id"`
	Host    string `json:"host"`
	Path    string `json:"path,omitempty"`
	Mode    string `json:"mode,omitempty"`
	Address string `json:"address"`
}

type BackendGroup struct {
	Host      string
	Path      string
	Mode      string
	Backends  []Backend
	counter   uint64
	mu        sync.RWMutex
}

type Router struct {
	hostGroups    map[string]*BackendGroup
	exactRoutes   map[string]*BackendGroup
	prefixRoutes  map[string]*BackendGroup
	allBackends   map[string]*Backend
	mu            sync.RWMutex
	nextID        int64
}

var router = &Router{
	hostGroups:   make(map[string]*BackendGroup),
	exactRoutes:  make(map[string]*BackendGroup),
	prefixRoutes: make(map[string]*BackendGroup),
	allBackends:  make(map[string]*Backend),
}

var addressRegex = regexp.MustCompile(`^([a-zA-Z0-9.-]+|\[[a-fA-F0-9:]+\]):(\d+)$`)

func validateAddress(address string) error {
	matches := addressRegex.FindStringSubmatch(address)
	if matches == nil {
		return fmt.Errorf("invalid address format, expected host:port")
	}
	port, err := strconv.Atoi(matches[2])
	if err != nil {
		return fmt.Errorf("invalid port number")
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	return nil
}

func (r *Router) generateID() string {
	id := atomic.AddInt64(&r.nextID, 1)
	return strconv.FormatInt(id, 10)
}

func (r *Router) AddBackend(host, path, mode, address string) (*Backend, error) {
	if err := validateAddress(address); err != nil {
		return nil, err
	}

	if path != "" && mode == "" {
		mode = "prefix"
	}

	if mode != "" && mode != "exact" && mode != "prefix" {
		return nil, fmt.Errorf("invalid mode, must be 'exact' or 'prefix'")
	}

	if path != "" && mode == "" {
		mode = "prefix"
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	backend := &Backend{
		ID:      r.generateID(),
		Host:    host,
		Path:    path,
		Mode:    mode,
		Address: address,
	}

	var group *BackendGroup

	if host != "" && path == "" {
		group = r.hostGroups[host]
		if group == nil {
			group = &BackendGroup{Host: host}
			r.hostGroups[host] = group
		}
	} else if path != "" {
		if mode == "exact" {
			group = r.exactRoutes[path]
			if group == nil {
				group = &BackendGroup{Path: path, Mode: "exact"}
				r.exactRoutes[path] = group
			}
		} else {
			group = r.prefixRoutes[path]
			if group == nil {
				group = &BackendGroup{Path: path, Mode: "prefix"}
				r.prefixRoutes[path] = group
			}
		}
	} else {
		return nil, fmt.Errorf("host or path is required")
	}

	group.mu.Lock()
	group.Backends = append(group.Backends, *backend)
	group.mu.Unlock()

	r.allBackends[backend.ID] = backend

	return backend, nil
}

func (r *Router) RemoveBackend(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	backend, ok := r.allBackends[id]
	if !ok {
		return false
	}

	var group *BackendGroup

	if backend.Path != "" {
		if backend.Mode == "exact" {
			group = r.exactRoutes[backend.Path]
		} else {
			group = r.prefixRoutes[backend.Path]
		}
	} else {
		group = r.hostGroups[backend.Host]
	}

	if group == nil {
		delete(r.allBackends, id)
		return true
	}

	group.mu.Lock()
	newBackends := make([]Backend, 0, len(group.Backends))
	for _, b := range group.Backends {
		if b.ID != id {
			newBackends = append(newBackends, b)
		}
	}
	group.Backends = newBackends
	group.mu.Unlock()

	if len(group.Backends) == 0 {
		if backend.Path != "" {
			if backend.Mode == "exact" {
				delete(r.exactRoutes, backend.Path)
			} else {
				delete(r.prefixRoutes, backend.Path)
			}
		} else {
			delete(r.hostGroups, backend.Host)
		}
	}

	delete(r.allBackends, id)
	return true
}

func (r *Router) GetAllBackends() []Backend {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []Backend
	for _, b := range r.allBackends {
		all = append(all, *b)
	}
	return all
}

func (r *Router) GetGroupByHost(host string) *BackendGroup {
	r.mu.RLock()
	defer r.mu.RUnlock()

	host = strings.ToLower(strings.Split(host, ":")[0])
	if group, ok := r.hostGroups[host]; ok {
		return group
	}
	return nil
}

func (r *Router) GetGroupByPath(path string) *BackendGroup {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if group, ok := r.exactRoutes[path]; ok {
		return group
	}

	var bestPrefix string
	for prefix := range r.prefixRoutes {
		if strings.HasPrefix(path, prefix) && len(prefix) > len(bestPrefix) {
			bestPrefix = prefix
		}
	}

	if bestPrefix != "" {
		return r.prefixRoutes[bestPrefix]
	}
	return nil
}

func (g *BackendGroup) GetNextBackend() (Backend, uint64) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if len(g.Backends) == 0 {
		return Backend{}, 0
	}

	count := atomic.AddUint64(&g.counter, 1) - 1
	idx := int(count % uint64(len(g.Backends)))
	return g.Backends[idx], count
}

func handleManageBackends(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		backends := router.GetAllBackends()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(backends)

	case http.MethodPost:
		var req struct {
			Host    string `json:"host"`
			Path    string `json:"path"`
			Mode    string `json:"mode"`
			Address string `json:"address"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if req.Address == "" {
			http.Error(w, "address is required", http.StatusBadRequest)
			return
		}

		if req.Host == "" && req.Path == "" {
			http.Error(w, "host or path is required", http.StatusBadRequest)
			return
		}

		backend, err := router.AddBackend(req.Host, req.Path, req.Mode, req.Address)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(backend)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleDeleteBackend(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/backends/")
	if id == "" {
		http.Error(w, "Backend ID is required", http.StatusBadRequest)
		return
	}

	if router.RemoveBackend(id) {
		w.WriteHeader(http.StatusNoContent)
	} else {
		http.Error(w, "Backend not found", http.StatusNotFound)
	}
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func hopHeader(h string) bool {
	switch strings.ToLower(h) {
	case "connection", "keep-alive", "proxy-authenticate", "proxy-authorization",
		"te", "trailers", "transfer-encoding", "upgrade":
		return true
	}
	return false
}

func isWebSocketUpgrade(req *http.Request) bool {
	conn := strings.ToLower(req.Header.Get("Connection"))
	upgrade := strings.ToLower(req.Header.Get("Upgrade"))
	return strings.Contains(conn, "upgrade") && upgrade == "websocket"
}

func forwardHTTP(w http.ResponseWriter, r *http.Request, group *BackendGroup) {
	var tried []string

	group.mu.RLock()
	backends := make([]Backend, len(group.Backends))
	copy(backends, group.Backends)
	group.mu.RUnlock()

	if len(backends) == 0 {
		http.Error(w, "No backends available", http.StatusBadGateway)
		return
	}

	_, startIdx := group.GetNextBackend()

	for i := 0; i < len(backends); i++ {
		idx := int((startIdx + uint64(i)) % uint64(len(backends)))
		backend := backends[idx]
		tried = append(tried, backend.Address)

		targetURL, err := url.Parse("http://" + backend.Address)
		if err != nil {
			continue
		}

		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			req.Host = r.Host
		}

		proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
			log.Printf("Backend error: %v", err)
		}

		rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		proxy.ServeHTTP(rw, r)

		if rw.status != http.StatusBadGateway {
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusBadGateway)
	var identifier string
	if group.Host != "" {
		identifier = "Host: " + group.Host
	} else if group.Path != "" {
		identifier = "Path: " + group.Path + " (" + group.Mode + ")"
	} else {
		identifier = "Unknown"
	}
	fmt.Fprintf(w, "502 Bad Gateway\n%s\nTried backends: %v", identifier, tried)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func forwardWebSocket(w http.ResponseWriter, r *http.Request, group *BackendGroup) {
	var tried []string

	group.mu.RLock()
	backends := make([]Backend, len(group.Backends))
	copy(backends, group.Backends)
	group.mu.RUnlock()

	if len(backends) == 0 {
		http.Error(w, "No backends available", http.StatusBadGateway)
		return
	}

	_, startIdx := group.GetNextBackend()

	for i := 0; i < len(backends); i++ {
		idx := int((startIdx + uint64(i)) % uint64(len(backends)))
		backend := backends[idx]
		tried = append(tried, backend.Address)

		dialer := &net.Dialer{Timeout: 10 * time.Second}
		conn, err := dialer.Dial("tcp", backend.Address)
		if err != nil {
			log.Printf("Failed to connect to backend %s: %v", backend.Address, err)
			continue
		}

		hijacker, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
			conn.Close()
			return
		}

		clientConn, _, err := hijacker.Hijack()
		if err != nil {
			http.Error(w, "Failed to hijack connection", http.StatusInternalServerError)
			conn.Close()
			return
		}

		outReq := r.Clone(r.Context())
		outReq.URL.Host = backend.Address
		outReq.URL.Scheme = "http"

		if err := outReq.Write(conn); err != nil {
			log.Printf("Failed to write request to backend: %v", err)
			conn.Close()
			clientConn.Close()
			continue
		}

		go func() {
			defer conn.Close()
			defer clientConn.Close()
			io.Copy(conn, clientConn)
		}()

		go func() {
			defer conn.Close()
			defer clientConn.Close()
			io.Copy(clientConn, conn)
		}()

		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusBadGateway)
	var identifier string
	if group.Host != "" {
		identifier = "Host: " + group.Host
	} else if group.Path != "" {
		identifier = "Path: " + group.Path + " (" + group.Mode + ")"
	} else {
		identifier = "Unknown"
	}
	fmt.Fprintf(w, "502 Bad Gateway\n%s\nTried backends: %v", identifier, tried)
}

func handleProxy(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	group := router.GetGroupByHost(host)

	if group == nil {
		group = router.GetGroupByPath(r.URL.Path)
	}

	if group == nil {
		http.Error(w, "No matching route found", http.StatusNotFound)
		return
	}

	group.mu.RLock()
	count := len(group.Backends)
	group.mu.RUnlock()

	if count == 0 {
		http.Error(w, "No backends available for this route", http.StatusBadGateway)
		return
	}

	if isWebSocketUpgrade(r) {
		forwardWebSocket(w, r, group)
	} else {
		forwardHTTP(w, r, group)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8400"
	}

	http.HandleFunc("/backends", handleManageBackends)
	http.HandleFunc("/backends/", handleDeleteBackend)
	http.HandleFunc("/", handleProxy)

	log.Printf("Reverse proxy server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
