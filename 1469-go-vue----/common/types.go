package common

import "time"

type SupervisoryRecord struct {
	ID               string    `json:"id"`
	ProjectPart      string    `json:"project_part"`
	Process          string    `json:"process"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	DurationMinutes  int       `json:"duration_minutes"`
	IsAbnormal       bool      `json:"is_abnormal"`
	ConstructionDesc string    `json:"construction_desc"`
	IssuesFound      string    `json:"issues_found"`
	Supervisor       string    `json:"supervisor"`
	Supplements      []string  `json:"supplements"`
	CreatedAt        time.Time `json:"created_at"`
	Submitted        bool      `json:"submitted"`
}

type AcceptanceItem struct {
	Name   string `json:"name"`
	Result string `json:"result"`
}

type AcceptanceRecord struct {
	ID              string           `json:"id"`
	ProjectPart     string           `json:"project_part"`
	Items           []AcceptanceItem `json:"items"`
	Conclusion      string           `json:"conclusion"`
	Status          AcceptanceStatus `json:"status"`
	Supervisors     []string         `json:"supervisors"`
	AcceptanceDate  time.Time        `json:"acceptance_date"`
	RectificationID string           `json:"rectification_id,omitempty"`
	RecheckCount    int              `json:"recheck_count"`
	CreatedAt       time.Time        `json:"created_at"`
}

type RectificationNotice struct {
	ID              string           `json:"id"`
	AcceptanceID    string           `json:"acceptance_id"`
	Contractor      string           `json:"contractor"`
	Requirements    string           `json:"requirements"`
	Deadline        time.Time        `json:"deadline"`
	Status          string           `json:"status"`
	CreatedAt       time.Time        `json:"created_at"`
	CompletedAt     *time.Time       `json:"completed_at,omitempty"`
}

type Issue struct {
	ID               string      `json:"id"`
	Severity         Severity    `json:"severity"`
	Description      string      `json:"description"`
	DiscoveredDate   time.Time   `json:"discovered_date"`
	ResponsibleUnit  string      `json:"responsible_unit"`
	RectificationReq string      `json:"rectification_req"`
	PlanFinishDate   time.Time   `json:"plan_finish_date"`
	Status           IssueStatus `json:"status"`
	ReportedToPM     bool        `json:"reported_to_pm"`
	Supervisor       string      `json:"supervisor"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

type DailyLog struct {
	ID                string              `json:"id"`
	Supervisor        string              `json:"supervisor"`
	LogDate           time.Time           `json:"log_date"`
	SupervisoryRecord []*SupervisoryRecord `json:"supervisory_records"`
	AcceptanceRecords []*AcceptanceRecord `json:"acceptance_records"`
	AdditionalContent string              `json:"additional_content"`
	Submitted         bool                `json:"submitted"`
	CreatedAt         time.Time           `json:"created_at"`
}
