package main

import (
	"sync"

	"realtime-dashboard/common"
)

type LayoutManager struct {
	layouts map[string]*common.DashboardLayout
	mu      sync.RWMutex
}

func NewLayoutManager() *LayoutManager {
	return &LayoutManager{
		layouts: make(map[string]*common.DashboardLayout),
	}
}

func (lm *LayoutManager) SaveLayout(clientID string, metrics []string, order []int) bool {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if len(order) > 0 && len(order) != len(metrics) {
		return false
	}

	lm.layouts[clientID] = &common.DashboardLayout{
		ClientID: clientID,
		Metrics:  metrics,
		Order:    order,
	}

	return true
}

func (lm *LayoutManager) GetLayout(clientID string) (*common.DashboardLayout, bool) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	layout, exists := lm.layouts[clientID]
	if !exists {
		return nil, false
	}

	result := &common.DashboardLayout{
		ClientID: layout.ClientID,
		Metrics:  make([]string, len(layout.Metrics)),
		Order:    make([]int, len(layout.Order)),
	}
	copy(result.Metrics, layout.Metrics)
	copy(result.Order, layout.Order)

	return result, true
}

func (lm *LayoutManager) DeleteLayout(clientID string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	delete(lm.layouts, clientID)
}

func (lm *LayoutManager) GetOrderedMetrics(clientID string, availableMetrics map[string]*common.Metric) []*common.Metric {
	layout, exists := lm.GetLayout(clientID)
	if !exists {
		result := make([]*common.Metric, 0, len(availableMetrics))
		for _, m := range availableMetrics {
			result = append(result, m)
		}
		return result
	}

	result := make([]*common.Metric, 0, len(layout.Metrics))
	for _, key := range layout.Metrics {
		if metric, exists := availableMetrics[key]; exists {
			result = append(result, metric)
		}
	}

	return result
}
