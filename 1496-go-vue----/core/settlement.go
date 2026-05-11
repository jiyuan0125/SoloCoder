package core

import (
	"sync"
	"time"
)

type SettlementManager struct {
	mu           sync.RWMutex
	monthlyTotals map[string]map[string]float64
}

func NewSettlementManager() *SettlementManager {
	return &SettlementManager{
		monthlyTotals: make(map[string]map[string]float64),
	}
}

func (sm *SettlementManager) getMonthKey(year, month int) string {
	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local).Format("2006-01")
}

func (sm *SettlementManager) AddToMonthlyTotal(customerID string, year, month int, amount float64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	monthKey := sm.getMonthKey(year, month)
	if sm.monthlyTotals[monthKey] == nil {
		sm.monthlyTotals[monthKey] = make(map[string]float64)
	}
	sm.monthlyTotals[monthKey][customerID] += amount
}

func (sm *SettlementManager) CalculateRebate(customerID string, year, month int) (float64, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	monthKey := sm.getMonthKey(year, month)
	totals, exists := sm.monthlyTotals[monthKey]
	if !exists {
		return 0, false
	}

	total, exists := totals[customerID]
	if !exists || total <= 5000 {
		return 0, total > 0
	}

	return roundToCents(total * 0.03), true
}

func (sm *SettlementManager) GetMonthlyTotal(customerID string, year, month int) (float64, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	monthKey := sm.getMonthKey(year, month)
	totals, exists := sm.monthlyTotals[monthKey]
	if !exists {
		return 0, false
	}

	total, exists := totals[customerID]
	return total, exists
}
