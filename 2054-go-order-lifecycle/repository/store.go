package repository

import (
	"order-lifecycle/models"
	"sync"
	"time"
)

type TxState struct {
	orders    map[string]*models.Order
	inventory map[string]int
	histories []*models.OrderStatusHistory
}

type Store struct {
	mu        sync.RWMutex
	orders    map[string]*models.Order
	histories map[string][]*models.OrderStatusHistory
	inventory map[string]int

	txMu sync.Mutex
	open bool
	tx   *TxState
}

func NewStore() *Store {
	return &Store{
		orders:    make(map[string]*models.Order),
		histories: make(map[string][]*models.OrderStatusHistory),
		inventory: make(map[string]int),
		open:      false,
	}
}

func (s *Store) Begin() (*TxState, error) {
	s.txMu.Lock()
	defer s.txMu.Unlock()
	if s.open {
		return nil, nil
	}
	s.mu.RLock()
	s.tx = &TxState{
		orders:    make(map[string]*models.Order),
		inventory: make(map[string]int),
		histories: nil,
	}
	for k, v := range s.orders {
		orderCopy := *v
		s.tx.orders[k] = &orderCopy
	}
	for k, v := range s.inventory {
		s.tx.inventory[k] = v
	}
	s.mu.RUnlock()
	s.open = true
	return s.tx, nil
}

func (s *Store) Commit() error {
	s.txMu.Lock()
	defer s.txMu.Unlock()
	if !s.open {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range s.tx.orders {
		s.orders[k] = v
	}
	for k, v := range s.tx.inventory {
		s.inventory[k] = v
	}
	for _, h := range s.tx.histories {
		s.histories[h.OrderID] = append(s.histories[h.OrderID], h)
	}
	s.tx = nil
	s.open = false
	return nil
}

func (s *Store) Rollback() {
	s.txMu.Lock()
	defer s.txMu.Unlock()
	if !s.open {
		return
	}
	s.tx = nil
	s.open = false
}

func (s *Store) GetOrder(id string) (*models.Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.open && s.tx != nil {
		if order, ok := s.tx.orders[id]; ok {
			return order, true
		}
		return nil, false
	}
	order, ok := s.orders[id]
	return order, ok
}

func (s *Store) SaveOrder(order *models.Order) {
	order.UpdatedAt = time.Now()
	if s.open && s.tx != nil {
		orderCopy := *order
		s.tx.orders[order.ID] = &orderCopy
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[order.ID] = order
}

func (s *Store) GetInventory(productID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.open && s.tx != nil {
		return s.tx.inventory[productID]
	}
	return s.inventory[productID]
}

func (s *Store) SetInventory(productID string, quantity int) {
	if s.open && s.tx != nil {
		s.tx.inventory[productID] = quantity
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.inventory[productID] = quantity
}

func (s *Store) AddInventory(productID string, quantity int) {
	if s.open && s.tx != nil {
		s.tx.inventory[productID] += quantity
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.inventory[productID] += quantity
}

func (s *Store) AddHistory(history *models.OrderStatusHistory) {
	history.CreatedAt = time.Now()
	if s.open && s.tx != nil {
		s.tx.histories = append(s.tx.histories, history)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.histories[history.OrderID] = append(s.histories[history.OrderID], history)
}

func (s *Store) GetHistories(orderID string) []*models.OrderStatusHistory {
	s.mu.RLock()
	defer s.mu.RUnlock()
	histories := s.histories[orderID]
	result := make([]*models.OrderStatusHistory, len(histories))
	for i, h := range histories {
		hCopy := *h
		result[i] = &hCopy
	}
	return result
}

func (s *Store) GetExpiredOrders() []*models.Order {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cutoff := time.Now().Add(-30 * time.Minute)
	var expired []*models.Order
	for _, order := range s.orders {
		if order.Status == models.StatusCreated && order.CreatedAt.Before(cutoff) {
			expired = append(expired, order)
		}
	}
	return expired
}
