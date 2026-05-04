package api

type GenerateReferralCodeRequest struct {
	UserID string `json:"user_id"`
}

type GenerateReferralCodeResponse struct {
	ReferralCode string `json:"referral_code"`
}

type BindReferralRequest struct {
	NewUserID    string `json:"new_user_id"`
	ReferralCode string `json:"referral_code"`
}

type BindReferralResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Referrer string `json:"referrer,omitempty"`
}

type CompleteFirstOrderRequest struct {
	UserID string `json:"user_id"`
	OrderID string `json:"order_id"`
}

type CompleteFirstOrderResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	PointsAwarded int64  `json:"points_awarded,omitempty"`
}

type RefundFirstOrderRequest struct {
	UserID  string `json:"user_id"`
	OrderID string `json:"order_id"`
}

type RefundFirstOrderResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	PointsDeducted int64  `json:"points_deducted,omitempty"`
}

type GetUserStatsRequest struct {
	UserID string `json:"user_id"`
}

type GetUserStatsResponse struct {
	UserID               string `json:"user_id"`
	ReferralCode         string `json:"referral_code,omitempty"`
	TotalReferrals       int64  `json:"total_referrals"`
	CompletedFirstOrders int64  `json:"completed_first_orders"`
	TotalPoints          int64  `json:"total_points"`
}

type SetRewardPointsRequest struct {
	Points int64 `json:"points"`
}

type SetRewardPointsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type GetRewardPointsResponse struct {
	Points int64 `json:"points"`
}

type GetStatsResponse struct {
	TotalReferrals      int64   `json:"total_referrals"`
	CompletedFirstOrders int64  `json:"completed_first_orders"`
	FirstOrderRate      float64 `json:"first_order_rate"`
	TotalPointsIssued   int64   `json:"total_points_issued"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
