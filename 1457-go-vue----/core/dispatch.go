package core

import (
	"complaint-system/common"
	"errors"
	"time"
)

func (s *Service) DispatchTicket(ticketNo string, req *common.DispatchTicketRequest) error {
	if req.ResponsibleDepartment == "" || req.ResponsiblePerson == "" {
		return errors.New("responsible_department and responsible_person are required")
	}
	if req.UrgencyLevel != common.UrgencyNormal && req.UrgencyLevel != common.UrgencyUrgent && req.UrgencyLevel != common.UrgencySuper {
		return errors.New("invalid urgency_level")
	}

	t, ok := s.store.get(ticketNo)
	if !ok {
		return ErrTicketNotFound
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Status != common.StatusPendingDispatch {
		return ErrInvalidStatus
	}

	now := time.Now()
	s.checkTimeout(t, now)

	t.ResponsibleDepartment = req.ResponsibleDepartment
	t.ResponsiblePerson = req.ResponsiblePerson
	t.UrgencyLevel = &req.UrgencyLevel
	t.DispatcherID = req.DispatcherID
	t.DispatchedAt = &now
	t.Status = common.StatusDispatched

	if req.UrgencyLevel == common.UrgencySuper {
		s.sendSuperUrgentNotification(t)
	}

	return nil
}

func (s *Service) StartProcessing(ticketNo string) error {
	t, ok := s.store.get(ticketNo)
	if !ok {
		return ErrTicketNotFound
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Status != common.StatusDispatched {
		return ErrInvalidStatus
	}

	t.Status = common.StatusProcessing
	return nil
}

func (s *Service) CompleteProcessing(ticketNo string, req *common.CompleteProcessingRequest) error {
	if req.ProcessingResult == "" {
		return errors.New("processing_result is required")
	}

	t, ok := s.store.get(ticketNo)
	if !ok {
		return ErrTicketNotFound
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Status != common.StatusProcessing {
		return ErrInvalidStatus
	}

	now := time.Now()
	t.ProcessingResult = req.ProcessingResult
	t.ProcessedAt = &now
	t.Status = common.StatusPendingReview

	return nil
}

func (s *Service) sendSuperUrgentNotification(t *Ticket) {
}
