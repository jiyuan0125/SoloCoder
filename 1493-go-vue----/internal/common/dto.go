package common

import "time"

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type CreateClientRequest struct {
	Name           string    `json:"name"`
	ServiceAddress string    `json:"service_address"`
	Area           float64   `json:"area"`
	Frequency      Frequency `json:"frequency"`
	SpecialNotes   string    `json:"special_notes"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
}

type CreateClientResponse struct {
	ClientID string `json:"client_id"`
}

type CreateServiceAreaRequest struct {
	Name     string  `json:"name"`
	Address  string  `json:"address"`
	Area     float64 `json:"area"`
	ClientID string  `json:"client_id"`
	ZoneID   string  `json:"zone_id"`
}

type CreateServiceAreaResponse struct {
	ServiceAreaID string `json:"service_area_id"`
}

type CreateZoneRequest struct {
	Name string `json:"name"`
}

type CreateZoneResponse struct {
	ZoneID string `json:"zone_id"`
}

type CreateTeamRequest struct {
	Name   string `json:"name"`
	ZoneID string `json:"zone_id"`
}

type CreateTeamResponse struct {
	TeamID string `json:"team_id"`
}

type CreateCleanerRequest struct {
	Name   string     `json:"name"`
	Skill  SkillLevel `json:"skill"`
	TeamID string     `json:"team_id"`
}

type CreateCleanerResponse struct {
	CleanerID string `json:"cleaner_id"`
}

type CreateScheduleRequest struct {
	TeamID    string                          `json:"team_id"`
	WeekStart time.Time                       `json:"week_start"`
	WeekEnd   time.Time                       `json:"week_end"`
	Days      map[time.Weekday][]ShiftRequest `json:"days"`
}

type ShiftRequest struct {
	StartHour int `json:"start_hour"`
	EndHour   int `json:"end_hour"`
}

type CreateScheduleResponse struct {
	ScheduleID string `json:"schedule_id"`
}

type CreateLeaveRequest struct {
	CleanerID string    `json:"cleaner_id"`
	Date      time.Time `json:"date"`
	Reason    string    `json:"reason"`
}

type CreateLeaveResponse struct {
	LeaveID string `json:"leave_id"`
}

type ApproveLeaveRequest struct {
	Approved bool `json:"approved"`
}

type GenerateTasksRequest struct {
	WeekStart time.Time `json:"week_start"`
	WeekEnd   time.Time `json:"week_end"`
}

type GenerateTasksResponse struct {
	TotalTasks   int `json:"total_tasks"`
	Assigned     int `json:"assigned"`
	Pending      int `json:"pending"`
}

type CompleteTaskRequest struct {
	TaskID string `json:"task_id"`
}

type CreateQualityCheckRequest struct {
	TaskID      string `json:"task_id"`
	InspectorID string `json:"inspector_id"`
	Inspector   string `json:"inspector"`
	FloorScore  int    `json:"floor_score"`
	DeskScore   int    `json:"desk_score"`
	TrashScore  int    `json:"trash_score"`
	Notes       string `json:"notes"`
}

type ListRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type ListResponse struct {
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Items    interface{} `json:"items"`
}
