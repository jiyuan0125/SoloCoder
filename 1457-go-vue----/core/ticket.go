package core

import (
	"complaint-system/common"
	"errors"
	"fmt"
	"time"
)

const (
	maxRetryCount = 2
)

var (
	ErrTicketNotFound = errors.New("ticket not found")
	ErrInvalidStatus  = errors.New("invalid ticket status")
)

type Service struct {
	store *TicketStore
}

func NewService() *Service {
	return &Service{
		store: NewTicketStore(),
	}
}

func (s *Service) generateTicketNo(now time.Time) string {
	day := now.Format("20060102")
	seq := s.store.nextSeq(day)
	return fmt.Sprintf("TS%s%04d", day, seq)
}

func getResponseDeadline(t common.ComplaintType, created time.Time) time.Time {
	switch t {
	case common.TypeServiceQuality, common.TypeProductQuality, common.TypeAfterSales, common.TypeOther:
		return created.Add(24 * time.Hour)
	case common.TypeLogistics:
		return created.Add(12 * time.Hour)
	default:
		return created.Add(24 * time.Hour)
	}
}

func (s *Service) CreateTicket(req *common.CreateTicketRequest) (string, error) {
	if req.ComplainerName == "" || req.ContactPhone == "" || req.Content == "" {
		return "", errors.New("complainer_name, contact_phone and content are required")
	}
	if req.ComplaintChannel == "" || req.ComplaintType == "" || req.Region == "" {
		return "", errors.New("complaint_channel, complaint_type and region are required")
	}

	now := time.Now()
	ticketNo := s.generateTicketNo(now)
	deadline := getResponseDeadline(req.ComplaintType, now)

	ticket := &Ticket{
		TicketNo:         ticketNo,
		ComplainerName:   req.ComplainerName,
		ContactPhone:     req.ContactPhone,
		Content:          req.Content,
		ComplaintChannel: req.ComplaintChannel,
		ComplaintType:    req.ComplaintType,
		Region:           req.Region,
		Status:           common.StatusPendingDispatch,
		RetryCount:       0,
		IsEscalated:      false,
		IsTimeout:        false,
		ExternalSystemID: req.ExternalSystemID,
		CreatedAt:        now,
		ResponseDeadline: deadline,
	}

	s.store.add(ticket)
	return ticketNo, nil
}

func (s *Service) GetTicket(ticketNo string) (*common.TicketResponse, error) {
	t, ok := s.store.get(ticketNo)
	if !ok {
		return nil, ErrTicketNotFound
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	return ticketToResponse(t), nil
}

func (s *Service) ListTickets() []common.TicketResponse {
	tickets := s.store.list()
	now := time.Now()
	result := make([]common.TicketResponse, 0, len(tickets))
	for _, t := range tickets {
		t.mu.Lock()
		s.checkTimeout(t, now)
		t.mu.Unlock()
		t.mu.RLock()
		result = append(result, *ticketToResponse(t))
		t.mu.RUnlock()
	}
	return result
}

func (s *Service) checkTimeout(t *Ticket, now time.Time) {
	if t.Status == common.StatusPendingDispatch || t.Status == common.StatusDispatched || t.Status == common.StatusProcessing {
		if now.After(t.ResponseDeadline) && !t.IsTimeout {
			t.Status = common.StatusTimeout
			t.IsTimeout = true
		}
	}
}

func ticketToResponse(t *Ticket) *common.TicketResponse {
	resp := &common.TicketResponse{
		TicketNo:              t.TicketNo,
		ComplainerName:        t.ComplainerName,
		ContactPhone:          t.ContactPhone,
		Content:               t.Content,
		ComplaintChannel:      t.ComplaintChannel,
		ComplaintType:         t.ComplaintType,
		Region:                t.Region,
		Status:                t.Status,
		UrgencyLevel:          t.UrgencyLevel,
		ResponsibleDepartment: t.ResponsibleDepartment,
		ResponsiblePerson:     t.ResponsiblePerson,
		ProcessingResult:      t.ProcessingResult,
		ReviewResult:          t.ReviewResult,
		ReviewRemark:          t.ReviewRemark,
		RetryCount:            t.RetryCount,
		IsEscalated:           t.IsEscalated,
		ExternalSystemID:      t.ExternalSystemID,
		CreatedAt:             t.CreatedAt,
		DispatchedAt:          t.DispatchedAt,
		ProcessedAt:           t.ProcessedAt,
		ClosedAt:              t.ClosedAt,
		ResponseDeadline:      t.ResponseDeadline,
	}
	return resp
}
