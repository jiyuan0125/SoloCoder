package models

import (
	"sync"
	"time"
)

type PluginType string

const (
	PluginAuth    PluginType = "auth"
	PluginRatelim PluginType = "ratelim"
	PluginLog     PluginType = "log"
	PluginRewrite PluginType = "rewrite"
)

type PluginConfig struct {
	Type   PluginType         `json:"type"`
	Name   string             `json:"name"`
	Enable bool               `json:"enable"`
	Config map[string]any     `json:"config"`
}

type Route struct {
	ID           string         `json:"id"`
	Path         string         `json:"path"`
	Methods      []string       `json:"methods"`
	Upstream     string         `json:"upstream"`
	Plugins      []PluginConfig `json:"plugins"`
	Enabled      bool           `json:"enabled"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	Version      int            `json:"version"`
}

type RouteVersion struct {
	Route
	VersionTime time.Time `json:"version_time"`
}

type RouteStats struct {
	sync.RWMutex
	TotalRequests uint64        `json:"total_requests"`
	TotalErrors   uint64        `json:"total_errors"`
	TotalLatency  time.Duration `json:"total_latency"`
	StartTime     time.Time     `json:"start_time"`
}

func (s *RouteStats) Record(latency time.Duration, isError bool) {
	s.Lock()
	s.TotalRequests++
	s.TotalLatency += latency
	if isError {
		s.TotalErrors++
	}
	s.Unlock()
}

func (s *RouteStats) AvgLatency() time.Duration {
	s.RLock()
	defer s.RUnlock()
	if s.TotalRequests == 0 {
		return 0
	}
	return s.TotalLatency / time.Duration(s.TotalRequests)
}

func (s *RouteStats) QPS() float64 {
	s.RLock()
	defer s.RUnlock()
	elapsed := time.Since(s.StartTime).Seconds()
	if elapsed <= 0 {
		return 0
	}
	return float64(s.TotalRequests) / elapsed
}

func (s *RouteStats) ErrorRate() float64 {
	s.RLock()
	defer s.RUnlock()
	if s.TotalRequests == 0 {
		return 0
	}
	return float64(s.TotalErrors) / float64(s.TotalRequests)
}
