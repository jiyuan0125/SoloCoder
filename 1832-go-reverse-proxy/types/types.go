package types

import (
	"net/url"
	"sync"
	"time"
)

type BackendStatus string

const (
	StatusHealthy   BackendStatus = "healthy"
	StatusUnhealthy BackendStatus = "unhealthy"
)

type Backend struct {
	ID             string
	Hosts          []string
	TargetURL      *url.URL
	Status         BackendStatus
	LastCheckTime  time.Time
	LastChangeTime time.Time
	mu             sync.RWMutex
}

func (b *Backend) UpdateStatus(status BackendStatus) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.Status != status {
		b.Status = status
		b.LastChangeTime = time.Now()
	}
	b.LastCheckTime = time.Now()
}

func (b *Backend) GetStatus() BackendStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Status
}

func (b *Backend) GetLastCheckTime() time.Time {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.LastCheckTime
}

func (b *Backend) GetLastChangeTime() time.Time {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.LastChangeTime
}
