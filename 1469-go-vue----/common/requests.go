package common

import "time"

type CreateSupervisoryRecordRequest struct {
	ProjectPart      string    `json:"project_part"`
	Process          string    `json:"process"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	ConstructionDesc string    `json:"construction_desc"`
	IssuesFound      string    `json:"issues_found"`
	Supervisor       string    `json:"supervisor"`
}

type AddSupervisorySupplementRequest struct {
	ID          string `json:"id"`
	Supplement  string `json:"supplement"`
}

type CreateAcceptanceRecordRequest struct {
	ProjectPart    string           `json:"project_part"`
	Items          []AcceptanceItem `json:"items"`
	Conclusion     string           `json:"conclusion"`
	Status         AcceptanceStatus `json:"status"`
	Supervisors    []string         `json:"supervisors"`
	AcceptanceDate time.Time        `json:"acceptance_date"`
}

type CreateRectificationNoticeRequest struct {
	AcceptanceID  string    `json:"acceptance_id"`
	Contractor    string    `json:"contractor"`
	Requirements  string    `json:"requirements"`
	Deadline      time.Time `json:"deadline"`
}

type CompleteRectificationRequest struct {
	ID string `json:"id"`
}

type RecheckAcceptanceRequest struct {
	AcceptanceID    string           `json:"acceptance_id"`
	Items           []AcceptanceItem `json:"items"`
	Conclusion      string           `json:"conclusion"`
	Status          AcceptanceStatus `json:"status"`
	Supervisors     []string         `json:"supervisors"`
	AcceptanceDate  time.Time        `json:"acceptance_date"`
}

type CreateIssueRequest struct {
	Severity         Severity    `json:"severity"`
	Description      string      `json:"description"`
	DiscoveredDate   time.Time   `json:"discovered_date"`
	ResponsibleUnit  string      `json:"responsible_unit"`
	RectificationReq string      `json:"rectification_req"`
	PlanFinishDate   time.Time   `json:"plan_finish_date"`
	Status           IssueStatus `json:"status"`
	Supervisor       string      `json:"supervisor"`
}

type UpdateIssueStatusRequest struct {
	ID     string      `json:"id"`
	Status IssueStatus `json:"status"`
}

type SubmitDailyLogRequest struct {
	Supervisor        string    `json:"supervisor"`
	LogDate           time.Time `json:"log_date"`
	AdditionalContent string    `json:"additional_content"`
}

type GetDailyLogRequest struct {
	Supervisor string    `json:"supervisor"`
	LogDate    time.Time `json:"log_date"`
}
