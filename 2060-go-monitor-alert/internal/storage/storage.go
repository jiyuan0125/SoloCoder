package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"monitor-alert/internal/types"
)

type Storage struct {
	dataDir    string
	mu         sync.RWMutex
	metricsIdx   map[string][]types.Metric
	stats      types.StatsInfo
	statsFile  string
	metricFile string
}

func NewStorage(dataDir string) (*Storage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}
	
	s := &Storage{
		dataDir:    dataDir,
		metricsIdx:   make(map[string][]types.Metric),
		statsFile:  filepath.Join(dataDir, "stats.json"),
		metricFile: filepath.Join(dataDir, "metrics.json"),
	}
	
	if err := s.load(); err != nil {
		log.Printf("WARN: 加载存储数据失败: %v，将使用空数据", err)
	}
	
	return s, nil
}

func (s *Storage) Save(metrics []types.Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	for _, m := range metrics {
		name := m.Name
		if _, exists := s.metricsIdx[name]; !exists {
			s.metricsIdx[name] = make([]types.Metric, 0)
		}
		s.metricsIdx[name] = append(s.metricsIdx[name], m)
	}
	
	if err := s.saveMetrics(); err != nil {
		return err
	}
	
	return s.updateStats()
}

func (s *Storage) Query(name string, start, end time.Time) ([]types.Metric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	metrics, exists := s.metricsIdx[name]
	if !exists {
		return []types.Metric{}, nil
	}
	
	result := make([]types.Metric, 0)
	for _, m := range metrics {
		if m.Timestamp.After(start) && m.Timestamp.Before(end) {
			result = append(result, m)
		}
	}
	
	return result, nil
}

func (s *Storage) ListMetrics() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	names := make([]string, 0, len(s.metricsIdx))
	for name := range s.metricsIdx {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (s *Storage) TrendAnalysis(name string, duration time.Duration) (TrendResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	metrics, exists := s.metricsIdx[name]
	if !exists {
		return TrendResult{}, fmt.Errorf("指标不存在: %s", name)
	}
	
	if len(metrics) == 0 {
		return TrendResult{}, fmt.Errorf("指标 %s 没有数据", name)
	}
	
	now := time.Now()
	var filtered []types.Metric
	cutoff := now.Add(-duration)
	for _, m := range metrics {
		if m.Timestamp.After(cutoff) {
			filtered = append(filtered, m)
		}
	}
	
	if len(filtered) == 0 {
		return TrendResult{}, fmt.Errorf("指标 %s 在指定时间段内没有数据", name)
	}
	
	var sum, min, max float64
	min = math.Inf(1)
	max = math.Inf(-1)
	for _, m := range filtered {
		sum += m.Value
		if m.Value < min {
			min = m.Value
		}
		if m.Value > max {
			max = m.Value
		}
	}
	
	avg := sum / float64(len(filtered))
	
	var trend string
	if len(filtered) >= 2 {
		first := filtered[:len(filtered)/2]
		second := filtered[len(filtered)/2:]
		var firstAvg, secondAvg float64
		for _, m := range first {
			firstAvg += m.Value
		}
		firstAvg /= float64(len(first))
		for _, m := range second {
			secondAvg += m.Value
		}
		secondAvg /= float64(len(second))
		
		change := (secondAvg - firstAvg) / firstAvg * 100
		if math.Abs(change) < 1 {
			trend = "stable"
		} else if change > 0 {
			trend = "rising"
		} else {
			trend = "falling"
		}
	} else {
		trend = "insufficient_data"
	}
	
	return TrendResult{
		MetricName: name,
		Count:    len(filtered),
		Average:  avg,
		Min:      min,
		Max:      max,
		Latest:   filtered[len(filtered)-1].Value,
		Trend:    trend,
	}, nil
}

type TrendResult struct {
	MetricName string  `json:"metric_name"`
	Count      int     `json:"count"`
	Average    float64 `json:"average"`
	Min        float64 `json:"min"`
	Max        float64 `json:"max"`
	Latest     float64 `json:"latest"`
	Trend      string  `json:"trend"`
}

func (s *Storage) Stats() types.StatsInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stats
}

