package database

import "time"

// 订单状态常量
const (
	StatusPending    = "pending"    // 待审核
	StatusApproved   = "approved"   // 已同意
	StatusRejected   = "rejected"   // 已拒绝
	StatusProcessing = "processing" // 处理中
	StatusCompleted  = "completed"  // 已完成
)

// 退换货类型常量
const (
	TypeReturn   = "return"   // 退货
	TypeExchange = "exchange" // 换货
)

// 运费承担方常量
const (
	BearerBuyer    = "buyer"    // 买家承担
	BearerPlatform = "platform" // 平台承担
)

// Order 订单模型
type Order struct {
	ID              int64     `json:"id"`
	OrderNo         string    `json:"order_no"`
	UserID          int64     `json:"user_id"`
	SKU             string    `json:"sku"`
	Specification   string    `json:"specification"`
	OriginalPrice   float64   `json:"original_price"`
	ActualPayment   float64   `json:"actual_payment"`
	ShippingFee     float64   `json:"shipping_fee"`
	CreatedAt       time.Time `json:"created_at"`
}

// ReturnExchange 退换货申请模型
type ReturnExchange struct {
	ID                   int64     `json:"id"`
	OrderNo              string    `json:"order_no"`
	UserID               int64     `json:"user_id"`
	Type                 string    `json:"type"` // 'return' 或 'exchange'
	Reason               string    `json:"reason"`
	ReasonDetail         string    `json:"reason_detail"`
	Status               string    `json:"status"` // pending, approved, rejected, processing, completed
	OriginalSKU          string    `json:"original_sku"`
	OriginalSpecification string   `json:"original_specification"`
	OriginalPrice        float64   `json:"original_price"`
	ActualPayment        float64   `json:"actual_payment"`
	ShippingFee          float64   `json:"shipping_fee"`
	ShippingBearer       string    `json:"shipping_bearer"` // 'buyer' 或 'platform'
	// 换货相关字段
	NewSKU               string    `json:"new_sku"`
	NewSpecification     string    `json:"new_specification"`
	NewPrice             float64   `json:"new_price"`
	PriceDifference      float64   `json:"price_difference"`
	// 审核相关
	AdminID              int64     `json:"admin_id"`
	ReviewComment        string    `json:"review_comment"`
	ReviewedAt           time.Time `json:"reviewed_at"`
	// 时间戳
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// Evidence 凭证图片模型
type Evidence struct {
	ID                int64     `json:"id"`
	ReturnExchangeID  int64     `json:"return_exchange_id"`
	ImageBase64       string    `json:"image_base64"`
	CreatedAt         time.Time `json:"created_at"`
}

// Refund 退款记录模型
type Refund struct {
	ID                int64     `json:"id"`
	ReturnExchangeID  int64     `json:"return_exchange_id"`
	RefundNo          string    `json:"refund_no"`
	RefundAmount      float64   `json:"refund_amount"`
	ShippingFeeRefund float64   `json:"shipping_fee_refund"`
	TotalRefund       float64   `json:"total_refund"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	CompletedAt       time.Time `json:"completed_at"`
}

// Shipment 发货单模型
type Shipment struct {
	ID                int64     `json:"id"`
	ReturnExchangeID  int64     `json:"return_exchange_id"`
	ShipmentNo        string    `json:"shipment_no"`
	SKU               string    `json:"sku"`
	Specification     string    `json:"specification"`
	Quantity          int       `json:"quantity"`
	ShippingAddress   string    `json:"shipping_address"`
	Status            string    `json:"status"`
	ShippedAt         time.Time `json:"shipped_at"`
	DeliveredAt       time.Time `json:"delivered_at"`
	CreatedAt         time.Time `json:"created_at"`
}
