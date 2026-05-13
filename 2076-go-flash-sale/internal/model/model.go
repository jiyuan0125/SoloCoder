package model

import "time"

type ActivityStatus string

const (
	ActivityStatusNotStarted ActivityStatus = "not_started"
	ActivityStatusActive     ActivityStatus = "active"
	ActivityStatusEnded      ActivityStatus = "ended"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusPaid     OrderStatus = "paid"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Activity struct {
	ID           int64       `json:"id"`
	ProductName    string      `json:"product_name"`
	OriginalPrice float64     `json:"original_price"`
	FlashPrice    float64     `json:"flash_price"`
	TotalStock    int         `json:"total_stock"`
	StartTime     time.Time   `json:"start_time"`
	Status         ActivityStatus `json:"status"`
}

type Stock struct {
	ID             int64     `json:"id"`
	ActivityID     int64     `json:"activity_id"`
	AvailableStock int     `json:"available_stock"`
	Version       int       `json:"version"`
}

type Order struct {
	ID          int64       `json:"id"`
	OrderNo     string      `json:"order_no"`
	ActivityID  int64       `json:"activity_id"`
	UserID      string      `json:"user_id"`
	Status      OrderStatus `json:"status"`
	Price       float64     `json:"price"`
	CreatedAt   time.Time   `json:"created_at"`
	PaidAt      *time.Time  `json:"paid_at"`
	CancelledAt *time.Time  `json:"cancelled_at"`
	ExpireAt    time.Time   `json:"expire_at"`
}

type Report struct {
	ID               int64     `json:"id"`
	ActivityID        int64     `json:"activity_id"`
	TotalOrders       int       `json:"total_orders"`
	PaidOrders      int       `json:"paid_orders"`
	UnpaidOrders    int       `json:"unpaid_orders"`
	CancelledOrders int       `json:"cancelled_orders"`
	TotalRevenue    float64   `json:"total_revenue"`
	GeneratedAt     time.Time `json:"generated_at"`
	VerifiedAt    *time.Time `json:"verified_at"`
}

type CacheNotification struct {
	ID          int64     `json:"id"`
	ModuleName    string    `json:"module_name"`
	EventType     string    `json:"event_type"`
	ActivityID   int64     `json:"activity_id"`
	Payload       string    `json:"payload"`
	CreatedAt     time.Time `json:"created_at"`
	ProcessedAt   *time.Time `json:"processed_at"`
}

type QuotaAllocation struct {
	ID               int64     `json:"id"`
	ActivityID        int64     `json:"activity_id"`
	SegmentName      string    `json:"segment_name"`
	TotalQuota       int       `json:"total_quota"`
	AllocatedQuota int       `json:"allocated_quota"`
	Ratio            float64   `json:"ratio"`
}
