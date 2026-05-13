package models

import "time"

const (
	TicketStatusNew         = "new"
	TicketStatusAssigned    = "assigned"
	TicketStatusInProgress  = "in_progress"
	TicketStatusPendingCust = "pending_customer"
	TicketStatusResolved    = "resolved"
	TicketStatusClosed      = "closed"
)

const (
	PriorityUrgent = "urgent"
	PriorityHigh   = "high"
	PriorityMedium = "medium"
	PriorityLow    = "low"
)

const (
	CustomerConfirmationDays = 3
	AutoCloseDays            = 7
)

var StatusTransitions = map[string][]string{
	TicketStatusNew:         {TicketStatusAssigned, TicketStatusClosed},
	TicketStatusAssigned:    {TicketStatusInProgress},
	TicketStatusInProgress:  {TicketStatusPendingCust, TicketStatusAssigned},
	TicketStatusPendingCust: {TicketStatusResolved},
	TicketStatusResolved:    {TicketStatusClosed},
	TicketStatusClosed:      {},
}

var PriorityOrder = map[string]int{
	PriorityUrgent: 0,
	PriorityHigh:   1,
	PriorityMedium: 2,
	PriorityLow:    3,
}

type Ticket struct {
	ID               int64      `json:"id"`
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	Status           string     `json:"status"`
	Priority         string     `json:"priority"`
	CreatorID        string     `json:"creator_id"`
	AssigneeID       *string    `json:"assignee_id"`
	ParentTicketID   *int64     `json:"parent_ticket_id"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	PendingSince     *time.Time `json:"pending_since"`
	ResolvedSince    *time.Time `json:"resolved_since"`
}

type TicketHistory struct {
	ID         int64     `json:"id"`
	TicketID   int64     `json:"ticket_id"`
	FromStatus string    `json:"from_status"`
	ToStatus   string    `json:"to_status"`
	UserID     string    `json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type TicketLink struct {
	ID          int64     `json:"id"`
	ParentID    int64     `json:"parent_id"`
	ChildID     int64     `json:"child_id"`
	IsProcessed bool      `json:"is_processed"`
	CreatedAt   time.Time `json:"created_at"`
}

type ErrorResponse struct {
	Error        string   `json:"error"`
	AllowedStatuses []string `json:"allowed_statuses,omitempty"`
	CurrentAssignee string  `json:"current_assignee,omitempty"`
}

type CreateTicketRequest struct {
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	Priority       string  `json:"priority"`
	CreatorID      string  `json:"creator_id"`
	ParentTicketID *int64  `json:"parent_ticket_id"`
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
	UserID string `json:"user_id"`
}

type AssignRequest struct {
	UserID string `json:"user_id"`
}
