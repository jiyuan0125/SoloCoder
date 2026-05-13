package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type UsageStats struct {
	TotalProcessed   int64            `json:"total_processed"`
	TotalBytes       int64            `json:"total_bytes"`
	ByType           map[string]int64 `json:"by_type"`
	BySize           map[string]int64 `json:"by_size"`
	LastUpdated      int64            `json:"last_updated"`
}

type QuotaAllocation struct {
	TotalQuota       int64            `json:"total_quota"`
	Used             int64            `json:"used"`
	Allocations      map[string]int64 `json:"allocations"`
}

type Cache struct {
	workDir string
	mu      sync.RWMutex
}

func New(workDir string) (*Cache, error) {
	if err := os.MkdirAll(filepath.Join(workDir, "cache"), 0755); err != nil {
		return nil, err
	}
	return &Cache{workDir: workDir}, nil
}

func (c *Cache) UpdateUsage(totalBytes int64, fileType string, sizeCategory string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	stats := c.loadStats()
	stats.TotalProcessed++
	stats.TotalBytes += totalBytes

	if stats.ByType == nil {
		stats.ByType = make(map[string]int64)
	}
	stats.ByType[fileType]++

	if stats.BySize == nil {
		stats.BySize = make(map[string]int64)
	}
	stats.BySize[sizeCategory]++

	stats.LastUpdated = 0

	return c.saveStats(stats)
}

func (c *Cache) AllocateQuotas(totalQuota int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	stats := c.loadStats()
	alloc := &QuotaAllocation{
		TotalQuota:  totalQuota,
		Allocations: make(map[string]int64),
	}

	if stats.TotalProcessed > 0 {
		for typ, count := range stats.ByType {
			ratio := float64(count) / float64(stats.TotalProcessed)
			alloc.Allocations[typ] = int64(float64(totalQuota) * ratio)
		}
	}

	return c.saveQuotas(alloc)
}

func (c *Cache) GetStats() (*UsageStats, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loadStats(), nil
}

func (c *Cache) GetQuotas() (*QuotaAllocation, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loadQuotas(), nil
}

func (c *Cache) loadStats() *UsageStats {
	data, err := os.ReadFile(filepath.Join(c.workDir, "cache", "stats.json"))
	if err != nil {
		return &UsageStats{
			ByType: make(map[string]int64),
			BySize: make(map[string]int64),
		}
	}
	var stats UsageStats
	json.Unmarshal(data, &stats)
	if stats.ByType == nil {
		stats.ByType = make(map[string]int64)
	}
	if stats.BySize == nil {
		stats.BySize = make(map[string]int64)
	}
	return &stats
}

func (c *Cache) saveStats(stats *UsageStats) error {
	data, _ := json.MarshalIndent(stats, "", "  ")
	return os.WriteFile(filepath.Join(c.workDir, "cache", "stats.json"), data, 0644)
}

func (c *Cache) loadQuotas() *QuotaAllocation {
	data, err := os.ReadFile(filepath.Join(c.workDir, "cache", "quotas.json"))
	if err != nil {
		return &QuotaAllocation{
			Allocations: make(map[string]int64),
		}
	}
	var quotas QuotaAllocation
	json.Unmarshal(data, &quotas)
	if quotas.Allocations == nil {
		quotas.Allocations = make(map[string]int64)
	}
	return &quotas
}

func (c *Cache) saveQuotas(quotas *QuotaAllocation) error {
	data, _ := json.MarshalIndent(quotas, "", "  ")
	return os.WriteFile(filepath.Join(c.workDir, "cache", "quotas.json"), data, 0644)
}
