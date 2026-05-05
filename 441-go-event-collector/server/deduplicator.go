package server

import (
	"fmt"
	"sync"
	"time"
)

type Deduplicator struct {
	mu         sync.RWMutex
	recentKeys map[string]time.Time
	window     time.Duration
}

func NewDeduplicator() *Deduplicator {
	return &Deduplicator{
		recentKeys: make(map[string]time.Time),
		window:     1 * time.Second,
	}
}

func (d *Deduplicator) GetKey(eventName, deviceID string) string {
	return fmt.Sprintf("%s:%s", eventName, deviceID)
}

func (d *Deduplicator) IsDuplicate(key string, eventTime time.Time) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	d.cleanup()

	if lastTime, exists := d.recentKeys[key]; exists {
		return eventTime.Sub(lastTime) < d.window
	}
	return false
}

func (d *Deduplicator) Record(key string, eventTime time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.recentKeys[key] = eventTime
	d.cleanup()
}

func (d *Deduplicator) cleanup() {
	now := time.Now()
	for key, timestamp := range d.recentKeys {
		if now.Sub(timestamp) > d.window {
			delete(d.recentKeys, key)
		}
	}
}
