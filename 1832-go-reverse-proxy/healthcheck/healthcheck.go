package healthcheck

import (
	"net/http"
	"net/url"
	"sync"
	"time"

	"reverse-proxy/store"
	"reverse-proxy/types"
)

type HealthChecker struct {
	store    *store.Store
	interval time.Duration
	stop     chan struct{}
	once     sync.Once
	client   *http.Client
}

func New(store *store.Store) *HealthChecker {
	return &HealthChecker{
		store:    store,
		interval: 10 * time.Second,
		stop:     make(chan struct{}),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (hc *HealthChecker) Start() {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hc.checkAll()
		case <-hc.stop:
			return
		}
	}
}

func (hc *HealthChecker) Stop() {
	hc.once.Do(func() {
		close(hc.stop)
	})
}

func (hc *HealthChecker) checkAll() {
	backends := hc.store.GetAll()
	var wg sync.WaitGroup
	for _, b := range backends {
		wg.Add(1)
		go func(backend *types.Backend) {
			defer wg.Done()
			hc.checkOne(backend)
		}(b)
	}
	wg.Wait()
}

func (hc *HealthChecker) checkOne(backend *types.Backend) {
	healthURL := backend.TargetURL.ResolveReference(&url.URL{Path: "/health"})
	req, err := http.NewRequest(http.MethodGet, healthURL.String(), nil)
	if err != nil {
		backend.UpdateStatus(types.StatusUnhealthy)
		return
	}

	resp, err := hc.client.Do(req)
	if err != nil {
		backend.UpdateStatus(types.StatusUnhealthy)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		backend.UpdateStatus(types.StatusHealthy)
	} else {
		backend.UpdateStatus(types.StatusUnhealthy)
	}
}
