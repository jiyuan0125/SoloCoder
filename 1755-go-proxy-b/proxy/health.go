package proxy

import (
	"log"
	"net/http"
	"time"

	"proxy-service/config"
)

type HealthChecker struct {
	manager *BackendManager
	config  *config.Config
}

func NewHealthChecker(manager *BackendManager, cfg *config.Config) *HealthChecker {
	return &HealthChecker{
		manager: manager,
		config:  cfg,
	}
}

func (hc *HealthChecker) Start() {
	ticker := time.NewTicker(time.Duration(hc.config.HealthCheck.Interval) * time.Second)
	go func() {
		for range ticker.C {
			hc.checkAllBackends()
		}
	}()
}

func (hc *HealthChecker) checkAllBackends() {
	backends := hc.manager.GetAllBackends()
	for _, backend := range backends {
		go hc.checkBackend(backend)
	}
}

func (hc *HealthChecker) checkBackend(backend *Backend) {
	client := &http.Client{
		Timeout: time.Duration(hc.config.HealthCheck.Timeout) * time.Second,
	}

	healthURL := backend.URL.String() + hc.config.HealthCheck.Path
	resp, err := client.Get(healthURL)
	if err != nil {
		log.Printf("Backend %s is unhealthy: %v", backend.Name, err)
		hc.manager.SetBackendHealth(backend.Name, false)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		backend.mu.RLock()
		currentHealth := backend.Healthy
		backend.mu.RUnlock()
		
		if !currentHealth {
			log.Printf("Backend %s has recovered", backend.Name)
		}
		hc.manager.SetBackendHealth(backend.Name, true)
	} else {
		log.Printf("Backend %s returned non-2xx status: %d", backend.Name, resp.StatusCode)
		hc.manager.SetBackendHealth(backend.Name, false)
	}
}
