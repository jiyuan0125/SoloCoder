package model

import "time"

type Department string

const (
	DeptLab        Department = "检验科"
	DeptImaging    Department = "影像科"
	DeptUltrasound Department = "超声科"
	DeptECG        Department = "心电图室"
	DeptOphthalmology Department = "眼科"
	DeptENT        Department = "耳鼻喉科"
	DeptGeneral    Department = "内科"
	DeptSurgery    Department = "外科"
)

type RefRange struct {
	Min *float64
	Max *float64
}

type ExamItem struct {
	ID          string
	Name        string
	Department  Department
	RefRange    RefRange
	Unit        string
	Price       int
	Description string
}

type Package struct {
	ID          string
	Name        string
	Price       int
	Description string
	ItemIDs     []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TimeSlot string

const (
	TimeSlotMorning TimeSlot = "morning"
	TimeSlotAfternoon TimeSlot = "afternoon"
)

type AppointmentStatus string

const (
	AppointmentStatusPending   AppointmentStatus = "pending"
	AppointmentStatusCheckedIn AppointmentStatus = "checked_in"
	AppointmentStatusCompleted AppointmentStatus = "completed"
	AppointmentStatusCancelled AppointmentStatus = "cancelled"
)

type Appointment struct {
	ID           string
	ExamNumber   string
	CustomerName string
	IDCard       string
	Phone        string
	Gender       string
	Age          int
	PackageID    string
	AddItemIDs   []string
	TotalPrice   int
	ExamDate     time.Time
	TimeSlot     TimeSlot
	Status       AppointmentStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ExamResultStatus string

const (
	ResultStatusPending ExamResultStatus = "pending"
	ResultStatusCompleted ExamResultStatus = "completed"
)

type ExamResult struct {
	ID            string
	AppointmentID string
	ItemID        string
	ResultValue   *float64
	IsAbnormal    bool
	AbnormalType  string
	IsCritical    bool
	Status        ExamResultStatus
	ModifyCount   int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ModifiedAt    []time.Time
}

type CriticalAlert struct {
	ID            string
	AppointmentID string
	ExamNumber    string
	CustomerName  string
	ItemID        string
	ItemName      string
	ResultValue   float64
	RefRange      string
	CreatedAt     time.Time
	Resolved      bool
}

type ReportStatus string

const (
	ReportStatusDraft    ReportStatus = "draft"
	ReportStatusReviewing ReportStatus = "reviewing"
	ReportStatusPublished ReportStatus = "published"
)

type Report struct {
	ID              string
	AppointmentID   string
	ExamNumber      string
	CustomerName    string
	Gender          string
	Age             int
	PackageName     string
	Items           []ReportItem
	HasAbnormal     bool
	GeneralAdvice   string
	FollowUpAdvice  string
	Status          ReportStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
	PublishedAt     *time.Time
}

type ReportItem struct {
	ItemID       string
	ItemName     string
	Department   Department
	ResultValue  *float64
	Unit         string
	RefRange     string
	IsAbnormal   bool
	AbnormalType string
	IsCritical   bool
}
