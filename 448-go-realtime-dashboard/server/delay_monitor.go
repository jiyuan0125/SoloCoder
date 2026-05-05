package main

import (
	"sync"
	"time"

	"realtime-dashboard/common"
)

type DelayMonitor struct {
	latencies         []time.Duration
	warningThreshold  time.Duration
	criticalThreshold time.Duration
	maxSamples        int
	mu                sync.RWMutex
	alertChan         chan common.Alert
}

func NewDelayMonitor() *DelayMonitor {
	return &DelayMonitor{
		latencies:         make([]time.Duration, 0),
		warningThreshold:  time.Duration(common.DefaultDelayWarning) * time.Millisecond,
		criticalThreshold: time.Duration(common.DefaultDelayCritical) * time.Millisecond,
		maxSamples:        1000,
		alertChan:         make(chan common.Alert, 100),
	}
}

func (dm *DelayMonitor) SetThresholds(warningMS, criticalMS int) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.warningThreshold = time.Duration(warningMS) * time.Millisecond
	dm.criticalThreshold = time.Duration(criticalMS) * time.Millisecond
}

func (dm *DelayMonitor) RecordLatency(latency time.Duration) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.latencies = append(dm.latencies, latency)

	if len(dm.latencies) > dm.maxSamples {
		dm.latencies = dm.latencies[len(dm.latencies)-dm.maxSamples:]
	}

	if latency > dm.criticalThreshold {
		dm.sendAlert("critical", latency)
	} else if latency > dm.warningThreshold {
		dm.sendAlert("warning", latency)
	}
}

func (dm *DelayMonitor) sendAlert(level string, latency time.Duration) {
	alert := common.Alert{
		ID:          common.GenerateAlertID(),
		MetricKey:   "push_latency",
		Message:     "数据推送延迟超过阈值",
		Threshold:   float64(dm.warningThreshold.Milliseconds()),
		ActualValue: float64(latency.Milliseconds()),
		AlertType:   level,
		Timestamp:   time.Now(),
	}

	select {
	case dm.alertChan <- alert:
	default:
	}
}

func (dm *DelayMonitor) GetStatus() (warning bool, critical bool, avgMS float64) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if len(dm.latencies) == 0 {
		return false, false, 0
	}

	var total time.Duration
	for _, l := range dm.latencies {
		total += l
	}
	avg := total / time.Duration(len(dm.latencies))

	warning = avg > dm.warningThreshold
	critical = avg > dm.criticalThreshold
	avgMS = float64(avg.Microseconds()) / 1000.0

	return warning, critical, avgMS
}

func (dm *DelayMonitor) GetAlertChannel() chan common.Alert {
	return dm.alertChan
}

func (dm *DelayMonitor) GetLatencyStats() (currentMS, avgMS, maxMS, p99MS float64) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if len(dm.latencies) == 0 {
		return 0, 0, 0, 0
	}

	current := dm.latencies[len(dm.latencies)-1]
	currentMS = float64(current.Microseconds()) / 1000.0

	var total time.Duration
	var max time.Duration
	for _, l := range dm.latencies {
		total += l
		if l > max {
			max = l
		}
	}
	avg := total / time.Duration(len(dm.latencies))
	avgMS = float64(avg.Microseconds()) / 1000.0
	maxMS = float64(max.Microseconds()) / 1000.0

	p99Index := int(float64(len(dm.latencies)) * 0.99)
	if p99Index >= 0 && p99Index < len(dm.latencies) {
		sorted := make([]time.Duration, len(dm.latencies))
		copy(sorted, dm.latencies)
		for i := range sorted {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[i] > sorted[j] {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
		p99MS = float64(sorted[p99Index].Microseconds()) / 1000.0
	}

	return currentMS, avgMS, maxMS, p99MS
}
