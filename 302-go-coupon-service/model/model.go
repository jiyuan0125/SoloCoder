package model

import "time"

type CouponBatch struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	DiscountAmount   int       `json:"discount_amount"`
	ThresholdAmount  int       `json:"threshold_amount"`
	TotalQuantity    int       `json:"total_quantity"`
	IssuedQuantity   int       `json:"issued_quantity"`
	RedeemedQuantity int       `json:"redeemed_quantity"`
	ValidStart       time.Time `json:"valid_start"`
	ValidEnd         time.Time `json:"valid_end"`
	LimitPerUser     int       `json:"limit_per_user"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Coupon struct {
	ID          int64     `json:"id"`
	BatchID     int64     `json:"batch_id"`
	UserID      string    `json:"user_id"`
	Code        string    `json:"code"`
	Status      string    `json:"status"`
	RedeemedAt  time.Time `json:"redeemed_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BatchProgress struct {
	BatchID          int64 `json:"batch_id"`
	TotalQuantity    int   `json:"total_quantity"`
	IssuedQuantity   int   `json:"issued_quantity"`
	RedeemedQuantity int   `json:"redeemed_quantity"`
	RemainingQuantity int  `json:"remaining_quantity"`
}

const (
	CouponStatusIssued    = "issued"
	CouponStatusRedeemed  = "redeemed"
	CouponStatusReturned  = "returned"
)

type CreateBatchRequest struct {
	Name            string    `json:"name" binding:"required"`
	DiscountAmount  int       `json:"discount_amount" binding:"required"`
	ThresholdAmount int       `json:"threshold_amount" binding:"required"`
	TotalQuantity   int       `json:"total_quantity" binding:"required"`
	ValidStart      time.Time `json:"valid_start" binding:"required"`
	ValidEnd        time.Time `json:"valid_end" binding:"required"`
	LimitPerUser    int       `json:"limit_per_user" binding:"required"`
}

type UpdateBatchRequest struct {
	Name            *string    `json:"name"`
	DiscountAmount  *int       `json:"discount_amount"`
	ThresholdAmount *int       `json:"threshold_amount"`
	ValidStart      *time.Time `json:"valid_start"`
	ValidEnd        *time.Time `json:"valid_end"`
	LimitPerUser    *int       `json:"limit_per_user"`
}

type ClaimCouponRequest struct {
	UserID  string `json:"user_id" binding:"required"`
	BatchID int64  `json:"batch_id" binding:"required"`
}

type RedeemCouponRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	CouponCode  string `json:"coupon_code" binding:"required"`
	OrderAmount int    `json:"order_amount" binding:"required"`
}

type RedeemCouponResponse struct {
	CouponID     int64     `json:"coupon_id"`
	BatchID      int64     `json:"batch_id"`
	Code         string    `json:"code"`
	RedeemedAt   time.Time `json:"redeemed_at"`
	DiscountAmt  int       `json:"discount_amount"`
}

type ReturnCouponRequest struct {
	UserID  string `json:"user_id" binding:"required"`
	BatchID int64  `json:"batch_id" binding:"required"`
}
