package memory

import (
	"json-validator/config"
	"runtime"
	"sync"
)

type MemoryMonitor struct {
	mu              sync.RWMutex
	currentUsage    uint64
	maxUsage        uint64
	enabled         bool
}

var monitor *MemoryMonitor
var once sync.Once

func GetMonitor() *MemoryMonitor {
	once.Do(func() {
		monitor = &MemoryMonitor{
			maxUsage: config.MaxMemoryUsage,
			enabled:  config.MemoryCheckEnabled,
		}
	})
	return monitor
}

func (m *MemoryMonitor) Update() {
	if !m.enabled {
		return
	}
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.mu.Lock()
	m.currentUsage = memStats.Alloc
	m.mu.Unlock()
}

func (m *MemoryMonitor) IsOverLimit() bool {
	if !m.enabled {
		return false
	}
	m.Update()
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentUsage > m.maxUsage
}

func (m *MemoryMonitor) GetCurrentUsage() uint64 {
	m.Update()
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentUsage
}

func (m *MemoryMonitor) GetMaxUsage() uint64 {
	return m.maxUsage
}
