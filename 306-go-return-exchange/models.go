package main

import (
	"time"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCompleted OrderStatus = "completed"
)

type ReturnReason string

const (
	ReturnReasonQualityIssue   ReturnReason = "质量问题"
	ReturnReasonDamaged        ReturnReason = "商品破损"
	ReturnReasonNotMatch       ReturnReason = "与描述不符"
	ReturnReasonSizeNotFit     ReturnReason = "尺寸不合适"
	ReturnReasonDontWant       ReturnReason = "不想要了"
	ReturnReasonWrongOrder     ReturnReason = "拍错了"
)

type ReturnExchangeType string

const (
	TypeReturn    ReturnExchangeType = "return"
	TypeExchange  ReturnExchangeType = "exchange"
)

type ReturnExchangeStatus string

const (
	StatusPending    ReturnExchangeStatus = "待审核"
	StatusApproved   ReturnExchangeStatus = "已同意"
	StatusRejected   ReturnExchangeStatus = "已拒绝"
	StatusProcessing ReturnExchangeStatus = "处理中"
	StatusCompleted  ReturnExchangeStatus = "已完成"
)

type Order struct {
	OrderID        string      `json:"order_id"`
	UserID         string      `json:"user_id"`
	SKU            string      `json:"sku"`
	Spec           string      `json:"spec"`
	Quantity       int         `json:"quantity"`
	UnitPrice      float64     `json:"unit_price"`
	TotalAmount    float64     `json:"total_amount"`
	PaidAmount     float64     `json:"paid_amount"`
	ShippingFee    float64     `json:"shipping_fee"`
	Status         OrderStatus `json:"status"`
	CreatedAt      time.Time   `json:"created_at"`
}

type ReturnExchangeApplication struct {
	ApplicationID      string               `json:"application_id"`
	OrderID            string               `json:"order_id"`
	UserID             string               `json:"user_id"`
	Type               ReturnExchangeType   `json:"type"`
	Reason             ReturnReason         `json:"reason"`
	ReasonDetail       string               `json:"reason_detail"`
	NewSKU             string               `json:"new_sku,omitempty"`
	NewSpec            string               `json:"new_spec,omitempty"`
	Status             ReturnExchangeStatus `json:"status"`
	RejectReason       string               `json:"reject_reason,omitempty"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
}

type Voucher struct {
	VoucherID     string    `json:"voucher_id"`
	ApplicationID string    `json:"application_id"`
	ImageBase64   string    `json:"image_base64"`
	CreatedAt     time.Time `json:"created_at"`
}

type RefundRecord struct {
	RefundID      string    `json:"refund_id"`
	ApplicationID string    `json:"application_id"`
	OrderID       string    `json:"order_id"`
	RefundAmount  float64   `json:"refund_amount"`
	ShippingFee   float64   `json:"shipping_fee"`
	TotalRefund   float64   `json:"total_refund"`
	CreatedAt     time.Time `json:"created_at"`
}

type ShippingOrder struct {
	ShippingID      string    `json:"shipping_id"`
	ApplicationID   string    `json:"application_id"`
	OrderID         string    `json:"order_id"`
	SKU             string    `json:"sku"`
	Spec            string    `json:"spec"`
	Quantity        int       `json:"quantity"`
	PriceDifference float64   `json:"price_difference"`
	NeedsTopUp      bool      `json:"needs_top_up"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

type CreateReturnRequest struct {
	OrderID      string       `json:"order_id"`
	UserID       string       `json:"user_id"`
	Reason       ReturnReason `json:"reason"`
	ReasonDetail string       `json:"reason_detail"`
	Vouchers     []string     `json:"vouchers,omitempty"`
}

type CreateExchangeRequest struct {
	OrderID      string       `json:"order_id"`
	UserID       string       `json:"user_id"`
	Reason       ReturnReason `json:"reason"`
	ReasonDetail string       `json:"reason_detail"`
	NewSKU       string       `json:"new_sku"`
	NewSpec      string       `json:"new_spec"`
	Vouchers     []string     `json:"vouchers,omitempty"`
}

type ReviewRequest struct {
	ApplicationID string `json:"application_id"`
	Approved      bool   `json:"approved"`
	RejectReason  string `json:"reject_reason,omitempty"`
}

type SKUPrice struct {
	SKU   string  `json:"sku"`
	Spec  string  `json:"spec"`
	Price float64 `json:"price"`
}
