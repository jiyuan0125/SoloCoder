package server

import (
	"errors"
	"sync"
	"time"

	"seating-arrangement/pkg/protocol"
)

type OrderManager struct {
	orders         map[string]*protocol.Order
	sessionManager *SessionManager
	mu             sync.RWMutex
}

func NewOrderManager(sm *SessionManager) *OrderManager {
	return &OrderManager{
		orders:         make(map[string]*protocol.Order),
		sessionManager: sm,
	}
}

func (om *OrderManager) GetOrder(id string) (*protocol.Order, bool) {
	om.mu.RLock()
	defer om.mu.RUnlock()
	order, exists := om.orders[id]
	return order, exists
}

func (om *OrderManager) ConfirmOrder(orderID string) error {
	om.mu.RLock()
	order, exists := om.orders[orderID]
	if !exists {
		om.mu.RUnlock()
		return errors.New("订单不存在")
	}
	if order.Status != protocol.OrderStatusCreated {
		om.mu.RUnlock()
		return errors.New("订单状态不正确，无法确认支付")
	}
	sessionID := order.SessionID
	seatID := order.SeatID
	om.mu.RUnlock()

	if err := om.sessionManager.ConfirmSeatSold(sessionID, seatID); err != nil {
		return err
	}

	om.mu.Lock()
	defer om.mu.Unlock()

	order, exists = om.orders[orderID]
	if !exists {
		return errors.New("订单不存在")
	}
	if order.Status != protocol.OrderStatusCreated {
		return errors.New("订单状态已改变")
	}

	order.Status = protocol.OrderStatusPaid
	order.UpdatedAt = time.Now()
	return nil
}

func (om *OrderManager) CancelOrder(orderID string) error {
	om.mu.RLock()
	order, exists := om.orders[orderID]
	if !exists {
		om.mu.RUnlock()
		return errors.New("订单不存在")
	}
	if order.Status == protocol.OrderStatusPaid {
		om.mu.RUnlock()
		return errors.New("已支付的订单无法取消")
	}
	if order.Status == protocol.OrderStatusCancelled {
		om.mu.RUnlock()
		return errors.New("订单已取消")
	}
	sessionID := order.SessionID
	seatID := order.SeatID
	om.mu.RUnlock()

	if err := om.sessionManager.ReleaseSeatFromOrder(sessionID, seatID); err != nil {
		if err.Error() != "座位当前未被锁定" {
			return err
		}
	}

	om.mu.Lock()
	defer om.mu.Unlock()

	order, exists = om.orders[orderID]
	if !exists {
		return errors.New("订单不存在")
	}
	if order.Status == protocol.OrderStatusPaid {
		return errors.New("已支付的订单无法取消")
	}
	if order.Status == protocol.OrderStatusCancelled {
		return errors.New("订单已取消")
	}

	order.Status = protocol.OrderStatusCancelled
	order.UpdatedAt = time.Now()
	return nil
}

func (om *OrderManager) AddOrder(order *protocol.Order) {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.orders[order.ID] = order
}

func (om *OrderManager) GetOrdersBySession(sessionID string) []*protocol.Order {
	om.mu.RLock()
	defer om.mu.RUnlock()

	var orders []*protocol.Order
	for _, order := range om.orders {
		if order.SessionID == sessionID {
			orders = append(orders, order)
		}
	}
	return orders
}

func (om *OrderManager) SetOrders(orders map[string]*protocol.Order) {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.orders = orders
}

func (om *OrderManager) GetOrders() map[string]*protocol.Order {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return om.orders
}

func (om *OrderManager) GetOrderDetails(sessionID string) []*protocol.OrderDetail {
	om.mu.RLock()
	defer om.mu.RUnlock()

	var details []*protocol.OrderDetail
	for _, order := range om.orders {
		if order.SessionID == sessionID && order.Status == protocol.OrderStatusPaid {
			detail := &protocol.OrderDetail{
				ID:          order.ID,
				SessionID:   order.SessionID,
				SectionName: order.SectionName,
				SeatID:      order.SeatID,
				Status:      order.Status,
				CreatedAt:   order.CreatedAt,
			}
			details = append(details, detail)
		}
	}
	return details
}
