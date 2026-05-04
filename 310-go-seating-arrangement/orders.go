package main

import (
	"errors"
	"time"
)

var (
	ErrOrderNotFound     = errors.New("order not found")
	ErrOrderNotPending   = errors.New("order is not pending")
	ErrOrderAlreadyPaid  = errors.New("order is already paid")
)

type CreateOrderRequest struct {
	EventID   string `json:"event_id"`
	Section   string `json:"section"`
	RowNumber int    `json:"row_number"`
	SeatIndex int    `json:"seat_index"`
}

func CreateOrder(req CreateOrderRequest) (*Order, error) {
	event, err := GetEvent(req.EventID)
	if err != nil {
		return nil, err
	}

	venue, err := GetVenue(event.VenueID)
	if err != nil {
		return nil, err
	}

	_, err = ValidateSeatLocation(venue, req.Section, req.RowNumber, req.SeatIndex)
	if err != nil {
		return nil, err
	}

	orderMutex.Lock()
	defer orderMutex.Unlock()

	orderID := generateID()
	seatKey := generateSeatKey(req.Section, req.RowNumber, req.SeatIndex)

	_, err = LockSeat(req.EventID, req.Section, req.RowNumber, req.SeatIndex, orderID)
	if err != nil {
		return nil, err
	}

	order := &Order{
		ID:         orderID,
		EventID:    req.EventID,
		Section:    req.Section,
		RowNumber:  req.RowNumber,
		SeatIndex:  req.SeatIndex,
		SeatKey:    seatKey,
		Status:     OrderStatusPending,
		CreatedAt:  time.Now(),
	}

	orders[orderID] = order

	return order, nil
}

func PayOrder(orderID string) (*Order, error) {
	orderMutex.Lock()
	defer orderMutex.Unlock()

	order, exists := orders[orderID]
	if !exists {
		return nil, ErrOrderNotFound
	}

	if order.Status == OrderStatusPaid {
		return nil, ErrOrderAlreadyPaid
	}

	if order.Status == OrderStatusCancelled {
		return nil, ErrOrderNotPending
	}

	err := ConfirmSeatSale(order.EventID, order.SeatKey)
	if err != nil {
		return nil, err
	}

	order.Status = OrderStatusPaid
	order.PaidAt = time.Now()

	return order, nil
}

func CancelOrder(orderID string) (*Order, error) {
	orderMutex.Lock()
	defer orderMutex.Unlock()

	order, exists := orders[orderID]
	if !exists {
		return nil, ErrOrderNotFound
	}

	if order.Status == OrderStatusPaid {
		return nil, ErrOrderAlreadyPaid
	}

	if order.Status == OrderStatusCancelled {
		return order, nil
	}

	err := ReleaseSeat(order.EventID, order.SeatKey)
	if err != nil && err != ErrSeatNotLocked {
		return nil, err
	}

	order.Status = OrderStatusCancelled
	order.CancelledAt = time.Now()

	return order, nil
}

func GetOrder(orderID string) (*Order, error) {
	orderMutex.RLock()
	defer orderMutex.RUnlock()

	order, exists := orders[orderID]
	if !exists {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

func GetAllOrders() []*Order {
	orderMutex.RLock()
	defer orderMutex.RUnlock()

	result := make([]*Order, 0, len(orders))
	for _, o := range orders {
		result = append(result, o)
	}
	return result
}

func GetOrdersByEvent(eventID string) []*Order {
	orderMutex.RLock()
	defer orderMutex.RUnlock()

	var result []*Order
	for _, o := range orders {
		if o.EventID == eventID {
			result = append(result, o)
		}
	}
	return result
}
