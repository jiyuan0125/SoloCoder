package main

import (
	"sync"
	"time"

	"realtime-dashboard/common"
)

type MetricManager struct {
	metrics       map[string]*common.Metric
	historyData   map[string][]common.MetricDataPoint
	alertManager  *AlertManager
	delayMonitor  *DelayMonitor
	mu            sync.RWMutex
}

func NewMetricManager() *MetricManager {
	return &MetricManager{
		metrics:     make(map[string]*common.Metric),
		historyData: make(map[string][]common.MetricDataPoint),
		alertManager: NewAlertManager(),
		delayMonitor: NewDelayMonitor(),
	}
}

func (mm *MetricManager) RegisterMetric(key string, metadata common.MetricMetadata) bool {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if _, exists := mm.metrics[key]; exists {
		return false
	}

	mm.metrics[key] = &common.Metric{
		Key:          key,
		Metadata:     metadata,
		Value:        0,
		Timestamp:    time.Now(),
		IsExpired:    false,
		AlertThreshold: nil,
	}
	mm.historyData[key] = make([]common.MetricDataPoint, 0)
	return true
}

func (mm *MetricManager) UpdateMetric(key string, value float64, timestamp time.Time) bool {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	metric, exists := mm.metrics[key]
	if !exists {
		return false
	}

	startTime := time.Now()
	metric.Value = value
	metric.Timestamp = timestamp
	metric.IsExpired = false

	mm.historyData[key] = append(mm.historyData[key], common.MetricDataPoint{
		Value:     value,
		Timestamp: timestamp,
	})

	if metric.AlertThreshold != nil {
		mm.alertManager.CheckThreshold(key, value, metric.AlertThreshold)
	}

	mm.delayMonitor.RecordLatency(time.Since(startTime))

	return true
}

func (mm *MetricManager) GetMetric(key string) (*common.Metric, bool) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	metric, exists := mm.metrics[key]
	if !exists {
		return nil, false
	}

	metric.IsExpired = common.IsDataExpired(metric.Timestamp, time.Now())
	return metric, true
}

func (mm *MetricManager) GetAllMetrics() map[string]*common.Metric {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	result := make(map[string]*common.Metric)
	now := time.Now()
	for key, metric := range mm.metrics {
		metricCopy := *metric
		metricCopy.IsExpired = common.IsDataExpired(metric.Timestamp, now)
		result[key] = &metricCopy
	}
	return result
}

func (mm *MetricManager) GetLatestValues() []common.MetricLatestValue {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	now := time.Now()
	result := make([]common.MetricLatestValue, 0, len(mm.metrics))

	for key, metric := range mm.metrics {
		result = append(result, common.MetricLatestValue{
			Key:       key,
			Name:      metric.Metadata.Name,
			Value:     metric.Value,
			Unit:      metric.Metadata.Unit,
			Timestamp: metric.Timestamp,
			IsExpired: common.IsDataExpired(metric.Timestamp, now),
		})
	}

	return result
}

func (mm *MetricManager) SetAlertThreshold(key string, threshold *common.AlertThreshold) bool {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	metric, exists := mm.metrics[key]
	if !exists {
		return false
	}

	metric.AlertThreshold = threshold
	return true
}

func (mm *MetricManager) GetHistory(key string, startTime time.Time, endTime time.Time) []common.MetricDataPoint {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	data, exists := mm.historyData[key]
	if !exists {
		return nil
	}

	var result []common.MetricDataPoint
	for _, dp := range data {
		if (dp.Timestamp.Equal(startTime) || dp.Timestamp.After(startTime)) &&
			(dp.Timestamp.Equal(endTime) || dp.Timestamp.Before(endTime)) {
			result = append(result, dp)
		}
	}

	return result
}

func (mm *MetricManager) GetTrend(key string) []common.TrendDataPoint {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	data, exists := mm.historyData[key]
	if !exists {
		return nil
	}

	return common.CalculateTrendPoints(data, time.Now())
}

func (mm *MetricManager) CleanupOldData() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	now := time.Now()
	for key, data := range mm.historyData {
		var filtered []common.MetricDataPoint
		for _, dp := range data {
			if common.IsDataWithinRetention(dp.Timestamp, now) {
				filtered = append(filtered, dp)
			}
		}
		mm.historyData[key] = filtered
	}

	for _, metric := range mm.metrics {
		metric.IsExpired = common.IsDataExpired(metric.Timestamp, now)
	}
}

func (mm *MetricManager) GetComparison(key string) (*common.ComparisonData, bool) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	metric, exists := mm.metrics[key]
	if !exists {
		return nil, false
	}

	now := time.Now()
	oneDayAgo := now.Add(-24 * time.Hour)
	oneMonthAgo := now.Add(-30 * 24 * time.Hour)

	yoYValue := mm.getValueAtTime(key, oneDayAgo)
	moMValue := mm.getValueAtTime(key, oneMonthAgo)

	comparison := common.CalculateComparison(metric.Value, yoYValue, moMValue)
	return &comparison, true
}

func (mm *MetricManager) getValueAtTime(key string, targetTime time.Time) float64 {
	data, exists := mm.historyData[key]
	if !exists || len(data) == 0 {
		return 0
	}

	for i := len(data) - 1; i >= 0; i-- {
		if data[i].Timestamp.Before(targetTime) || data[i].Timestamp.Equal(targetTime) {
			return data[i].Value
		}
	}

	if len(data) > 0 {
		return data[0].Value
	}
	return 0
}

func (mm *MetricManager) GetPendingAlerts() []common.Alert {
	return mm.alertManager.GetPendingAlerts()
}

func (mm *MetricManager) GetDelayStatus() (warning bool, critical bool, avgMS float64) {
	return mm.delayMonitor.GetStatus()
}

func (mm *MetricManager) GetLatencyStats() (currentMS, avgMS, maxMS, p99MS float64) {
	return mm.delayMonitor.GetLatencyStats()
}
