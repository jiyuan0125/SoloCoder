package api

import "time"

type CreateDeviceRequest struct {
	Code          string       `json:"code"`
	Type          DeviceType   `json:"type"`
	Location      string       `json:"location"`
	InstallDate   time.Time    `json:"install_date"`
	ExpiryDate    time.Time    `json:"expiry_date"`
	LastCheckDate time.Time    `json:"last_check_date"`
}

type UpdateDeviceRequest struct {
	Code          *string       `json:"code,omitempty"`
	Location      *string       `json:"location,omitempty"`
	InstallDate   *time.Time    `json:"install_date,omitempty"`
	ExpiryDate    *time.Time    `json:"expiry_date,omitempty"`
	LastCheckDate *time.Time    `json:"last_check_date,omitempty"`
	Status        *DeviceStatus `json:"status,omitempty"`
}

type ListDevicesRequest struct {
	Type     *DeviceType   `json:"type,omitempty"`
	Location *string       `json:"location,omitempty"`
	Status   *DeviceStatus `json:"status,omitempty"`
}

type CreateInspectionPointRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

type CreateInspectionRouteRequest struct {
	Name      string   `json:"name"`
	PointIDs  []string `json:"point_ids"`
}

type CreateInspectionPlanRequest struct {
	RouteID       string              `json:"route_id"`
	InspectorID   string              `json:"inspector_id"`
	InspectorName string              `json:"inspector_name"`
	Frequency     InspectionFrequency `json:"frequency"`
	StartTime     time.Time           `json:"start_time"`
}

type CheckTaskPointRequest struct {
	TaskID  string `json:"task_id"`
	PointID string `json:"point_id"`
	Normal  bool   `json:"normal"`
}

type CreateDrillPlanRequest struct {
	Type             DrillType `json:"type"`
	Name             string    `json:"name"`
	ScheduledTime    time.Time `json:"scheduled_time"`
	PlannedAttendees int       `json:"planned_attendees"`
}

type CompleteDrillRequest struct {
	ActualAttendees int    `json:"actual_attendees"`
	DurationMinutes int    `json:"duration_minutes"`
	Improvements    string `json:"improvements"`
}

type ListRemindersRequest struct {
	Type *ReminderType `json:"type,omitempty"`
	Read *bool         `json:"read,omitempty"`
}

type MarkReminderReadRequest struct {
	Read bool `json:"read"`
}
