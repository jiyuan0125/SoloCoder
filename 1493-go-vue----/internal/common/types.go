package common

import "time"

type Frequency string

const (
	FrequencyDaily   Frequency = "daily"
	FrequencyTriWeek Frequency = "tri_weekly"
	FrequencyBiWeek  Frequency = "bi_weekly"
	FrequencyWeekly  Frequency = "weekly"
)

type CleanType string

const (
	CleanTypeBasic  CleanType = "basic"
	CleanTypeDeep   CleanType = "deep"
)

type SkillLevel string

const (
	SkillBasic SkillLevel = "basic"
	SkillDeep  SkillLevel = "deep"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusAssigned   TaskStatus = "assigned"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusRecheck    TaskStatus = "recheck"
	TaskStatusFailed     TaskStatus = "failed"
)

type LeaveStatus string

const (
	LeaveStatusPending   LeaveStatus = "pending"
	LeaveStatusApproved  LeaveStatus = "approved"
	LeaveStatusRejected  LeaveStatus = "rejected"
)

type Client struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Contract     *Contract `json:"contract"`
	CreatedAt    time.Time `json:"created_at"`
}

type Contract struct {
	ID             string       `json:"id"`
	ClientID       string       `json:"client_id"`
	ServiceAddress string       `json:"service_address"`
	Area           float64      `json:"area"`
	Frequency      Frequency    `json:"frequency"`
	SpecialNotes   string       `json:"special_notes"`
	ServiceAreas   []string     `json:"service_areas"`
	StartDate      time.Time    `json:"start_date"`
	EndDate        time.Time    `json:"end_date"`
}

type ServiceArea struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Address  string  `json:"address"`
	Area     float64 `json:"area"`
	ClientID string  `json:"client_id"`
	ZoneID   string  `json:"zone_id"`
}

type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Cleaner struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Skill     SkillLevel `json:"skill"`
	TeamID    string     `json:"team_id"`
	CreatedAt time.Time  `json:"created_at"`
}

type Team struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	ZoneID string `json:"zone_id"`
}

type WeeklySchedule struct {
	ID         string                  `json:"id"`
	TeamID     string                  `json:"team_id"`
	WeekStart  time.Time               `json:"week_start"`
	WeekEnd    time.Time               `json:"week_end"`
	DaySchedules map[time.Weekday][]Shift `json:"day_schedules"`
}

type Shift struct {
	StartHour int `json:"start_hour"`
	EndHour   int `json:"end_hour"`
}

type LeaveRequest struct {
	ID          string      `json:"id"`
	CleanerID   string      `json:"cleaner_id"`
	Date        time.Time   `json:"date"`
	Reason      string      `json:"reason"`
	Status      LeaveStatus `json:"status"`
	RequestedAt time.Time   `json:"requested_at"`
	ApprovedAt  *time.Time  `json:"approved_at,omitempty"`
}

type Task struct {
	ID              string     `json:"id"`
	TaskNo          string     `json:"task_no"`
	ClientName      string     `json:"client_name"`
	ServiceAddress  string     `json:"service_address"`
	ServiceAreaID   string     `json:"service_area_id"`
	ServiceAreaName string     `json:"service_area_name"`
	CleanType       CleanType  `json:"clean_type"`
	EstimatedHours  float64    `json:"estimated_hours"`
	CleanerID       string     `json:"cleaner_id,omitempty"`
	CleanerName     string     `json:"cleaner_name,omitempty"`
	Date            time.Time  `json:"date"`
	Status          TaskStatus `json:"status"`
	QualityCheckID  string     `json:"quality_check_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type QualityCheck struct {
	ID          string    `json:"id"`
	TaskID      string    `json:"task_id"`
	InspectorID string    `json:"inspector_id"`
	Inspector   string    `json:"inspector"`
	FloorScore  int       `json:"floor_score"`
	DeskScore   int       `json:"desk_score"`
	TrashScore  int       `json:"trash_score"`
	TotalScore  int       `json:"total_score"`
	Notes       string    `json:"notes"`
	CheckedAt   time.Time `json:"checked_at"`
	CheckCount  int       `json:"check_count"`
}

type TodoItem struct {
	ID          string    `json:"id"`
	TaskID      string    `json:"task_id"`
	CleanerID   string    `json:"cleaner_id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
}
