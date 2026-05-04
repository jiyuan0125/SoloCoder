package main

import (
	"time"
)

type CouponBatch struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	DiscountAmount  float64   `json:"discount_amount"`
	ThresholdAmount float64   `json:"threshold_amount"`
	TotalQuantity   int       `json:"total_quantity"`
	ClaimedQuantity int       `json:"claimed_quantity"`
	RedeemedQuantity int      `json:"redeemed_quantity"`
	ValidFrom       time.Time `json:"valid_from"`
	ValidTo         time.Time `json:"valid_to"`
	MaxPerUser      int       `json:"max_per_user"`
	CreatedAt       time.Time `json:"created_at"`
}

type CouponClaim struct {
	ID           string     `json:"id"`
	BatchID      string     `json:"batch_id"`
	UserID       string     `json:"user_id"`
	CouponCode   string     `json:"coupon_code"`
	ClaimedAt    time.Time  `json:"claimed_at"`
	IsRedeemed   bool       `json:"is_redeemed"`
	RedeemedAt   *time.Time `json:"redeemed_at,omitempty"`
}

type RedeemRecord struct {
	ID             string    `json:"id"`
	BatchID        string    `json:"batch_id"`
	UserID         string    `json:"user_id"`
	CouponCode     string    `json:"coupon_code"`
	OrderID        string    `json:"order_id"`
	OrderAmount    float64   `json:"order_amount"`
	DiscountAmount float64   `json:"discount_amount"`
	RedeemedAt     time.Time `json:"redeemed_at"`
}