func (s *Storage) AdjustTotal(newTotal int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	oldTotal := s.stats.TotalRecords
	if oldTotal == 0 {
		return fmt.Errorf("当前总数为 0，无法按比例调整")
	}
	
	ratio := float64(newTotal) / float64(oldTotal)
	metricsCount := make(map[string]int64)
	var adjustedTotal int64
	
	for name, count := range s.stats.MetricsCount {
		metricsCount[name] = int64(float64(count) * ratio)
		adjustedTotal += metricsCount[name]
	}
	
	keys := make([]string, 0, len(metricsCount))
	for k := range metricsCount {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	diff := newTotal - adjustedTotal
	if diff != 0 && len(keys) > 0 {
		for i := int64(0); i < diff; i++ {
			idx := int(i) % len(keys)
			metricsCount[keys[idx]]++
		}
		for i := int64(0); i > diff; i-- {
			idx := int(-i) % len(keys)
			if metricsCount[keys[idx]] > 0 {
				metricsCount[keys[idx]]--
			}
		}
	}
	
	var checkTotal int64
	for _, count := range metricsCount {
		checkTotal += count
	}
	if checkTotal != newTotal {
		return fmt.Errorf("调整后总数不一致: %d != %d", checkTotal, newTotal)
	}
	
	s.stats.TotalRecords = newTotal
	s.stats.MetricsCount = metricsCount
	s.stats.LastUpdated = time.Now()
	
	return s.saveStats()
}

func (s *Storage) load() error {
	if err := s.loadMetrics(); err != nil && !os.IsNotExist(err) {
		return err
	}
	return s.loadStats()
}

func (s *Storage) loadMetrics() error {
	data, err := os.ReadFile(s.metricFile)
	if err != nil {
		return err
	}
	
	var allMetrics []types.Metric
	if err := json.Unmarshal(data, &allMetrics); err != nil {
		return fmt.Errorf("解析指标文件失败: %w", err)
	}
	
	s.metricsIdx = make(map[string][]types.Metric)
	for _, m := range allMetrics {
		name := m.Name
		s.metricsIdx[name] = append(s.metricsIdx[name], m)
	}
	
	return nil
}

func (s *Storage) loadStats() error {
	data, err := os.ReadFile(s.statsFile)
	if err != nil {
		if os.IsNotExist(err) {
			s.stats = types.StatsInfo{
				MetricsCount: make(map[string]int64),
			}
			return nil
		}
		return err
	}
	
	if err := json.Unmarshal(data, &s.stats); err != nil {
		return fmt.Errorf("解析统计文件失败: %w", err)
	}
	
	if s.stats.MetricsCount == nil {
		s.stats.MetricsCount = make(map[string]int64)
	}
	
	return nil
}

func (s *Storage) saveMetrics() error {
	allMetrics := make([]types.Metric, 0)
	for _, metrics := range s.metricsIdx {
		allMetrics = append(allMetrics, metrics...)
	}
	
	sort.Slice(allMetrics, func(i, j int) bool {
		return allMetrics[i].Timestamp.Before(allMetrics[j].Timestamp)
	})
	
	data, err := json.MarshalIndent(allMetrics, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化指标失败: %w", err)
	}
	
	return os.WriteFile(s.metricFile, data, 0644)
}

func (s *Storage) updateStats() error {
	total := int64(0)
	metricsCount := make(map[string]int64)
	for name, metrics := range s.metricsIdx {
		count := int64(len(metrics))
		metricsCount[name] = count
		total += count
	}
	
	var loadedCount int64 = 0
	for _, count := range s.stats.MetricsCount {
		loadedCount += count
	}
	if loadedCount != total {
		log.Printf("WARN: 统计总数不一致: 实际 %d, 存储 %d", total, loadedCount)
	}
	
	s.stats.TotalRecords = total
	s.stats.MetricsCount = metricsCount
	s.stats.LastUpdated = time.Now()
	
	return s.saveStats()
}

func (s *Storage) saveStats() error {
	data, err := json.MarshalIndent(s.stats, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化统计信息失败: %w", err)
	}
	
	return os.WriteFile(s.statsFile, data, 0644)
}
