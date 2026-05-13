package models

import (
	"time"
)

type MemberLevel string

const (
	LevelNormal MemberLevel = "normal"
	LevelSilver MemberLevel = "silver"
	LevelGold   MemberLevel = "gold"
	LevelDiamond MemberLevel = "diamond"
)

type Member struct {
	ID            int64       `json:"id"`
	Name          string      `json:"name"`
	Level         MemberLevel `json:"level"`
	Points        int64       `json:"points"`
	YearlyPoints  int64       `json:"yearly_points"`
	Year          int         `json:"year"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

type Consumption struct {
	ID           int64     `json:"id"`
	MemberID     int64     `json:"member_id"`
	Amount       int64     `json:"amount"`
	PointsEarned int64     `json:"points_earned"`
	Multiplier   float64   `json:"multiplier"`
	Processed    bool      `json:"processed"`
	CreatedAt    time.Time `json:"created_at"`
}

type Product struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	PointsCost  int64  `json:"points_cost"`
	Description string `json:"description"`
}

type Exchange struct {
	ID          int64     `json:"id"`
	MemberID    int64     `json:"member_id"`
	ProductID   int64     `json:"product_id"`
	PointsUsed  int64     `json:"points_used"`
	Processed   bool      `json:"processed"`
	CreatedAt   time.Time `json:"created_at"`
}

type YearEndRecord struct {
	ID           int64       `json:"id"`
	MemberID     int64       `json:"member_id"`
	Year         int         `json:"year"`
	TotalPoints  int64       `json:"total_points"`
	OldLevel     MemberLevel `json:"old_level"`
	NewLevel     MemberLevel `json:"new_level"`
	Processed    bool        `json:"processed"`
	CreatedAt    time.Time   `json:"created_at"`
}
