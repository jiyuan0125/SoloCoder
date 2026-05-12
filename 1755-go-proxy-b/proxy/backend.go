package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"proxy-service/config"
)

type Backend struct {
	Name      string
	URL       *url.URL
	Path      string
	Proxy     *httputil.ReverseProxy
	Healthy   bool
	mu        sync.RWMutex
}

type BackendManager struct {
	backends    map[string]*Backend
	pathMap     map[string]*Backend
	config      *config.Config
	mu          sync.RWMutex
	HealthCheck *HealthChecker
}

func NewBackendManager(cfg *config.Config) *BackendManager {
	bm := &BackendManager{
		backends: make(map[string]*Backend),
		pathMap:  make(map[string]*Backend),
		config:   cfg,
	}

	for _, bc := range cfg.Backends {
		parsedURL, err := url.Parse(bc.URL)
		if err != nil {
			fmt.Printf("Invalid URL for backend %s: %v\n", bc.Name, err)
			continue
		}

		proxy := httputil.NewSingleHostReverseProxy(parsedURL)
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			req.Header.Set("X-Forwarded-By", "go-proxy")
			req.Header.Set("X-Proxy-Time", time.Now().Format(time.RFC3339))
			req.Header.Set("X-Forwarded-Proto", "http")
		}

		backend := &Backend{
			Name:    bc.Name,
			URL:     parsedURL,
			Path:    bc.Path,
			Proxy:   proxy,
			Healthy: true,
		}

		bm.backends[bc.Name] = backend
		bm.pathMap[bc.Path] = backend
	}

	bm.HealthCheck = NewHealthChecker(bm, cfg)
	return bm
}

func (bm *BackendManager) MatchBackend(path string) *Backend {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	for prefix, backend := range bm.pathMap {
		if strings.HasPrefix(path, prefix) {
			backend.mu.RLock()
			healthy := backend.Healthy
			backend.mu.RUnlock()
			if healthy {
				return backend
			}
		}
	}
	return nil
}

func (bm *BackendManager) GetAllBackends() map[string]*Backend {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	copy := make(map[string]*Backend)
	for k, v := range bm.backends {
		copy[k] = v
	}
	return copy
}

func (bm *BackendManager) SetBackendHealth(name string, healthy bool) {
	bm.mu.RLock()
	backend, exists := bm.backends[name]
	bm.mu.RUnlock()

	if exists {
		backend.mu.Lock()
		backend.Healthy = healthy
		backend.mu.Unlock()
	}
}
