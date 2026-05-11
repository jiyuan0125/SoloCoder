package api

import "time"

type DeviceType string

const (
	DeviceTypeFireExtinguisher DeviceType = "fire_extinguisher"
	DeviceTypeFireHydrant      DeviceType = "fire_hydrant"
	DeviceTypeSmokeDetector    DeviceType = "smoke_detector"
	DeviceTypeSprinklerHead    DeviceType = "sprinkler_head"
	DeviceTypeEmergencyLight   DeviceType = "emergency_light"
)

func (t DeviceType) Valid() bool {
	switch t {
	case DeviceTypeFireExtinguisher, DeviceTypeFireHydrant, DeviceTypeSmokeDetector, DeviceTypeSprinklerHead, DeviceTypeEmergencyLight:
		return true
	default:
		return false
	}
}

func (t DeviceType) String() string {
	switch t {
	case DeviceTypeFireExtinguisher:
		return "灭火器"
	case DeviceTypeFireHydrant:
		return "消火栓"
	case DeviceTypeSmokeDetector:
		return "烟感探测器"
	case DeviceTypeSprinklerHead:
		return "喷淋头"
	case DeviceTypeEmergencyLight:
		return "应急灯"
	default:
		return string(t)
	}
}

type DeviceStatus string

const (
	DeviceStatusNormal    DeviceStatus = "normal"
	DeviceStatusPending   DeviceStatus = "pending_repair"
	DeviceStatusScrapped  DeviceStatus = "scrapped"
)

func (s DeviceStatus) Valid() bool {
	switch s {
	case DeviceStatusNormal, DeviceStatusPending, DeviceStatusScrapped:
		return true
	default:
		return false
	}
}

type InspectionFrequency string

const (
	InspectionFrequencyDaily   InspectionFrequency = "daily"
	InspectionFrequencyWeekly  InspectionFrequency = "weekly"
	InspectionFrequencyMonthly InspectionFrequency = "monthly"
)

func (f InspectionFrequency) Valid() bool {
	switch f {
	case InspectionFrequencyDaily, InspectionFrequencyWeekly, InspectionFrequencyMonthly:
		return true
	default:
		return false
	}
}

type DrillType string

const (
	DrillTypeFireExtinguisherUse DrillType = "fire_extinguisher_use"
	DrillTypeEvacuationEscape    DrillType = "evacuation_escape"
	DrillTypeComprehensive       DrillType = "comprehensive"
)

func (t DrillType) Valid() bool {
	switch t {
	case DrillTypeFireExtinguisherUse, DrillTypeEvacuationEscape, DrillTypeComprehensive:
		return true
	default:
		return false
	}
}

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

type ReminderType string

const (
	ReminderTypeDeviceExpiry ReminderType = "device_expiry"
	ReminderTypeQuarterlyDrill ReminderType = "quarterly_drill"
)

type Device struct {
	ID            string       `json:"id"`
	Code          string       `json:"code"`
	Type          DeviceType   `json:"type"`
	Location      string       `json:"location"`
	InstallDate   time.Time    `json:"install_date"`
	ExpiryDate    time.Time    `json:"expiry_date"`
	LastCheckDate time.Time    `json:"last_check_date"`
	Status        DeviceStatus `json:"status"`
}

type InspectionPoint struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

type InspectionRoute struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Points []InspectionPoint `json:"points"`
}

type InspectionPlan struct {
	ID           string              `json:"id"`
	RouteID      string              `json:"route_id"`
	Route        *InspectionRoute    `json:"route,omitempty"`
	InspectorID  string              `json:"inspector_id"`
	InspectorName string             `json:"inspector_name"`
	Frequency    InspectionFrequency `json:"frequency"`
	StartTime    time.Time           `json:"start_time"`
}

type InspectionTask struct {
	ID          string       `json:"id"`
	PlanID      string       `json:"plan_id"`
	Plan        *InspectionPlan `json:"plan,omitempty"`
	RouteID     string       `json:"route_id"`
	Route       *InspectionRoute `json:"route,omitempty"`
	InspectorID string       `json:"inspector_id"`
	InspectorName string      `json:"inspector_name"`
	Status      TaskStatus   `json:"status"`
	CreatedAt   time.Time    `json:"created_at"`
	CompletedAt *time.Time   `json:"completed_at"`
	Points      []TaskPoint  `json:"points"`
}

type TaskPoint struct {
	PointID    string     `json:"point_id"`
	Point      *InspectionPoint `json:"point,omitempty"`
	OrderIndex int        `json:"order_index"`
	Checked    bool       `json:"checked"`
	Normal     *bool      `json:"normal"`
	CheckedAt  *time.Time `json:"checked_at"`
}

type DrillPlan struct {
	ID              string    `json:"id"`
	Type            DrillType `json:"type"`
	Name            string    `json:"name"`
	ScheduledTime   time.Time `json:"scheduled_time"`
	PlannedAttendees int       `json:"planned_attendees"`
	CreatedAt       time.Time `json:"created_at"`
}

type DrillRecord struct {
	ID                string     `json:"id"`
	PlanID            string     `json:"plan_id"`
	Plan              *DrillPlan `json:"plan,omitempty"`
	ActualAttendees   int        `json:"actual_attendees"`
	DurationMinutes   int        `json:"duration_minutes"`
	Improvements      string     `json:"improvements"`
	CompletedAt       time.Time  `json:"completed_at"`
}

type Reminder struct {
	ID        string       `json:"id"`
	Type      ReminderType `json:"type"`
	Title     string       `json:"title"`
	Content   string       `json:"content"`
	ReferenceID string     `json:"reference_id"`
	CreatedAt time.Time    `json:"created_at"`
	Read      bool         `json:"read"`
}

type ScheduledTask struct {
	ID           string
	Name         string
	Cron         string
	Handler      func() error
	MaxRetries   int
	RetryDelays  []time.Duration
}
