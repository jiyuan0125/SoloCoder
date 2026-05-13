package repository

import (
	"fmt"
	"time"

	"order-lifecycle/models"
)

type OrderRepository struct {
	store *Store
}

func NewOrderRepository(store *Store) *OrderRepository {
	return &OrderRepository{store: store}
}

func (r *OrderRepository) BeginTx() error {
	_, err := r.store.Begin()
	return err
}

func (r *OrderRepository) Commit() error {
	return r.store.Commit()
}

func (r *OrderRepository) Rollback() {
	r.store.Rollback()
}

func (r *OrderRepository) Create(order *models.Order) error {
	if _, exists := r.store.GetOrder(order.ID); exists {
		return fmt.Errorf("订单已存在")
	}
	now := time.Now()
	order.CreatedAt = now
	order.UpdatedAt = now
	r.store.SaveOrder(order)
	history := &models.OrderStatusHistory{
		ID:         generateHistoryID(),
		OrderID:    order.ID,
		FromStatus: nil,
		ToStatus:   order.Status,
		Event:      "订单创建",
	}
	r.store.AddHistory(history)
	return nil
}

func (r *OrderRepository) GetByID(id string) (*models.Order, error) {
	order, ok := r.store.GetOrder(id)
	if !ok {
		return nil, nil
	}
	orderCopy := *order
	return &orderCopy, nil
}

func (r *OrderRepository) UpdateStatus(order *models.Order, fromStatus, toStatus models.OrderStatus, event string) error {
	current, ok := r.store.GetOrder(order.ID)
	if !ok {
		return fmt.Errorf("订单不存在")
	}
	if current.Status != fromStatus {
		return fmt.Errorf("状态不一致，并发冲突")
	}
	current.Status = toStatus
	r.store.SaveOrder(current)
	history := &models.OrderStatusHistory{
		ID:         generateHistoryID(),
		OrderID:    order.ID,
		FromStatus: &fromStatus,
		ToStatus:   toStatus,
		Event:      event,
	}
	r.store.AddHistory(history)
	order.Status = toStatus
	order.UpdatedAt = current.UpdatedAt
	return nil
}

func (r *OrderRepository) UpdatePaymentMethod(order *models.Order, method models.PaymentMethod) error {
	current, ok := r.store.GetOrder(order.ID)
	if !ok {
		return fmt.Errorf("订单不存在")
	}
	current.PaymentMethod = &method
	r.store.SaveOrder(current)
	order.PaymentMethod = &method
	order.UpdatedAt = current.UpdatedAt
	return nil
}

func (r *OrderRepository) UpdateTrackingNumber(order *models.Order, trackingNumber string) error {
	current, ok := r.store.GetOrder(order.ID)
	if !ok {
		return fmt.Errorf("订单不存在")
	}
	current.TrackingNumber = &trackingNumber
	r.store.SaveOrder(current)
	order.TrackingNumber = &trackingNumber
	order.UpdatedAt = current.UpdatedAt
	return nil
}

func (r *OrderRepository) UpdateDeliveredAt(order *models.Order) error {
	current, ok := r.store.GetOrder(order.ID)
	if !ok {
		return fmt.Errorf("订单不存在")
	}
	now := time.Now()
	current.DeliveredAt = &now
	r.store.SaveOrder(current)
	order.DeliveredAt = &now
	order.UpdatedAt = current.UpdatedAt
	return nil
}

func (r *OrderRepository) GetHistory(orderID string) ([]models.OrderStatusHistory, error) {
	histories := r.store.GetHistories(orderID)
	result := make([]models.OrderStatusHistory, 0, len(histories))
	for _, h := range histories {
		result = append(result, *h)
	}
	return result, nil
}

func (r *OrderRepository) GetExpiredOrders() ([]*models.Order, error) {
	return r.store.GetExpiredOrders(), nil
}

func generateHistoryID() string {
	return "H" + time.Now().Format("20060102150405.000000000")
}
