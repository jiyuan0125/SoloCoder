package statemachine

import (
	"fmt"
	"sync"

	"order-lifecycle/models"
)

type StateTransition struct {
	FromStatus models.OrderStatus
	ToStatus   models.OrderStatus
	Event       string
}

var validTransitions = map[models.OrderStatus][]models.OrderStatus{
	models.StatusCreated: {
		models.StatusPaid, models.StatusCancelled},
	models.StatusPaid: {
		models.StatusShipped},
	models.StatusShipped: {
		models.StatusDelivered},
	models.StatusDelivered: {
		models.StatusCompleted, models.StatusRefunding},
	models.StatusRefunding: {
		models.StatusRefunded},
}

type StateMachine struct {
	mu sync.RWMutex
}

func NewStateMachine() *StateMachine {
	return &StateMachine{}
}

func (sm *StateMachine) IsValidTransition(current models.OrderStatus, next models.OrderStatus) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if nextStatuses, exists := validTransitions[current]; exists {
		for _, status := range nextStatuses {
			if status == next {
				return true
			}
		}
	}
	return false
}

func (sm *StateMachine) IsTerminal(status models.OrderStatus) bool {
	return status == models.StatusCompleted || 
		   status == models.StatusCancelled || 
		   status == models.StatusRefunded
}

func (sm *StateMachine) GetNextAllowed(status models.OrderStatus) ([]models.OrderStatus, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	nextStatuses, exists := validTransitions[status]
	if !exists {
		return nil, false
	}
	result := make([]models.OrderStatus, len(nextStatuses))
	copy(result, nextStatuses)
	return result, true
}

func (sm *StateMachine) ValidateTransition(current models.OrderStatus, next models.OrderStatus) error {
	if sm.IsTerminal(current) {
		return fmt.Errorf("当前状态 %s 是终态，无法进行状态变更", current)
	}
	if !sm.IsValidTransition(current, next) {
		return fmt.Errorf("无法从状态 %s 跳转到 %s，当前状态说明：%s", current, next, current)
	}
	return nil
}
