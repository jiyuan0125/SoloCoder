package core

import (
	"time"
)

type TicketPurchaseRequest struct {
	ScheduleNo    string
	Date          string
	SeatNos       []int
	PassengerName string
}

type PurchaseResult struct {
	Tickets     []*Ticket
	TotalAmount int
}

func (s *Store) PurchaseTickets(req *TicketPurchaseRequest) (*PurchaseResult, error) {
	if len(req.SeatNos) <= 0 || len(req.SeatNos) > 5 {
		return nil, ErrInvalidSeatCount
	}
	
	key := s.GetScheduleKey(req.ScheduleNo, req.Date)
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	schedule, exists := s.schedules[key]
	if !exists {
		return nil, ErrScheduleNotFound
	}
	
	if schedule.Status == ScheduleStatusCancelled {
		return nil, ErrScheduleCancelled
	}
	
	seats, exists := s.scheduleSeats[key]
	if !exists {
		return nil, ErrScheduleNotFound
	}
	
	for _, seatNo := range req.SeatNos {
		if seatNo <= 0 || seatNo > schedule.TotalSeats {
			return nil, ErrSeatOutOfRange
		}
	}
	
	for _, seatNo := range req.SeatNos {
		if seats[seatNo-1].IsSold {
			return nil, ErrSeatAlreadySold
		}
	}
	
	createdTickets := make([]*Ticket, 0, len(req.SeatNos))
	
	for _, seatNo := range req.SeatNos {
		seat := seats[seatNo-1]
		ticketNo := s.generateTicketNo()
		
		ticket := &Ticket{
			TicketNo:      ticketNo,
			ScheduleNo:    req.ScheduleNo,
			Date:          req.Date,
			SeatNo:        seatNo,
			PassengerName: req.PassengerName,
			Price:         schedule.Price,
			Status:        TicketStatusSold,
			SoldAt:        time.Now(),
		}
		
		s.tickets[ticketNo] = ticket
		seat.IsSold = true
		seat.TicketID = ticketNo
		schedule.SoldSeats++
		createdTickets = append(createdTickets, ticket)
	}
	
	s.updateAlmostSoldOut(schedule)
	
	schedule.UpdatedAt = time.Now()
	
	totalAmount := schedule.Price * len(req.SeatNos)
	
	return &PurchaseResult{
		Tickets:     createdTickets,
		TotalAmount: totalAmount,
	}, nil
}

func (s *Store) GetTicket(ticketNo string) (*Ticket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	ticket, exists := s.tickets[ticketNo]
	if !exists {
		return nil, ErrTicketNotFound
	}
	
	return ticket, nil
}

func (s *Store) GetTicketsBySchedule(scheduleNo, date string) []*Ticket {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := make([]*Ticket, 0)
	for _, ticket := range s.tickets {
		if ticket.ScheduleNo == scheduleNo && ticket.Date == date {
			result = append(result, ticket)
		}
	}
	
	return result
}

func (s *Store) updateAlmostSoldOut(schedule *Schedule) {
	if schedule.TotalSeats == 0 {
		schedule.IsAlmostSoldOut = false
		return
	}
	
	ratio := float64(schedule.SoldSeats) / float64(schedule.TotalSeats)
	schedule.IsAlmostSoldOut = ratio >= 0.9
}

func (s *Store) GetNotBoardedTickets(scheduleNo, date string) []*Ticket {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := make([]*Ticket, 0)
	for _, ticket := range s.tickets {
		if ticket.ScheduleNo == scheduleNo &&
			ticket.Date == date &&
			ticket.Status == TicketStatusSold {
			result = append(result, ticket)
		}
	}
	
	return result
}
