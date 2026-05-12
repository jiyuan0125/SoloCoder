package proxy

import (
	"sync"
	"time"
)

type BackendStats struct {
	RequestCount   int64
	TotalDuration  int64
	mu             sync.Mutex
}

func NewBackendStats() *BackendStats {
	return &BackendStats{}
}

func (bs *BackendStats) RecordRequest(duration time.Duration) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.RequestCount++
	bs.TotalDuration += duration.Milliseconds()
}

func (bs *BackendStats) GetStats() (count int64, avgMs float64) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	count = bs.RequestCount
	if count > 0 {
		avgMs = float64(bs.TotalDuration) / float64(count)
	}
	return count, avgMs
}

type StatsCollector struct {
	stats map[string]*BackendStats
	mu    sync.RWMutex
}

func NewStatsCollector() *StatsCollector {
	return &StatsCollector{
		stats: make(map[string]*BackendStats),
	}
}

func (sc *StatsCollector) getOrCreateStats(name string) *BackendStats {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if _, exists := sc.stats[name]; !exists {
		sc.stats[name] = NewBackendStats()
	}
	return sc.stats[name]
}

func (sc *StatsCollector) RecordRequest(backendName string, duration time.Duration) {
	stats := sc.getOrCreateStats(backendName)
	stats.RecordRequest(duration)
}

func (sc *StatsCollector) GetStats(backendName string) (int64, float64) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	if stats, exists := sc.stats[backendName]; exists {
		return stats.GetStats()
	}
	return 0, 0
}

func (sc *StatsCollector) GetAllStats() map[string]struct {
	RequestCount int64
	AvgDuration  float64
} {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	result := make(map[string]struct {
		RequestCount int64
		AvgDuration  float64
	})
	for name, stats := range sc.stats {
		count, avg := stats.GetStats()
		result[name] = struct {
			RequestCount int64
			AvgDuration  float64
		}{
			RequestCount: count,
			AvgDuration:  avg,
		}
	}
	return result
}
