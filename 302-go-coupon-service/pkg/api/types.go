package api

import "time"

type CouponBatch struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Discount      float64   `json:"discount"`
	Threshold     float64   `json:"threshold"`
	TotalQuantity int       `json:"total_quantity"`
	Claimed       int       `json:"claimed"`
	Redeemed      int       `json:"redeemed"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	PerUserLimit  int       `json:"per_user_limit"`
	IsAllClaimed  bool      `json:"is_all_claimed"`
}

type CouponClaim struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	BatchID     string    `json:"batch_id"`
	ClaimTime   time.Time `json:"claim_time"`
	IsRedeemed  bool      `json:"is_redeemed"`
	RedeemTime  time.Time `json:"redeem_time"`
}

type CouponRedemption struct {
	ID            string    `json:"id"`
	ClaimID       string    `json:"claim_id"`
	BatchID       string    `json:"batch_id"`
	UserID        string    `json:"user_id"`
	OrderAmount   float64   `json:"order_amount"`
	RedeemTime    time.Time `json:"redeem_time"`
	DiscountUsed  float64   `json:"discount_used"`
}

type CreateBatchRequest struct {
	Name          string  `json:"name"`
	Discount      float64 `json:"discount"`
	Threshold     float64 `json:"threshold"`
	TotalQuantity int     `json:"total_quantity"`
	StartTime     string  `json:"start_time"`
	EndTime       string  `json:"end_time"`
	PerUserLimit  int     `json:"per_user_limit"`
}

type CreateBatchResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Batch   CouponBatch `json:"batch,omitempty"`
}

type ClaimCouponRequest struct {
	UserID  string `json:"user_id"`
	BatchID string `json:"batch_id"`
}

type ClaimCouponResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Claim   CouponClaim `json:"claim,omitempty"`
}

type RedeemCouponRequest struct {
	UserID      string  `json:"user_id"`
	ClaimID     string  `json:"claim_id"`
	OrderAmount float64 `json:"order_amount"`
}

type RedeemCouponResponse struct {
	Success    bool             `json:"success"`
	Message    string           `json:"message"`
	Redemption CouponRedemption `json:"redemption,omitempty"`
}

type GetBatchStatsRequest struct {
	BatchID string `json:"batch_id"`
}

type GetBatchStatsResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Batch     CouponBatch `json:"batch,omitempty"`
	Claimed   int         `json:"claimed"`
	Redeemed  int         `json:"redeemed"`
	Remaining int         `json:"remaining"`
}

type GetUserCouponsRequest struct {
	UserID string `json:"user_id"`
}

type GetUserCouponsResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Coupons []CouponClaim `json:"coupons,omitempty"`
}

type ReturnCouponRequest struct {
	UserID  string `json:"user_id"`
	ClaimID string `json:"claim_id"`
}

type ReturnCouponResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type UpdateBatchRequest struct {
	BatchID       string  `json:"batch_id"`
	Name          string  `json:"name"`
	Discount      float64 `json:"discount"`
	Threshold     float64 `json:"threshold"`
	StartTime     string  `json:"start_time"`
	EndTime       string  `json:"end_time"`
	PerUserLimit  int     `json:"per_user_limit"`
}

type UpdateBatchResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Batch   CouponBatch `json:"batch,omitempty"`
}
