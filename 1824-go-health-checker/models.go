package main

import (
	"sync"
	"time"
)

type CheckType string

const (
	CheckTypeHTTP CheckType = "http"
	CheckTypeTCP  CheckType = "tcp"
)

type ComponentConfig struct {
	Name           string        `json:"name"`
	CheckType      CheckType     `json:"checkType"`
	Address        string        `json:"address"`
	Timeout        time.Duration `json:"timeout"`
	Interval       time.Duration `json:"interval"`
	CallbackURL    string        `json:"callbackUrl"`
}

type CheckResult struct {
	Success   bool          `json:"success"`
	Latency   time.Duration `json:"latency"`
	Timestamp time.Time     `json:"timestamp"`
	Error     string        `json:"error,omitempty"`
}

type ComponentStatus struct {
	Name               string         `json:"name"`
	Healthy            bool           `json:"healthy"`
	LastChecked        time.Time      `json:"lastChecked"`
	History            []CheckResult  `json:"history"`
	ConsecutiveSuccess int            `json:"-"`
	ConsecutiveFailure int            `json:"-"`
	Config             ComponentConfig `json:"config"`
}

type ComponentManager struct {
	mu         sync.RWMutex
	components map[string]*ComponentStatus
}

func NewComponentManager() *ComponentManager {
	return &ComponentManager{
		components: make(map[string]*ComponentStatus),
	}
}

func (cm *ComponentManager) Register(config ComponentConfig) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}
	if config.Interval == 0 {
		config.Interval = 10 * time.Second
	}
	
	if _, exists := cm.components[config.Name]; exists {
		return false
	}
	
	cm.components[config.Name] = &ComponentStatus{
		Name:               config.Name,
		Healthy:            true,
		History:            make([]CheckResult, 0, 10),
		ConsecutiveSuccess: 2,
		ConsecutiveFailure: 0,
		Config:             config,
	}
	return true
}

func (cm *ComponentManager) Unregister(name string) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	if _, exists := cm.components[name]; exists {
		delete(cm.components, name)
		return true
	}
	return false
}

func (cm *ComponentManager) Get(name string) (*ComponentStatus, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	comp, exists := cm.components[name]
	return comp, exists
}

func (cm *ComponentManager) GetAll() []*ComponentStatus {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	statuses := make([]*ComponentStatus, 0, len(cm.components))
	for _, comp := range cm.components {
		statuses = append(statuses, comp)
	}
	return statuses
}

func (cm *ComponentManager) AllHealthy() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	for _, comp := range cm.components {
		if !comp.Healthy {
			return false
		}
	}
	return true
}

func (cm *ComponentManager) UpdateResult(name string, result CheckResult) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	comp, exists := cm.components[name]
	if !exists {
		return false
	}
	
	oldHealthy := comp.Healthy
	
	if result.Success {
		comp.ConsecutiveSuccess++
		comp.ConsecutiveFailure = 0
		if comp.ConsecutiveSuccess >= 2 {
			comp.Healthy = true
		}
	} else {
		comp.ConsecutiveFailure++
		comp.ConsecutiveSuccess = 0
		if comp.ConsecutiveFailure >= 2 {
			comp.Healthy = false
		}
	}
	
	comp.LastChecked = result.Timestamp
	
	if len(comp.History) >= 10 {
		comp.History = comp.History[1:]
	}
	comp.History = append(comp.History, result)
	
	return comp.Healthy != oldHealthy
}
