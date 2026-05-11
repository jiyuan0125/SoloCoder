package common

import "time"

type GasStation struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Phone       string    `json:"phone"`
	IsOpen      bool      `json:"is_open"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type FuelType struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Price    int64  `json:"price"`
	Stock    int64  `json:"stock"`
}

type StationFuel struct {
	StationID string    `json:"station_id"`
	FuelCode  string    `json:"fuel_code"`
	Price     int64     `json:"price"`
	Stock     int64     `json:"stock"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MemberLevel string

const (
	LevelNormal MemberLevel = "normal"
	LevelSilver MemberLevel = "silver"
	LevelGold   MemberLevel = "gold"
)

const (
	UpgradeSilverAmount int64 = 100000
	UpgradeGoldAmount   int64 = 500000
)

type Member struct {
	ID               string      `json:"id"`
	Phone            string      `json:"phone"`
	Level            MemberLevel `json:"level"`
	Points           int64       `json:"points"`
	TotalSpent       int64       `json:"total_spent"`
	RegisteredAt     time.Time   `json:"registered_at"`
	LastLevelUpAt    *time.Time  `json:"last_level_up_at,omitempty"`
}

type RefuelRecord struct {
	ID              string    `json:"id"`
	StationID       string    `json:"station_id"`
	FuelCode        string    `json:"fuel_code"`
	Liters          float64   `json:"liters"`
	UnitPrice       int64     `json:"unit_price"`
	OriginalAmount  int64     `json:"original_amount"`
	DiscountAmount  int64     `json:"discount_amount"`
	PointsUsed      int64     `json:"points_used"`
	PointsDeducted  int64     `json:"points_deducted"`
	FinalAmount     int64     `json:"final_amount"`
	MemberID        *string   `json:"member_id,omitempty"`
	MemberLevel     *string   `json:"member_level,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type PriceApprovalStatus string

const (
	ApprovalStatusPending PriceApprovalStatus = "pending"
	ApprovalStatusApproved PriceApprovalStatus = "approved"
	ApprovalStatusRejected PriceApprovalStatus = "rejected"
)

type PriceChangeRequest struct {
	ID              string               `json:"id"`
	StationID       string               `json:"station_id"`
	FuelCode        string               `json:"fuel_code"`
	OldPrice        int64                `json:"old_price"`
	NewPrice        int64                `json:"new_price"`
	Status          PriceApprovalStatus  `json:"status"`
	SubmittedBy     string               `json:"submitted_by"`
	SubmittedAt     time.Time            `json:"submitted_at"`
	ApprovedBy      *string              `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time           `json:"approved_at,omitempty"`
}

type RestockTodo struct {
	ID          string    `json:"id"`
	StationID   string    `json:"station_id"`
	FuelCode    string    `json:"fuel_code"`
	CurrentStock int64    `json:"current_stock"`
	Threshold   int64     `json:"threshold"`
	IsCompleted bool      `json:"is_completed"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type CreateStationRequest struct {
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Phone     string  `json:"phone"`
	IsOpen    bool    `json:"is_open"`
}

type UpdateStationRequest struct {
	Name      *string  `json:"name,omitempty"`
	Address   *string  `json:"address,omitempty"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
	Phone     *string  `json:"phone,omitempty"`
	IsOpen    *bool    `json:"is_open,omitempty"`
}

type RegisterMemberRequest struct {
	Phone string `json:"phone"`
}

type RefuelRequest struct {
	StationID   string  `json:"station_id"`
	FuelCode    string  `json:"fuel_code"`
	Liters      float64 `json:"liters"`
	MemberPhone *string `json:"member_phone,omitempty"`
	UsePoints   *int64  `json:"use_points,omitempty"`
}

type SubmitPriceChangeRequest struct {
	StationID string `json:"station_id"`
	FuelCode  string `json:"fuel_code"`
	NewPrice  int64  `json:"new_price"`
}

type ReviewPriceChangeRequest struct {
	Approved bool   `json:"approved"`
	ReviewedBy string `json:"reviewed_by"`
}

type QueryRecordsRequest struct {
	StationID string    `json:"station_id,omitempty"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

type ListResponse struct {
	Items []interface{} `json:"items"`
	Total int           `json:"total"`
}
