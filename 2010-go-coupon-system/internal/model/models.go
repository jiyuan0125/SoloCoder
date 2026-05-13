package model

import "time"

type Coupon struct {
	ID           int64     `json:"id"`
	BatchID      string    `json:"batch_id"`
	Denomination int64     `json:"denomination"`
	Threshold    int64     `json:"threshold"`
	TotalCount   int64     `json:"total_count"`
	ClaimedCount int64     `json:"claimed_count"`
	UsedCount    int64     `json:"used_count"`
	LimitPerUser int64     `json:"limit_per_user"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserCoupon struct {
	ID            int64      `json:"id"`
	UserID        string     `json:"user_id"`
	CouponID      int64      `json:"coupon_id"`
	BatchID       string     `json:"batch_id"`
	Status        string     `json:"status"`
	ClaimedAt     time.Time  `json:"claimed_at"`
	RedeemedAt    *time.Time `json:"redeemed_at,omitempty"`
	OrderID       *string    `json:"order_id,omitempty"`
	RedeemedAmount *int64    `json:"redeemed_amount,omitempty"`
}

type RedemptionRecord struct {
	ID             int64     `json:"id"`
	UserCouponID   int64     `json:"user_coupon_id"`
	CouponID       int64     `json:"coupon_id"`
	BatchID        string    `json:"batch_id"`
	UserID         string    `json:"user_id"`
	OrderID        string    `json:"order_id"`
	OrderAmount    int64     `json:"order_amount"`
	RedeemedAmount int64     `json:"redeemed_amount"`
	RedeemedAt     time.Time `json:"redeemed_at"`
	IsRefunded     bool      `json:"is_refunded"`
}

type CouponStats struct {
	ID            int64     `json:"id"`
	BatchID       string    `json:"batch_id"`
	TotalCount    int64     `json:"total_count"`
	ClaimedCount  int64     `json:"claimed_count"`
	UsedCount     int64     `json:"used_count"`
	UsageRate     float64   `json:"usage_rate"`
	UpdatedAt     time.Time `json:"updated_at"`
}

const (
	CouponStatusActive   = "active"
	CouponStatusExpired  = "expired"
	CouponStatusExhausted = "exhausted"
	
	UserCouponStatusClaimed  = "claimed"
	UserCouponStatusRedeemed = "redeemed"
	UserCouponStatusExpired  = "expired"
)
