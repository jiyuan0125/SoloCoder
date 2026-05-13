package models

import (
	"errors"
	"strings"
	"time"
)

type TicketType string
type TicketStatus string

const (
	TicketTypeReturn  TicketType = "return"
	TicketTypeExchange TicketType = "exchange"
	TicketTypeRepair   TicketType = "repair"
)

const (
	StatusPendingReview  TicketStatus = "pending_review"
	StatusProcessing     TicketStatus = "processing"
	StatusPendingConfirm TicketStatus = "pending_confirm"
	StatusClosed         TicketStatus = "closed"
)

type Ticket struct {
	ID            string       `json:"id"`
	Type          TicketType   `json:"type"`
	OrderID       string       `json:"order_id"`
	Status        TicketStatus `json:"status"`
	Reason        string       `json:"reason"`
	Description   string       `json:"description,omitempty"`
	CustomerID    string       `json:"customer_id"`
	AssignedTo    string       `json:"assigned_to,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	ClosedReason  string       `json:"closed_reason,omitempty"`
	ClosedAt      *time.Time   `json:"closed_at,omitempty"`
}

type Order struct {
	ID          string    `json:"id"`
	CustomerID  string    `json:"customer_id"`
	ProductSKU  string    `json:"product_sku"`
	ProductName string    `json:"product_name"`
	Quantity    int       `json:"quantity"`
	TotalAmount float64   `json:"total_amount"`
	PaymentMethod string  `json:"payment_method"`
	CompletedAt time.Time `json:"completed_at"`
}

type LogisticsOrder struct {
	ID          string    `json:"id"`
	TicketID    string    `json:"ticket_id"`
	Type        string    `json:"type"`
	Carrier     string    `json:"carrier"`
	TrackingNo  string    `json:"tracking_no"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type QCResult struct {
	ID           string    `json:"id"`
	TicketID     string    `json:"ticket_id"`
	Passed       bool      `json:"passed"`
	Reason       string    `json:"reason,omitempty"`
	Inspector    string    `json:"inspector"`
	InspectedAt  time.Time `json:"inspected_at"`
}

type RepairRecord struct {
	ID          string    `json:"id"`
	TicketID    string    `json:"ticket_id"`
	Description string    `json:"description"`
	PartsUsed   []Part    `json:"parts_used"`
	CreatedAt   time.Time `json:"created_at"`
}

type Part struct {
	SKU      string `json:"sku"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type InventoryItem struct {
	SKU      string `json:"sku"`
	Name     string `json:"name"`
	Stock    int    `json:"stock"`
}

var statusTransitions = map[TicketStatus][]TicketStatus{
	StatusPendingReview:  {StatusProcessing},
	StatusProcessing:     {StatusPendingConfirm},
	StatusPendingConfirm: {StatusClosed},
	StatusClosed:         {},
}

func (t *Ticket) CanTransitionTo(target TicketStatus) (bool, error) {
	if t.Status == StatusClosed {
		return false, errors.New("closed ticket cannot be reopened")
	}

	allowedNext := statusTransitions[t.Status]
	for _, s := range allowedNext {
		if s == target {
			return true, nil
		}
	}

	return false, errors.New("invalid status transition")
}

func (t *Ticket) GetWarrantyDays() int {
	switch t.Type {
	case TicketTypeReturn:
		return 7
	case TicketTypeExchange:
		return 15
	case TicketTypeRepair:
		return 365
	default:
		return 0
	}
}

func ValidateOrderID(orderID string) error {
	if orderID == "" {
		return errors.New("order_id is required")
	}
	if len(orderID) < 6 || len(orderID) > 20 {
		return errors.New("order_id must be 6-20 characters")
	}
	if strings.ContainsAny(orderID, " '\"<>") {
		return errors.New("order_id contains invalid characters")
	}
	return nil
}
