package core

import (
	"errors"
	"time"
)

var (
	ErrOrderNotFound        = errors.New("order not found")
	ErrEmptyOrderItems      = errors.New("order items cannot be empty")
	ErrInvalidQuantity      = errors.New("invalid quantity")
	ErrPaymentMethodEmpty   = errors.New("payment method cannot be empty")
	ErrOrderAlreadyClosed   = errors.New("order is from a closed day and cannot be modified")
)

type OrderInputItem struct {
	ProductID string
	Quantity  int
}

func (s *Store) CreateOrder(cashierID, memberID string, items []OrderInputItem, paymentMethod string) (*Order, error) {
	if len(items) == 0 {
		return nil, ErrEmptyOrderItems
	}
	if paymentMethod == "" {
		return nil, ErrPaymentMethodEmpty
	}

	s.Lock()
	defer s.Unlock()

	orderItems := make([]OrderItem, 0, len(items))
	var totalAmount int64 = 0
	stockInsufficient := false

	for _, item := range items {
		if item.Quantity <= 0 {
			return nil, ErrInvalidQuantity
		}

		product, exists := s.products[item.ProductID]
		if !exists {
			return nil, ErrProductNotFound
		}

		subTotal := product.Price * int64(item.Quantity)
		totalAmount += subTotal

		orderItems = append(orderItems, OrderItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			Price:       product.Price,
			Quantity:    item.Quantity,
			SubTotal:    subTotal,
		})

		if product.StockQty < item.Quantity {
			stockInsufficient = true
		}
	}

	var discountAmount int64 = 0
	var discountRate float64 = 1.0

	if memberID != "" {
		member, exists := s.members[memberID]
		if !exists {
			return nil, ErrMemberNotFound
		}

		level, exists := s.memberLevels[member.MemberLevelID]
		if exists {
			discountRate = level.DiscountRate
		}
	}

	if discountRate < 1.0 {
		discountAmount = totalAmount - int64(float64(totalAmount)*discountRate)
	}

	payableAmount := totalAmount - discountAmount
	now := time.Now()

	stockStatus := StockStatusNormal
	if stockInsufficient {
		stockStatus = StockStatusPending
	}

	orderID := s.idGen.Generate()
	order := &Order{
		ID:             orderID,
		CashierID:      cashierID,
		MemberID:       memberID,
		Items:          orderItems,
		TotalAmount:    totalAmount,
		DiscountAmount: discountAmount,
		PayableAmount:  payableAmount,
		PaymentMethod:  paymentMethod,
		PaymentTime:    now,
		StockStatus:    stockStatus,
		CreatedAt:      now,
	}

	s.orders[orderID] = order

	if !stockInsufficient {
		for _, item := range items {
			product := s.products[item.ProductID]
			product.StockQty -= item.Quantity
			product.UpdatedAt = now

			if s.hasActiveStocktaking() {
				s.addPendingOperationToActiveStocktaking(product.ID, item.Quantity, PendingOpTypeSale, now)
			}
		}
	}

	return order, nil
}

func (s *Store) GetOrder(id string) (*Order, error) {
	s.RLock()
	defer s.RUnlock()

	order, exists := s.orders[id]
	if !exists {
		return nil, ErrOrderNotFound
	}

	orderCopy := *order
	orderCopy.Items = make([]OrderItem, len(order.Items))
	copy(orderCopy.Items, order.Items)

	return &orderCopy, nil
}

func (s *Store) GetOrdersByDate(dateStr string) []Order {
	s.RLock()
	defer s.RUnlock()

	var orders []Order
	for _, order := range s.orders {
		orderDate := order.PaymentTime.Format("2006-01-02")
		if orderDate == dateStr {
			orderCopy := *order
			orderCopy.Items = make([]OrderItem, len(order.Items))
			copy(orderCopy.Items, order.Items)
			orders = append(orders, orderCopy)
		}
	}
	return orders
}

func (s *Store) hasActiveStocktaking() bool {
	for _, st := range s.stocktakings {
		if st.Status != StocktakingStatusCompleted {
			return true
		}
	}
	return false
}

func (s *Store) addPendingOperationToActiveStocktaking(productID string, quantity int, opType PendingOperationType, opTime time.Time) {
	for id, st := range s.stocktakings {
		if st.Status != StocktakingStatusCompleted {
			op := &PendingOperation{
				ID:            s.idGen.Generate(),
				StocktakingID: id,
				OperationType: opType,
				ProductID:     productID,
				Quantity:      quantity,
				OperationTime: opTime,
			}
			s.pendingOps[id] = append(s.pendingOps[id], op)
		}
	}
}
