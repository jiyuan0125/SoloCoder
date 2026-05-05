package common

import "time"

type Ticket struct {
	ID            string            `json:"id"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Category      TicketCategory    `json:"category"`
	Priority      TicketPriority    `json:"priority"`
	Status        TicketStatus      `json:"status"`
	SubmitterID   string            `json:"submitter_id"`
	SubmitterName string            `json:"submitter_name"`
	AssigneeID    string            `json:"assignee_id,omitempty"`
	AssigneeName  string            `json:"assignee_name,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	FirstResponse time.Time         `json:"first_response,omitempty"`
	ClosedAt      time.Time         `json:"closed_at,omitempty"`
	Rating        int               `json:"rating,omitempty"`
	RatingComment string            `json:"rating_comment,omitempty"`
	ParentID      string            `json:"parent_id,omitempty"`
	ChildIDs      []string          `json:"child_ids,omitempty"`
	LinkedTicketID string           `json:"linked_ticket_id,omitempty"`
	EscalationLevel int             `json:"escalation_level"`
	SLAResponseDeadline time.Time  `json:"sla_response_deadline"`
	SLAIsMet      bool              `json:"sla_is_met"`
}

type OperationLog struct {
	ID            string         `json:"id"`
	TicketID      string         `json:"ticket_id"`
	OperationType OperationType  `json:"operation_type"`
	OperatorID    string         `json:"operator_id"`
	OperatorName  string         `json:"operator_name"`
	OldValue      string         `json:"old_value,omitempty"`
	NewValue      string         `json:"new_value,omitempty"`
	Comment       string         `json:"comment,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

type Handler struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Category TicketCategory `json:"category"`
	Level    int            `json:"level"`
	IsManager bool          `json:"is_manager"`
}

type CreateTicketRequest struct {
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	Category      TicketCategory `json:"category,omitempty"`
	Priority      TicketPriority `json:"priority"`
	SubmitterID   string         `json:"submitter_id"`
	SubmitterName string         `json:"submitter_name"`
	AutoClassify  bool           `json:"auto_classify,omitempty"`
}

type AssignTicketRequest struct {
	AssigneeID   string `json:"assignee_id"`
	AssigneeName string `json:"assignee_name"`
}

type ReassignTicketRequest struct {
	NewAssigneeID   string `json:"new_assignee_id"`
	NewAssigneeName string `json:"new_assignee_name"`
	Reason          string `json:"reason"`
}

type UpdateStatusRequest struct {
	Status     TicketStatus `json:"status"`
	OperatorID string       `json:"operator_id"`
	Comment    string       `json:"comment,omitempty"`
}

type RateTicketRequest struct {
	Rating    int    `json:"rating"`
	Comment   string `json:"comment,omitempty"`
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
}

type LinkTicketsRequest struct {
	ParentID string `json:"parent_id"`
	ChildID  string `json:"child_id"`
}

type TransferTicketRequest struct {
	TargetTeam     string `json:"target_team"`
	Reason         string `json:"reason"`
	OperatorID     string `json:"operator_id"`
	OperatorName   string `json:"operator_name"`
}

type QuickReplyTemplate struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Category TicketCategory `json:"category"`
}

type StatisticsReport struct {
	ByCategory       map[TicketCategory]CategoryStats `json:"by_category"`
	AverageHandleTime float64                          `json:"average_handle_time_hours"`
	OverallSLARate    float64                          `json:"overall_sla_rate"`
	TotalTickets      int                              `json:"total_tickets"`
}

type CategoryStats struct {
	Count           int     `json:"count"`
	HandleTimeHours float64 `json:"average_handle_time_hours"`
	SLARate         float64 `json:"sla_rate"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Code    int         `json:"code,omitempty"`
}
