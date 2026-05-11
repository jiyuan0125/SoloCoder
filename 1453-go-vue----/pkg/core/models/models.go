package models

import "time"

type RepairType string

const (
	RepairTypeWaterElectric RepairType = "water_electric"
	RepairTypeDoorWindow    RepairType = "door_window"
	RepairTypePublicFacility RepairType = "public_facility"
	RepairTypeOther         RepairType = "other"
)

type RepairStatus string

const (
	RepairStatusPendingDispatch RepairStatus = "pending_dispatch"
	RepairStatusAssigned        RepairStatus = "assigned"
	RepairStatusProcessing      RepairStatus = "processing"
	RepairStatusCompleted       RepairStatus = "completed"
	RepairStatusClosed          RepairStatus = "closed"
	RepairStatusEscalated       RepairStatus = "escalated"
)

type Role string

const (
	RoleOwner     Role = "owner"
	RoleStaff     Role = "staff"
	RoleSupervisor Role = "supervisor"
	RoleAdmin     Role = "admin"
)

type BillType string

const (
	BillTypeProperty BillType = "property"
	BillTypeWater    BillType = "water"
	BillTypeElectric BillType = "electric"
	BillTypeParking  BillType = "parking"
	BillTypeRepair   BillType = "repair"
)

type BillStatus string

const (
	BillStatusPending  BillStatus = "pending"
	BillStatusPartial  BillStatus = "partial"
	BillStatusPaid     BillStatus = "paid"
)

type AnnouncementScope string

const (
	AnnouncementScopeAll AnnouncementScope = "all"
	AnnouncementScopeBuilding AnnouncementScope = "building"
)

type User struct {
	ID           string
	Username     string
	Name         string
	Role         Role
	Building     string
	Unit         string
	Phone        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Repair struct {
	ID               string
	OwnerID          string
	OwnerName        string
	Building         string
	Unit             string
	RepairType       RepairType
	Description      string
	ExpectedTime     time.Time
	Status           RepairStatus
	AssignedStaffID  string
	AssignedStaffName string
	Result           string
	ResultImageURLs  []string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	AssignedAt       time.Time
	CompletedAt      time.Time
	ConfirmedAt      time.Time
	ClosedAt         time.Time
	IsAutoConfirmed  bool
}

type Bill struct {
	ID              string
	UserID          string
	UserName        string
	Building        string
	Unit            string
	BillType        BillType
	Amount          int64
	PaidAmount      int64
	UnpaidAmount    int64
	DueDate         time.Time
	Status          BillStatus
	RelatedRepairID string
	PenaltyAmount   int64
	Month           int
	Year            int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Payment struct {
	ID              string
	BillID          string
	UserID          string
	Amount          int64
	PaidAt          time.Time
	PaymentMethod   string
	TransactionNo   string
}

type Announcement struct {
	ID              string
	Title           string
	Content         string
	Scope           AnnouncementScope
	TargetBuilding  string
	EffectiveTime   time.Time
	ExpiryTime      time.Time
	CreatedBy       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
