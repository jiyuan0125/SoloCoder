package core

import (
	"time"
)

type GasOutlet struct {
	ID          string
	Location    string
	Status      OutletStatus
	LastReport  *time.Time
	ContinuousExceedingStart *time.Time
}

type OutletStatus int

const (
	OutletStatusNormal OutletStatus = iota
	OutletStatusWarning
	OutletStatusEmergency
)

type GasReport struct {
	ID          string
	OutletID    string
	ReportedAt  time.Time
	ReceivedAt  time.Time
	DataStatus  DataStatus
	Measurements map[string]Measurement
	ExceededFactors []string
	AbnormalFactors []string
}

type Measurement struct {
	Value       float64
	Factor      string
	Unit        string
	IsValid     bool
	IsAbnormal  bool
	IsExceeded  bool
}

type DataStatus int

const (
	DataStatusNormal DataStatus = iota
	DataStatusPartialAbnormal
	DataStatusAllAbnormal
)

type GasStandard struct {
	Factor     string
	Limit      float64
	Unit       string
	MaxValid   float64
}

type WastewaterReport struct {
	ID          string
	OutletID    string
	ReportedAt  time.Time
	ReceivedAt  time.Time
	DataStatus  DataStatus
	Measurements map[string]Measurement
	ExceededFactors []string
	AbnormalFactors []string
}

type WastewaterStandard struct {
	Factor     string
	Limit      float64
	Unit       string
	MaxValid   float64
	MinValid   float64
}

type DailyReport struct {
	Date        time.Time
	OutletID    string
	Status      ReportStatus
	Data        map[string]DailyFactorData
}

type DailyFactorData struct {
	Max      float64
	Min      float64
	Average  float64
	Count    int
	Unit     string
	Exceeded bool
}

type ReportStatus int

const (
	ReportStatusNormal ReportStatus = iota
	ReportStatusExceeded
)

type SolidWasteRecord struct {
	ID             string
	Name           string
	Category       WasteCategory
	Amount         float64
	StorageLocation string
	DisposalMethod string
	GeneratedAt    time.Time
	DisposedAt     *time.Time
	Status         WasteStatus
}

type WasteCategory int

const (
	WasteCategoryHazardous WasteCategory = iota
	WasteCategoryGeneral
)

type WasteStatus int

const (
	WasteStatusStored WasteStatus = iota
	WasteStatusDisposed
	WasteStatusExpired
)

type Alarm struct {
	ID              string
	Type            AlarmType
	Level           AlarmLevel
	EntityType      EntityType
	EntityID        string
	RelatedData     map[string]interface{}
	Message         string
	CreatedAt       time.Time
	ResolvedAt      *time.Time
	IsResolved      bool
}

type AlarmType int

const (
	AlarmTypeExceeding AlarmType = iota
	AlarmTypeEquipmentFailure
	AlarmTypeStorageExpired
	AlarmTypeEmergency
)

type AlarmLevel int

const (
	AlarmLevelWarning AlarmLevel = iota
	AlarmLevelEmergency
)

type EntityType int

const (
	EntityTypeGasOutlet EntityType = iota
	EntityTypeWastewaterOutlet
	EntityTypeSolidWaste
	EntityTypeDailyReport
)
