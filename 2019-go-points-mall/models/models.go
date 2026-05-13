package models

import "time"

const (
	StatusNew        = "new"
	StatusPendingPay = "pending_pay"
	StatusRedeemed   = "redeemed"
	StatusUsed       = "used"
)

type Product struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Points     int       `json:"points"`
	Stock      int       `json:"stock"`
	DailyLimit int       `json:"daily_limit"`
	IsOnline   bool      `json:"is_online"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type User struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	PointsBalance int       `json:"points_balance"`
	CreatedAt     time.Time `json:"created_at"`
}

type Exchange struct {
	ID            int        `json:"id"`
	UserID        int        `json:"user_id"`
	ProductID     int        `json:"product_id"`
	Status        string     `json:"status"`
	Points        int        `json:"points"`
	LockExpiresAt *time.Time `json:"lock_expires_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type ExchangeDetail struct {
	ID          int       `json:"id"`
	ExchangeID  int       `json:"exchange_id"`
	ActionType  string    `json:"action_type"`
	OldStatus   *string   `json:"old_status,omitempty"`
	NewStatus   string    `json:"new_status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type RedemptionCode struct {
	ID         int        `json:"id"`
	Code       string     `json:"code"`
	ExchangeID int        `json:"exchange_id"`
	ProductID  int        `json:"product_id"`
	UserID     int        `json:"user_id"`
	IsUsed     bool       `json:"is_used"`
	ExpiresAt  time.Time  `json:"expires_at"`
	RedeemedAt *time.Time `json:"redeemed_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}
