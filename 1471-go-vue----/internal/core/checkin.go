package core

import (
	"time"
)

func (s *Store) CheckIn(ticketNo string) (*CheckInRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	ticket, exists := s.tickets[ticketNo]
	if !exists {
		return nil, ErrTicketNotFound
	}
	
	if ticket.Status == TicketStatusChecked {
		return nil, ErrAlreadyCheckedIn
	}
	
	if ticket.Status == TicketStatusRefunded {
		return nil, ErrAlreadyRefunded
	}
	
	if ticket.Status == TicketStatusNotBoarded {
		return nil, ErrRefundNotAllowed
	}
	
	now := time.Now()
	ticket.Status = TicketStatusChecked
	ticket.CheckedAt = &now
	
	record := &CheckInRecord{
		ID:         s.generateCheckInID(),
		TicketNo:   ticketNo,
		ScheduleNo: ticket.ScheduleNo,
		Date:       ticket.Date,
		SeatNo:     ticket.SeatNo,
		CheckedAt:  now,
	}
	s.checkInRecords[record.ID] = record
	
	return record, nil
}

func (s *Store) MarkNotBoarded(scheduleNo, date string) ([]*Ticket, error) {
	key := s.GetScheduleKey(scheduleNo, date)
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	schedule, exists := s.schedules[key]
	if !exists {
		return nil, ErrScheduleNotFound
	}
	
	markedTickets := make([]*Ticket, 0)
	for _, ticket := range s.tickets {
		if ticket.ScheduleNo == scheduleNo &&
			ticket.Date == date &&
			ticket.Status == TicketStatusSold {
			ticket.Status = TicketStatusNotBoarded
			
			refundReq := &RefundRequest{
				ID:           s.generateRefundID(),
				TicketNo:     ticket.TicketNo,
				Status:       RefundStatusPending,
				RefundAmount: 0,
				RequestTime:  time.Now(),
			}
			s.refundRequests[refundReq.ID] = refundReq
			
			markedTickets = append(markedTickets, ticket)
		}
	}
	
	schedule.UpdatedAt = time.Now()
	
	return markedTickets, nil
}

func (s *Store) GetCheckInRecord(ticketNo string) *CheckInRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	for _, record := range s.checkInRecords {
		if record.TicketNo == ticketNo {
			return record
		}
	}
	return nil
}

func (s *Store) ListCheckInRecords(scheduleNo, date string) []*CheckInRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := make([]*CheckInRecord, 0)
	for _, record := range s.checkInRecords {
		if record.ScheduleNo == scheduleNo && record.Date == date {
			result = append(result, record)
		}
	}
	return result
}
