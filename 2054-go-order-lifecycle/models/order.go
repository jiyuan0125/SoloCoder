package models

import (
	"time"
)

type OrderStatus string

const (
	StatusCreated    OrderStatus = "CREATED"
	StatusPaid       OrderStatus = "PAID"
	StatusShipped    OrderStatus = "SHIPPED"
	StatusDelivered  OrderStatus = "DELIVERED"
	StatusCompleted  OrderStatus = "COMPLETED"
	StatusCancelled  OrderStatus = "CANCELLED"
	StatusRefunding  OrderStatus = "REFUNDING"
	StatusRefunded   OrderStatus = "REFUNDED"
)

type PaymentMethod string

const (
	PaymentMethodBalance  PaymentMethod = "BALANCE"
	PaymentMethodBankCard PaymentMethod = "BANK_CARD"
	PaymentMethodThirdParty PaymentMethod = "THIRD_PARTY"
)

type Order struct {
	ID             string
	UserID         string
	ProductID      string
	Quantity       int
	TotalAmount    float64
	Status         OrderStatus
	PaymentMethod  *PaymentMethod
	TrackingNumber *string
	DeliveredAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type OrderItem struct {
	ID        string
	OrderID   string
	ProductID string
	Quantity  int
	Price     float64
}

type OrderStatusHistory struct {
	ID        string
	OrderID   string
	FromStatus *OrderStatus
	ToStatus  OrderStatus
	Event     string
	CreatedAt time.Time
}

type Inventory struct {
	ProductID string
	Quantity  int
	UpdatedAt time.Time
}

type PaymentRequest struct {
	OrderID       string
	PaymentMethod PaymentMethod
	Amount        float64
}

type PaymentResult struct {
	Success bool
	Message string
}
