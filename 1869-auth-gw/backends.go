package main

import (
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

type Backend struct {
	ID      string
	Address string
	proxy   *httputil.ReverseProxy
}

type BackendManager struct {
	mu       sync.RWMutex
	backends []*Backend
	index    uint64
}

func NewBackendManager(addresses []string) *BackendManager {
	bm := &BackendManager{}
	for _, addr := range addresses {
		bm.addBackend(addr)
	}
	return bm
}

func (bm *BackendManager) addBackend(address string) {
	u, err := url.Parse(address)
	if err != nil {
		return
	}
	bm.backends = append(bm.backends, &Backend{
		ID:      GenerateKeyID(),
		Address: address,
		proxy:   httputil.NewSingleHostReverseProxy(u),
	})
}

func (bm *BackendManager) AddBackends(addresses []string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	for _, addr := range addresses {
		bm.addBackend(addr)
	}
}

func (bm *BackendManager) GetNext() *Backend {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	if len(bm.backends) == 0 {
		return nil
	}

	idx := atomic.AddUint64(&bm.index, 1) % uint64(len(bm.backends))
	return bm.backends[idx]
}

func (bm *BackendManager) List() []*Backend {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	result := make([]*Backend, len(bm.backends))
	copy(result, bm.backends)
	return result
}
