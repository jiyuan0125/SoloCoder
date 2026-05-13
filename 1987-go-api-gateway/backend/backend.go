package backend

import (
	"fmt"
	"sync"
	"time"
)

type BackendStatus int

const (
	StatusActive BackendStatus = iota
	StatusUnavailable
	StatusProbing
)

const (
	MaxFailures     = 10
	RecoveryTimeout = 30 * time.Second
)

type Backend struct {
	Host       string
	Port       int
	Status     BackendStatus
	failures   int
	lastFail   time.Time
	probing    bool
	recoveryAt time.Time
	mu         sync.RWMutex
}

func (b *Backend) Address() string {
	return fmt.Sprintf("%s:%d", b.Host, b.Port)
}

func (b *Backend) Key() string {
	return b.Address()
}

func (b *Backend) GetStatus() BackendStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Status
}

func (b *Backend) SetStatus(status BackendStatus) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Status = status
}

func (b *Backend) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.failures = 0
	b.Status = StatusActive
	b.probing = false
}

func (b *Backend) RecordFailure(is5xx bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !is5xx {
		return
	}

	b.failures++
	b.lastFail = time.Now()

	if b.failures >= MaxFailures {
		b.Status = StatusUnavailable
		b.recoveryAt = time.Now().Add(RecoveryTimeout)
	}
}

func (b *Backend) TryRecover() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.Status != StatusUnavailable {
		return true
	}

	if time.Now().After(b.recoveryAt) {
		b.Status = StatusProbing
		b.probing = true
		b.failures = MaxFailures / 2
		return true
	}

	return false
}

func (b *Backend) CanReceive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.Status == StatusActive || b.Status == StatusProbing {
		return true
	}

	return false
}

type BackendManager struct {
	backends map[string]*Backend
	mu       sync.RWMutex
}

func NewBackendManager() *BackendManager {
	return &BackendManager{
		backends: make(map[string]*Backend),
	}
}

func (bm *BackendManager) AddBackend(host string, port int) *Backend {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	key := fmt.Sprintf("%s:%d", host, port)
	if b, exists := bm.backends[key]; exists {
		return b
	}

	backend := &Backend{
		Host:     host,
		Port:     port,
		Status:   StatusActive,
		failures: 0,
	}

	bm.backends[key] = backend
	return backend
}

func (bm *BackendManager) RemoveBackend(host string, port int) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	key := fmt.Sprintf("%s:%d", host, port)
	delete(bm.backends, key)
}

func (bm *BackendManager) GetBackend(host string, port int) *Backend {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	key := fmt.Sprintf("%s:%d", host, port)
	return bm.backends[key]
}

func (bm *BackendManager) ListBackends() []*Backend {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	backends := make([]*Backend, 0, len(bm.backends))
	for _, b := range bm.backends {
		backends = append(backends, b)
	}

	return backends
}

func (bm *BackendManager) HasActiveBackend(host string, port int) bool {
	backend := bm.GetBackend(host, port)
	if backend == nil {
		return false
	}

	return backend.CanReceive()
}
