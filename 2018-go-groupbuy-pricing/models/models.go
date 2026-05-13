package models

import "time"

type User struct {
	UserID string `json:"user_id"`
	Balance int64  `json:"balance"`
}

type Tier struct {
	MinPeople int     `json:"min_people"`
	Discount  float64 `json:"discount"`
}

type Activity struct {
	ActivityID   string    `json:"activity_id"`
	Name         string    `json:"name"`
	BasePrice    int64     `json:"base_price"`
	Tiers        []Tier    `json:"tiers"`
	Deadline     time.Time `json:"deadline"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	SettledAt    *time.Time `json:"settled_at,omitempty"`
	CurrentTier  int       `json:"current_tier"`
	ParticipantCount int   `json:"participant_count"`
}

type Order struct {
	OrderID     string    `json:"order_id"`
	ActivityID  string    `json:"activity_id"`
	UserID      string    `json:"user_id"`
	PaidPrice   int64     `json:"paid_price"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type RefundRecord struct {
	RefundID   string    `json:"refund_id"`
	UserID     string    `json:"user_id"`
	ActivityID string    `json:"activity_id"`
	OrderID    string    `json:"order_id"`
	Amount     int64     `json:"amount"`
	Reason     string    `json:"reason"`
	CreatedAt  time.Time `json:"created_at"`
}

type WithdrawalRecord struct {
	WithdrawalID string    `json:"withdrawal_id"`
	UserID       string    `json:"user_id"`
	Amount       int64     `json:"amount"`
	Fee          int64     `json:"fee"`
	NetAmount    int64     `json:"net_amount"`
	CreatedAt    time.Time `json:"created_at"`
}
