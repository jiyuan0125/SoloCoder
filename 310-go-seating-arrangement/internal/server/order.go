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
	om.mu.Lock()
	defer om.mu.Unlock()

	order, exists := om.orders[orderID]
	if !exists {
		return errors.New("订单不存在")
	}

	if order.Status != protocol.OrderStatusCreated {
		return errors.New("订单状态不正确，无法确认支付")
	}

	session, exists := om.sessionManager.GetSession(order.SessionID)
	if !exists {
		return errors.New("场次不存在")
	}

	seatState, exists := session.SeatStates[order.SeatID]
	if !exists {
		return errors.New("座位不存在")
	}

	now := time.Now()
	if seatState.Status == protocol.SeatStatusLocked && seatState.LockedAt != nil {
		if now.Sub(*seatState.LockedAt) >= LockTimeout {
			return errors.New("座位锁定已超时，请重新选座")
		}
	} else if seatState.Status != protocol.SeatStatusLocked {
		return errors.New("座位未被锁定，无法确认支付")
	}

	order.Status = protocol.OrderStatusPaid
	order.UpdatedAt = time.Now()

	seatState.Status = protocol.SeatStatusSold
	seatState.LockedAt = nil

	return nil
}

func (om *OrderManager) CancelOrder(orderID string) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	order, exists := om.orders[orderID]
	if !exists {
		return errors.New("订单不存在")
	}

	if order.Status == protocol.OrderStatusPaid {
		return errors.New("已支付的订单无法取消")
	}

	if order.Status == protocol.OrderStatusCancelled {
		return errors.New("订单已取消")
	}

	session, exists := om.sessionManager.GetSession(order.SessionID)
	if !exists {
		return errors.New("场次不存在")
	}

	seatState, exists := session.SeatStates[order.SeatID]
	if !exists {
		return errors.New("座位不存在")
	}

	order.Status = protocol.OrderStatusCancelled
	order.UpdatedAt = time.Now()

	if seatState.Status == protocol.SeatStatusLocked {
		seatState.Status = protocol.SeatStatusAvailable
		seatState.LockedAt = nil
		seatState.OrderID = ""
	}

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
		if order.SessionID == sessionID {
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
