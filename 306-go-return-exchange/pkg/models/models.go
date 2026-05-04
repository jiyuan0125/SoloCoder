package models

import "time"

type ReturnExchangeType string

const (
	ReturnType    ReturnExchangeType = "return"
	ExchangeType  ReturnExchangeType = "exchange"
)

type ReturnReason string

const (
	ReasonQualityIssue     ReturnReason = "quality_issue"
	ReasonDamaged          ReturnReason = "damaged"
	ReasonWrongSize        ReturnReason = "wrong_size"
	ReasonNotWanted        ReturnReason = "not_wanted"
	ReasonWrongOrder       ReturnReason = "wrong_order"
	ReasonDescriptionMismatch ReturnReason = "description_mismatch"
)

type ApplicationStatus string

const (
	StatusPending    ApplicationStatus = "pending"
	StatusApproved   ApplicationStatus = "approved"
	StatusRejected   ApplicationStatus = "rejected"
	StatusProcessing ApplicationStatus = "processing"
	StatusCompleted  ApplicationStatus = "completed"
)

type Order struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	SKU            string    `json:"sku"`
	Specification  string    `json:"specification"`
	Price          float64   `json:"price"`
	ShippingFee    float64   `json:"shipping_fee"`
	TotalAmount    float64   `json:"total_amount"`
	CreateTime     time.Time `json:"create_time"`
}

type ReturnExchangeApplication struct {
	ID                      string               `json:"id"`
	OrderID                 string               `json:"order_id"`
	UserID                  string               `json:"user_id"`
	Type                    ReturnExchangeType   `json:"type"`
	Reason                  ReturnReason         `json:"reason"`
	EvidenceImages          []string             `json:"evidence_images"`
	NewSKU                  string               `json:"new_sku,omitempty"`
	NewSpecification        string               `json:"new_specification,omitempty"`
	NewSKUPrice             float64              `json:"new_sku_price,omitempty"`
	Status                  ApplicationStatus    `json:"status"`
	RefundID                string               `json:"refund_id,omitempty"`
	RefundAmount            float64              `json:"refund_amount,omitempty"`
	ShippingOrderID         string               `json:"shipping_order_id,omitempty"`
	ShippingFee             float64              `json:"shipping_fee,omitempty"`
	ShippingFeePaidBy       string               `json:"shipping_fee_paid_by,omitempty"`
	PriceDifference         float64              `json:"price_difference,omitempty"`
	PriceDifferenceHandled  bool                 `json:"price_difference_handled"`
	PriceDifferenceRefundID string              `json:"price_difference_refund_id,omitempty"`
	CreateTime              time.Time            `json:"create_time"`
	UpdateTime              time.Time            `json:"update_time"`
}

type RefundRecord struct {
	ID              string    `json:"id"`
	ApplicationID   string    `json:"application_id"`
	OrderID         string    `json:"order_id"`
	UserID          string    `json:"user_id"`
	Amount          float64   `json:"amount"`
	ShippingFee     float64   `json:"shipping_fee,omitempty"`
	TotalRefund     float64   `json:"total_refund"`
	CreateTime      time.Time `json:"create_time"`
}

type ShippingOrder struct {
	ID              string    `json:"id"`
	ApplicationID   string    `json:"application_id"`
	OrderID         string    `json:"order_id"`
	UserID          string    `json:"user_id"`
	NewSKU          string    `json:"new_sku"`
	NewSpecification string   `json:"new_specification"`
	CreateTime      time.Time `json:"create_time"`
}
