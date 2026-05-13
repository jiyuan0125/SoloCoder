package models

import "time"

type AssetStatus string

const (
	StatusDraft     AssetStatus = "draft"
	StatusPending   AssetStatus = "pending"
	StatusApproved  AssetStatus = "approved"
	StatusRunning   AssetStatus = "running"
	StatusCompleted AssetStatus = "completed"
	StatusScrapped  AssetStatus = "scrapped"
)

type DepreciationMethod string

const (
	MethodStraightLine        DepreciationMethod = "straight_line"
	MethodDoubleDeclining     DepreciationMethod = "double_declining"
)

type Asset struct {
	ID                 int64              `json:"id"`
	Name               string             `json:"name"`
	OriginalValue      int64              `json:"original_value"`
	SalvageRate        float64            `json:"salvage_rate"`
	UsefulLifeMonths   int                `json:"useful_life_months"`
	PurchaseDate       string             `json:"purchase_date"`
	PurchaseDay        int                `json:"purchase_day"`
	Department         string             `json:"department"`
	Status             AssetStatus        `json:"status"`
	Method             DepreciationMethod `json:"method"`
	SalvageValue       int64              `json:"salvage_value"`
	AccumulatedDepr    int64              `json:"accumulated_depr"`
	CurrentBookValue   int64              `json:"current_book_value"`
	LastDeprMonth      string             `json:"last_depr_month"`
	ScrappedDate       *string            `json:"scrapped_date,omitempty"`
	ScrappedValue      *int64             `json:"scrapped_value,omitempty"`
	DisposalGainLoss   *int64             `json:"disposal_gain_loss,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

type OperationHistory struct {
	ID          int64       `json:"id"`
	AssetID     int64       `json:"asset_id"`
	FromStatus  AssetStatus `json:"from_status"`
	ToStatus    AssetStatus `json:"to_status"`
	Action      string      `json:"action"`
	Remark      string      `json:"remark"`
	CreatedAt   time.Time   `json:"created_at"`
}

type DepreciationRecord struct {
	ID              int64     `json:"id"`
	AssetID         int64     `json:"asset_id"`
	DeprMonth       string    `json:"depr_month"`
	DeprAmount      int64     `json:"depr_amount"`
	AccumulatedDepr int64     `json:"accumulated_depr"`
	BookValue       int64     `json:"book_value"`
	IsPartialMonth  bool      `json:"is_partial_month"`
	DaysInMonth     int       `json:"days_in_month"`
	CreatedAt       time.Time `json:"created_at"`
}

type DepartmentSummary struct {
	Department        string `json:"department"`
	CurrentDeprTotal  int64  `json:"current_depr_total"`
	AssetCount        int    `json:"asset_count"`
}

type CreateAssetRequest struct {
	Name             string             `json:"name"`
	OriginalValue    int64              `json:"original_value"`
	SalvageRate      float64            `json:"salvage_rate"`
	UsefulLifeYears  int                `json:"useful_life_years"`
	PurchaseDate     string             `json:"purchase_date"`
	Department       string             `json:"department"`
	Method           DepreciationMethod `json:"method"`
}

type UpdateStatusRequest struct {
	Action string `json:"action"`
	Remark string `json:"remark"`
}

type ScrapAssetRequest struct {
	ScrappedValue int64  `json:"scrapped_value"`
	Remark        string `json:"remark"`
}
