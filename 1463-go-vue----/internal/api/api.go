package api

type Frequency string

const (
	FrequencyDaily   Frequency = "daily"
	FrequencyWeekly  Frequency = "weekly"
	FrequencyMonthly Frequency = "monthly"
)

type HazardLevel string

const (
	HazardLevelGeneral    HazardLevel = "general"
	HazardLevelMajor      HazardLevel = "major"
	HazardLevelCritical   HazardLevel = "critical"
)

type InspectionItemResult string

const (
	ItemResultPass    InspectionItemResult = "pass"
	ItemResultFail    InspectionItemResult = "fail"
)

type TaskStatus string

const (
	TaskStatusPending  TaskStatus = "pending"
	TaskStatusComplete TaskStatus = "completed"
)

type HazardStatus string

const (
	HazardStatusOpen       HazardStatus = "open"
	HazardStatusInProgress HazardStatus = "in_progress"
	HazardStatusPendingReview HazardStatus = "pending_review"
	HazardStatusClosed     HazardStatus = "closed"
)

type LevelChangeRequestStatus string

const (
	LevelChangeStatusPending   LevelChangeRequestStatus = "pending"
	LevelChangeStatusApproved  LevelChangeRequestStatus = "approved"
	LevelChangeStatusRejected  LevelChangeRequestStatus = "rejected"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type CreateZoneRequest struct {
	Name     string  `json:"name"`
	ParentID *string `json:"parent_id,omitempty"`
}

type ZoneResponse struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	ParentID *string         `json:"parent_id,omitempty"`
	Children []ZoneResponse  `json:"children,omitempty"`
}

type CreateInspectionPlanRequest struct {
	Name         string    `json:"name"`
	ZoneID       string    `json:"zone_id"`
	InspectorIDs []string  `json:"inspector_ids"`
	Items        []string  `json:"items"`
	Frequency    Frequency `json:"frequency"`
}

type InspectionPlanResponse struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	ZoneID       string      `json:"zone_id"`
	ZoneName     string      `json:"zone_name"`
	InspectorIDs []string    `json:"inspector_ids"`
	Items        []string    `json:"items"`
	Frequency    Frequency   `json:"frequency"`
	CreatedAt    string      `json:"created_at"`
}

type InspectionTaskResponse struct {
	ID           string                 `json:"id"`
	PlanID       string                 `json:"plan_id"`
	PlanName     string                 `json:"plan_name"`
	ZoneID       string                 `json:"zone_id"`
	ZoneName     string                 `json:"zone_name"`
	InspectorID  string                 `json:"inspector_id"`
	Items        []TaskItemResponse     `json:"items"`
	Status       TaskStatus             `json:"status"`
	ScheduledFor string                 `json:"scheduled_for"`
	CompletedAt  *string                `json:"completed_at,omitempty"`
}

type TaskItemResponse struct {
	Index        int                    `json:"index"`
	Name         string                 `json:"name"`
	Result       *InspectionItemResult  `json:"result,omitempty"`
	Description  *string                `json:"description,omitempty"`
}

type SubmitInspectionRequest struct {
	TaskID string                   `json:"task_id"`
	Items  []SubmitInspectionItem   `json:"items"`
}

type SubmitInspectionItem struct {
	Index       int                    `json:"index"`
	Result      InspectionItemResult   `json:"result"`
	Description string                 `json:"description,omitempty"`
}

type HazardResponse struct {
	ID              string       `json:"id"`
	TaskID          string       `json:"task_id"`
	ZoneID          string       `json:"zone_id"`
	ZoneName        string       `json:"zone_name"`
	ItemName        string       `json:"item_name"`
	Description     string       `json:"description"`
	Level           HazardLevel  `json:"level"`
	Status          HazardStatus `json:"status"`
	ResponsibleDept string       `json:"responsible_dept"`
	DueAt           string       `json:"due_at"`
	RemediationNote *string      `json:"remediation_note,omitempty"`
	RemediatedAt    *string      `json:"remediated_at,omitempty"`
	ClosedAt        *string      `json:"closed_at,omitempty"`
	CreatedAt       string       `json:"created_at"`
}

type SubmitRemediationRequest struct {
	HazardID string `json:"hazard_id"`
	Note     string `json:"note"`
}

type ReviewRemediationRequest struct {
	HazardID string `json:"hazard_id"`
	Approved bool   `json:"approved"`
}

type RequestLevelChangeRequest struct {
	HazardID     string      `json:"hazard_id"`
	ProposedLevel HazardLevel `json:"proposed_level"`
	Reason       string      `json:"reason"`
}

type ReviewLevelChangeRequest struct {
	RequestID string `json:"request_id"`
	Approved  bool   `json:"approved"`
}

type LevelChangeRequestResponse struct {
	ID            string                  `json:"id"`
	HazardID      string                  `json:"hazard_id"`
	ProposedLevel HazardLevel             `json:"proposed_level"`
	Reason        string                  `json:"reason"`
	Status        LevelChangeRequestStatus `json:"status"`
	CreatedAt     string                  `json:"created_at"`
}

type AuditLogResponse struct {
	ID          string `json:"id"`
	Operation   string `json:"operation"`
	EntityType  string `json:"entity_type"`
	EntityID    string `json:"entity_id"`
	Details     string `json:"details"`
	PerformedAt string `json:"performed_at"`
}
