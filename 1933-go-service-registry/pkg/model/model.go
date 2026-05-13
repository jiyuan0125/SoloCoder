package model

import (
	"sync"
	"time"
)

type InstanceStatus string

const (
	StatusPending   InstanceStatus = "pending"
	StatusCanary    InstanceStatus = "canary"
	StatusActive    InstanceStatus = "active"
	StatusDraining  InstanceStatus = "draining"
	StatusOffline   InstanceStatus = "offline"
)

type HealthCheckConfig struct {
	Path    string        `json:"path"`
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
}

type Instance struct {
	ID             string           `json:"id"`
	ServiceName    string           `json:"-"`
	Address        string           `json:"address"`
	Port           int              `json:"port"`
	Weight         int              `json:"weight"`
	Status         InstanceStatus   `json:"status"`
	Healthy        bool             `json:"healthy"`
	HealthCheck    HealthCheckConfig `json:"health_check"`
	LastHeartbeat  time.Time        `json:"last_heartbeat"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	OfflineAt      *time.Time       `json:"offline_at,omitempty"`

	healthSuccessCount int
	healthFailCount    int
	activeConnections  int
	mu                 sync.RWMutex
}

func NewInstance(id, serviceName, address string, port, weight int, hc HealthCheckConfig) *Instance {
	now := time.Now()
	return &Instance{
		ID:             id,
		ServiceName:    serviceName,
		Address:        address,
		Port:           port,
		Weight:         weight,
		Status:         StatusPending,
		Healthy:        false,
		HealthCheck:    hc,
		LastHeartbeat:  now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (i *Instance) HealthSuccess() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.healthSuccessCount++
	i.healthFailCount = 0
	i.Healthy = true
	i.UpdatedAt = time.Now()
}

func (i *Instance) HealthFail() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.healthFailCount++
	i.healthSuccessCount = 0
	i.Healthy = false
	i.UpdatedAt = time.Now()
}

func (i *Instance) GetHealthSuccessCount() int {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.healthSuccessCount
}

func (i *Instance) GetHealthFailCount() int {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.healthFailCount
}

func (i *Instance) UpdateHeartbeat() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.LastHeartbeat = time.Now()
	i.UpdatedAt = time.Now()
}

func (i *Instance) GetActiveConnections() int {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.activeConnections
}

func (i *Instance) IncrActiveConnections() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.activeConnections++
}

func (i *Instance) DecrActiveConnections() {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.activeConnections > 0 {
		i.activeConnections--
	}
}

func (i *Instance) SetStatus(status InstanceStatus) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.Status = status
	i.UpdatedAt = time.Now()
	if status == StatusOffline {
		now := time.Now()
		i.OfflineAt = &now
	}
}

func (i *Instance) GetStatus() InstanceStatus {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.Status
}
