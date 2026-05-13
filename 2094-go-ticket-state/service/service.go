package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"ticket-system/database"
	"ticket-system/models"
)

type TicketService struct {
	db *database.DB
}

func NewTicketService(db *database.DB) *TicketService {
	return &TicketService{db: db}
}

type InvalidTransitionError struct {
	AllowedStatuses []string
}

func (e *InvalidTransitionError) Error() string {
	return "invalid status transition"
}

type AlreadyAssignedError struct {
	Assignee string
}

func (e *AlreadyAssignedError) Error() string {
	return fmt.Sprintf("ticket already assigned to %s", e.Assignee)
}

var (
	ErrTicketNotFound      = errors.New("ticket not found")
	ErrTicketClosed        = errors.New("ticket is closed")
	ErrInvalidPriority     = errors.New("invalid priority")
	ErrPriorityNotSpecified = errors.New("priority not specified")
)

func IsValidPriority(p string) bool {
	switch p {
	case models.PriorityUrgent, models.PriorityHigh, models.PriorityMedium, models.PriorityLow:
		return true
	default:
		return false
	}
}

func IsValidTransition(from, to string) bool {
	allowed, ok := models.StatusTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

func (s *TicketService) GetAllowedTransitions(status string) []string {
	if transitions, ok := models.StatusTransitions[status]; ok {
		return transitions
	}
	return []string{}
}

func (s *TicketService) CreateTicket(ctx context.Context, req *models.CreateTicketRequest) (*models.Ticket, error) {
	req.Priority = strings.ToLower(strings.TrimSpace(req.Priority))
	if req.Priority == "" {
		return nil, ErrPriorityNotSpecified
	}
	if !IsValidPriority(req.Priority) {
		return nil, ErrInvalidPriority
	}

	if req.ParentTicketID != nil {
		parent, err := s.db.GetTicketByID(ctx, *req.ParentTicketID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, ErrTicketNotFound
		}
	}

	return s.db.CreateTicket(ctx, req)
}

func (s *TicketService) GetTicket(ctx context.Context, id int64) (*models.Ticket, error) {
	ticket, err := s.db.GetTicketByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, ErrTicketNotFound
	}
	return ticket, nil
}

func (s *TicketService) AssignTicket(ctx context.Context, ticketID int64, userID string) error {
	ticket, err := s.db.GetTicketByID(ctx, ticketID)
	if err != nil {
		return err
	}
	if ticket == nil {
		return ErrTicketNotFound
	}
	if ticket.Status == models.TicketStatusClosed {
		return ErrTicketClosed
	}

	err = s.db.AssignTicket(ctx, ticketID, userID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "already assigned to ") {
			assignee := strings.TrimPrefix(err.Error(), "already assigned to ")
			return &AlreadyAssignedError{Assignee: assignee}
		}
		if err == sql.ErrNoRows {
			return ErrTicketNotFound
		}
		return err
	}

	return nil
}

func (s *TicketService) UpdateStatus(ctx context.Context, ticketID int64, toStatus, userID string) error {
	toStatus = strings.ToLower(strings.TrimSpace(toStatus))

	ticket, err := s.db.GetTicketByID(ctx, ticketID)
	if err != nil {
		return err
	}
	if ticket == nil {
		return ErrTicketNotFound
	}

	if ticket.Status == models.TicketStatusClosed {
		return ErrTicketClosed
	}

	if !IsValidTransition(ticket.Status, toStatus) {
		return &InvalidTransitionError{
			AllowedStatuses: s.GetAllowedTransitions(ticket.Status),
		}
	}

	if err := s.db.UpdateTicketStatus(ctx, nil, ticketID, ticket.Status, toStatus, userID); err != nil {
		if err == sql.ErrNoRows {
			return ErrTicketNotFound
		}
		return err
	}

	if err := s.SyncRelatedTickets(ctx, ticketID); err != nil {
		return err
	}

	return nil
}

func (s *TicketService) QueryTickets(ctx context.Context, params *database.QueryTicketsParams) ([]*models.Ticket, error) {
	return s.db.QueryTickets(ctx, params)
}

func (s *TicketService) SyncRelatedTickets(ctx context.Context, ticketID int64) error {
	links, err := s.db.GetUnprocessedLinks(ctx, ticketID)
	if err != nil {
		return err
	}
	if len(links) == 0 {
		return nil
	}

	processedIDs := make([]int64, 0)
	for _, link := range links {
		ticket, err := s.db.GetTicketByID(ctx, ticketID)
		if err != nil {
			return err
		}
		if ticket == nil {
			continue
		}

		child, err := s.db.GetTicketByID(ctx, link.ChildID)
		if err != nil {
			return err
		}
		if child == nil {
			continue
		}

		processedIDs = append(processedIDs, link.ID)
	}

	if len(processedIDs) > 0 {
		return s.db.MarkLinksProcessed(ctx, processedIDs)
	}
	return nil
}

func (s *TicketService) GetTicketHistory(ctx context.Context, ticketID int64) ([]*models.TicketHistory, error) {
	ticket, err := s.db.GetTicketByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, ErrTicketNotFound
	}
	return s.db.GetTicketHistory(ctx, ticketID)
}
