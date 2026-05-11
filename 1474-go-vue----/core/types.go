package core

import "time"

type FenceType string

const (
	FenceTypeRecommended FenceType = "recommended"
	FenceTypeForbidden   FenceType = "forbidden"
	FenceTypeDispatch    FenceType = "dispatch"
)

type Fence struct {
	ID           string
	Name         string
	Type         FenceType
	MinLat       float64
	MaxLat       float64
	MinLng       float64
	MaxLng       float64
	Capacity     int
	CreatedAt    time.Time
}

type BikeStatus string

const (
	BikeStatusIdle      BikeStatus = "idle"
	BikeStatusInUse     BikeStatus = "in_use"
	BikeStatusRepair    BikeStatus = "repair"
	BikeStatusScrapped  BikeStatus = "scrapped"
)

type Bike struct {
	ID            string
	Status        BikeStatus
	Lat           float64
	Lng           float64
	FenceID       string
	LastUsedAt    time.Time
	CreatedAt     time.Time
}

type Ride struct {
	ID         string
	BikeID     string
	StartTime  time.Time
	EndTime    *time.Time
	StartLat   float64
	StartLng   float64
	EndLat     *float64
	EndLng     *float64
	BaseFee    float64
	DispatchFee float64
	TotalFee   float64
}

type DispatchTask struct {
	ID          string
	FromFenceID string
	ToFenceID   string
	BikeIDs     []string
	Reason      string
	CreatedAt   time.Time
	CompletedAt *time.Time
}

type Point struct {
	Lat float64
	Lng float64
}
