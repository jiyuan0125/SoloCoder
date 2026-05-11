package lockservice

import (
	"sync"
	"time"
)

type OrderManager struct {
	mu     sync.RWMutex
	orders map[string]*Order
}

func NewOrderManager() *OrderManager {
	return &OrderManager{
		orders: make(map[string]*Order),
	}
}

func (om *OrderManager) CreateOrder(order *Order) {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.orders[order.ID] = order
}

func (om *OrderManager) GetOrder(id string) *Order {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return om.orders[id]
}

func (om *OrderManager) ListOrders() []*Order {
	om.mu.RLock()
	defer om.mu.RUnlock()
	result := make([]*Order, 0, len(om.orders))
	for _, o := range om.orders {
		result = append(result, o)
	}
	return result
}

func (om *OrderManager) ListByMaster(masterID string) []*Order {
	om.mu.RLock()
	defer om.mu.RUnlock()
	var result []*Order
	for _, o := range om.orders {
		if o.MasterID == masterID {
			result = append(result, o)
		}
	}
	return result
}

func (om *OrderManager) UpdateStatus(id string, status OrderStatus) bool {
	om.mu.Lock()
	defer om.mu.Unlock()
	if o, exists := om.orders[id]; exists {
		o.Status = status
		switch status {
		case OrderStatusAccepted:
			o.AcceptedAt = time.Now()
		case OrderStatusInService:
			o.InServiceAt = time.Now()
		case OrderStatusPendingPayment:
			o.PendingPayAt = time.Now()
		case OrderStatusCompleted:
			o.CompletedAt = time.Now()
		}
		return true
	}
	return false
}

func (om *OrderManager) UpdateDetail(id string, detail OrderDetail) bool {
	om.mu.Lock()
	defer om.mu.Unlock()
	if o, exists := om.orders[id]; exists {
		o.Detail = detail
		return true
	}
	return false
}

func (om *OrderManager) AddRating(id string, rating int, comment string) (bool, bool) {
	om.mu.Lock()
	defer om.mu.Unlock()
	if o, exists := om.orders[id]; exists {
		if rating < 1 || rating > 5 {
			return false, false
		}
		o.Rating = rating
		o.Comment = comment
		hasComplaint := rating < 3
		o.HasComplaint = hasComplaint
		return true, hasComplaint
	}
	return false, false
}
