package main

import (
	"sync"
	"time"

	"realtime-dashboard/common"
)

type AlertManager struct {
	pendingAlerts []common.Alert
	subscribers   map[string][]chan common.Alert
	mu            sync.RWMutex
}

func NewAlertManager() *AlertManager {
	return &AlertManager{
		pendingAlerts: make([]common.Alert, 0),
		subscribers:   make(map[string][]chan common.Alert),
	}
}

func (am *AlertManager) CheckThreshold(metricKey string, value float64, threshold *common.AlertThreshold) {
	am.mu.Lock()
	defer am.mu.Unlock()

	var alert *common.Alert

	if threshold.MaxValue != nil && value > *threshold.MaxValue {
		alert = &common.Alert{
			ID:          common.GenerateAlertID(),
			MetricKey:   metricKey,
			Message:     "指标值超过上限阈值",
			Threshold:   *threshold.MaxValue,
			ActualValue: value,
			AlertType:   "max_threshold_exceeded",
			Timestamp:   time.Now(),
		}
	} else if threshold.MinValue != nil && value < *threshold.MinValue {
		alert = &common.Alert{
			ID:          common.GenerateAlertID(),
			MetricKey:   metricKey,
			Message:     "指标值低于下限阈值",
			Threshold:   *threshold.MinValue,
			ActualValue: value,
			AlertType:   "min_threshold_below",
			Timestamp:   time.Now(),
		}
	}

	if alert != nil {
		am.pendingAlerts = append(am.pendingAlerts, *alert)
		am.notifySubscribers(metricKey, *alert)
	}
}

func (am *AlertManager) Subscribe(metricKey string, alertChan chan common.Alert) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.subscribers[metricKey] == nil {
		am.subscribers[metricKey] = make([]chan common.Alert, 0)
	}
	am.subscribers[metricKey] = append(am.subscribers[metricKey], alertChan)
}

func (am *AlertManager) Unsubscribe(metricKey string, alertChan chan common.Alert) {
	am.mu.Lock()
	defer am.mu.Unlock()

	subs := am.subscribers[metricKey]
	for i, ch := range subs {
		if ch == alertChan {
			am.subscribers[metricKey] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
}

func (am *AlertManager) UnsubscribeAll(alertChan chan common.Alert) {
	am.mu.Lock()
	defer am.mu.Unlock()

	for key, subs := range am.subscribers {
		for i, ch := range subs {
			if ch == alertChan {
				am.subscribers[key] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
	}
}

func (am *AlertManager) notifySubscribers(metricKey string, alert common.Alert) {
	subs := am.subscribers[metricKey]
	for _, ch := range subs {
		select {
		case ch <- alert:
		default:
		}
	}
}

func (am *AlertManager) GetPendingAlerts() []common.Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	result := make([]common.Alert, len(am.pendingAlerts))
	copy(result, am.pendingAlerts)
	return result
}

func (am *AlertManager) ClearAlerts() {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.pendingAlerts = make([]common.Alert, 0)
}
