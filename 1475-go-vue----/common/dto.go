package common

import "time"

type ParkingSpotType string

const (
	SpotTypeNormal      ParkingSpotType = "normal"
	SpotTypeCharging    ParkingSpotType = "charging"
	SpotTypeAccessible  ParkingSpotType = "accessible"
)

type ParkingSpotStatus string

const (
	SpotStatusAvailable  ParkingSpotStatus = "available"
	SpotStatusOccupied   ParkingSpotStatus = "occupied"
	SpotStatusReserved   ParkingSpotStatus = "reserved"
	SpotStatusMaintenance ParkingSpotStatus = "maintenance"
)

type MonthlyCardType string

const (
	CardTypeNormal        MonthlyCardType = "normal"
	CardTypeFixedSpot     MonthlyCardType = "fixed_spot"
	CardTypeCharging      MonthlyCardType = "charging"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ParkingSpotDTO struct {
	ID       string             `json:"id"`
	Area     string             `json:"area"`
	Number   string             `json:"number"`
	Type     ParkingSpotType    `json:"type"`
	Status   ParkingSpotStatus  `json:"status"`
}

type CreateParkingSpotRequest struct {
	ID       string           `json:"id"`
	Area     string           `json:"area"`
	Number   string           `json:"number"`
	Type     ParkingSpotType  `json:"type"`
}

type UpdateParkingSpotStatusRequest struct {
	Status ParkingSpotStatus `json:"status"`
}

type ParkingGuidanceDTO struct {
	TotalSpots        int            `json:"total_spots"`
	AvailableSpots    int            `json:"available_spots"`
	AvailablePercent  float64        `json:"available_percent"`
	IsFull            bool           `json:"is_full"`
	IsTight           bool           `json:"is_tight"`
	Areas             []AreaInfo     `json:"areas"`
}

type AreaInfo struct {
	Name            string `json:"name"`
	TotalSpots      int    `json:"total_spots"`
	AvailableSpots  int    `json:"available_spots"`
	IsFull          bool   `json:"is_full"`
}

type CheckInRequest struct {
	PlateNumber string `json:"plate_number"`
}

type CheckInResponse struct {
	PlateNumber   string    `json:"plate_number"`
	CheckInTime   time.Time `json:"check_in_time"`
	SpotID        string    `json:"spot_id"`
	IsMonthlyCard bool      `json:"is_monthly_card"`
}

type CheckOutRequest struct {
	PlateNumber string `json:"plate_number"`
}

type CheckOutResponse struct {
	PlateNumber string  `json:"plate_number"`
	CheckInTime time.Time `json:"check_in_time"`
	CheckOutTime time.Time `json:"check_out_time"`
	Duration    float64 `json:"duration_minutes"`
	Amount      float64 `json:"amount"`
	IsMonthlyCard bool    `json:"is_monthly_card"`
	FreeTimeUsed  float64 `json:"free_time_used_minutes,omitempty"`
}

type CreateMonthlyCardRequest struct {
	CardType   MonthlyCardType `json:"card_type"`
	OwnerName  string          `json:"owner_name"`
	PlateNumber string         `json:"plate_number"`
	SpotID     string          `json:"spot_id,omitempty"`
}

type MonthlyCardDTO struct {
	ID           string          `json:"id"`
	CardType     MonthlyCardType `json:"card_type"`
	OwnerName    string          `json:"owner_name"`
	PlateNumber  string          `json:"plate_number"`
	SpotID       string          `json:"spot_id,omitempty"`
	StartDate    time.Time       `json:"start_date"`
	EndDate      time.Time       `json:"end_date"`
	GraceEndDate time.Time       `json:"grace_end_date"`
	IsActive     bool            `json:"is_active"`
	Status       string          `json:"status"`
}

type RenewMonthlyCardRequest struct {
	CardID string `json:"card_id"`
}

type ParkingRecordDTO struct {
	PlateNumber  string    `json:"plate_number"`
	CheckInTime  time.Time `json:"check_in_time"`
	CheckOutTime time.Time `json:"check_out_time,omitempty"`
	SpotID       string    `json:"spot_id"`
	Amount       float64   `json:"amount,omitempty"`
	IsActive     bool      `json:"is_active"`
}

type QueryFeeRequest struct {
	PlateNumber string `json:"plate_number"`
}

type QueryFeeResponse struct {
	PlateNumber    string    `json:"plate_number"`
	CheckInTime    time.Time `json:"check_in_time"`
	CurrentTime    time.Time `json:"current_time"`
	Duration       float64   `json:"duration_minutes"`
	EstimatedAmount float64  `json:"estimated_amount"`
	IsMonthlyCard  bool      `json:"is_monthly_card"`
}

type ListParkingSpotsRequest struct {
	Area    string `json:"area,omitempty"`
	Status  string `json:"status,omitempty"`
}

type ListParkingSpotsResponse struct {
	Spots []ParkingSpotDTO `json:"spots"`
}

type ListParkingRecordsRequest struct {
	PlateNumber string `json:"plate_number,omitempty"`
	ActiveOnly  bool   `json:"active_only"`
}

type ListParkingRecordsResponse struct {
	Records []ParkingRecordDTO `json:"records"`
}

type ListMonthlyCardsRequest struct {
	PlateNumber string `json:"plate_number,omitempty"`
	ActiveOnly  bool   `json:"active_only"`
}

type ListMonthlyCardsResponse struct {
	Cards []MonthlyCardDTO `json:"cards"`
}
