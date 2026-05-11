package lockservice

import "time"

type LockType string

const (
	LockTypeSecurity  LockType = "security"
	LockTypeInterior  LockType = "interior"
	LockTypePassword  LockType = "password"
	LockTypeFingerprint LockType = "fingerprint"
	LockTypeCar       LockType = "car"
)

type Urgency string

const (
	UrgencyNormal  Urgency = "normal"
	UrgencyUrgent  Urgency = "urgent"
)

type TimeSlot string

const (
	TimeSlotMorning   TimeSlot = "morning"
	TimeSlotAfternoon TimeSlot = "afternoon"
	TimeSlotEvening   TimeSlot = "evening"
)

type MasterStatus string

const (
	MasterStatusIdle    MasterStatus = "idle"
	MasterStatusEnRoute MasterStatus = "en_route"
	MasterStatusBusy    MasterStatus = "busy"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusAccepted   OrderStatus = "accepted"
	OrderStatusInService  OrderStatus = "in_service"
	OrderStatusPendingPayment OrderStatus = "pending_payment"
	OrderStatusCompleted  OrderStatus = "completed"
)

type Master struct {
	ID              string
	Name            string
	Phone           string
	ServiceArea     string
	SupportedLocks  []LockType
	Status          MasterStatus
	Location        string
	BookedSlots     map[string]map[TimeSlot]bool
}

type OrderDetail struct {
	NeedReplaceLock bool
	LockBrand       string
	LockModel       string
	LockLevel       string
	PartsCost       float64
	LaborCost       float64
}

type Order struct {
	ID             string
	UserID         string
	Address        string
	LockType       LockType
	Urgency        Urgency
	TimeSlot       TimeSlot
	TimeSlotDate   time.Time
	MasterID       string
	Status         OrderStatus
	Detail         OrderDetail
	BaseCost       float64
	UrgentFee      float64
	TotalCost      float64
	CreatedAt      time.Time
	AcceptedAt     time.Time
	InServiceAt    time.Time
	PendingPayAt   time.Time
	CompletedAt    time.Time
	Rating         int
	Comment        string
	HasComplaint   bool
}

type DispatchResult struct {
	Success     bool
	Master      *Master
	Alternative *AlternativeSlots
}

type AlternativeSlots struct {
	OtherSlots   []TimeSlot
	OtherMasters []*Master
}

type Pricing struct {
	BaseFee       float64
	InstallationFee float64
	UrgentRate    float64
}
