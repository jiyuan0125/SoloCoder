package main

import "time"

type User struct {
	ID         string
	ReferralCode string
}

type ReferralRelation struct {
	ReferrerID   string
	RefereeID    string
	BindTime     time.Time
	FirstOrderID string
	FirstOrderPaid bool
	RewardPoints int64
}

type PointsTransaction struct {
	ID          string
	UserID      string
	Amount      int64
	Type        PointsTransactionType
	Description string
	RefOrderID  string
	CreatedAt   time.Time
}

type PointsTransactionType string

const (
	TransactionTypeReward    PointsTransactionType = "reward"
	TransactionTypeRefund    PointsTransactionType = "refund"
	TransactionTypeAdminAdjust PointsTransactionType = "admin_adjust"
)

type SystemConfig struct {
	RewardPointsPerFirstOrder int64
}

type UserStats struct {
	UserID           string
	TotalInvited     int
	CompletedFirstOrder int
	TotalPointsEarned int64
}

type AdminStats struct {
	TotalReferrals        int
	FirstOrderCompletionRate float64
	TotalPointsDistributed int64
}
