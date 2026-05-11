package common

import (
	"time"
)

type TenderMethod string

const (
	PublicTender  TenderMethod = "public"
	InviteTender  TenderMethod = "invite"
)

type EvaluationMethod string

const (
	LowestPriceMethod EvaluationMethod = "lowest_price"
	ComprehensiveScoreMethod EvaluationMethod = "comprehensive_score"
)

type ProjectStatus string

const (
	ProjectDraft ProjectStatus = "draft"
	ProjectPublished ProjectStatus = "published"
	ProjectOpened ProjectStatus = "opened"
)

type Project struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	Description      string           `json:"description"`
	Method           TenderMethod     `json:"method"`
	BidDeadline      time.Time        `json:"bid_deadline"`
	OpenTime         time.Time        `json:"open_time"`
	EvaluationMethod EvaluationMethod `json:"evaluation_method"`
	InvitedSuppliers []string         `json:"invited_suppliers,omitempty"`
	Status           ProjectStatus    `json:"status"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

type Bid struct {
	ID                string    `json:"id"`
	ProjectID         string    `json:"project_id"`
	SupplierID        string    `json:"supplier_id"`
	Amount            int64     `json:"amount"`
	TechnicalPlan     string    `json:"technical_plan"`
	TechnicalScore    float64   `json:"technical_score"`
	DurationDays      int       `json:"duration_days"`
	ComprehensiveScore float64  `json:"comprehensive_score,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type OpeningResult struct {
	ProjectID    string    `json:"project_id"`
	WinnerBidID  string    `json:"winner_bid_id"`
	WinnerSupplierID string `json:"winner_supplier_id"`
	WinnerAmount int64     `json:"winner_amount"`
	OpenedAt     time.Time `json:"opened_at"`
}
