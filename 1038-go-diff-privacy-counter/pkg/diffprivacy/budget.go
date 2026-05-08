package diffprivacy

import (
	"errors"
	"sync"
)

type BudgetManager struct {
	mu              sync.RWMutex
	totalBudget     float64
	remainingBudget float64
}

func NewBudgetManager(initialBudget float64) *BudgetManager {
	return &BudgetManager{
		totalBudget:     initialBudget,
		remainingBudget: initialBudget,
	}
}

func (b *BudgetManager) Remaining() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.remainingBudget
}

func (b *BudgetManager) Total() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.totalBudget
}

func (b *BudgetManager) Consume(epsilon float64) error {
	if epsilon <= 0 {
		return nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.remainingBudget < epsilon {
		return errors.New("insufficient privacy budget")
	}

	b.remainingBudget -= epsilon
	return nil
}

func (b *BudgetManager) Recharge(amount float64) error {
	if amount <= 0 {
		return errors.New("recharge amount must be positive")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.remainingBudget += amount
	b.totalBudget += amount
	return nil
}
