package core

import (
	"time"
)

func (s *Store) AddScheduleTemplate(tpl *ScheduleTemplate) error {
	if err := ValidateTime(tpl.DepartureTime); err != nil {
		return err
	}
	if err := ValidateTime(tpl.ArrivalTime); err != nil {
		return err
	}
	if err := ValidateDepartureArrival(tpl.DepartureTime, tpl.ArrivalTime); err != nil {
		return err
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scheduleTemplates = append(s.scheduleTemplates, tpl)
	return nil
}

func (s *Store) CreateScheduleFromTemplate(tpl *ScheduleTemplate, date string) error {
	key := s.GetScheduleKey(tpl.ScheduleNo, date)
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.schedules[key]; exists {
		return ErrScheduleAlreadyExist
	}
	
	schedule := &Schedule{
		ScheduleNo:       tpl.ScheduleNo,
		Date:             date,
		DepartureStation: tpl.DepartureStation,
		ArrivalStation:   tpl.ArrivalStation,
		DepartureTime:    tpl.DepartureTime,
		ArrivalTime:      tpl.ArrivalTime,
		BusType:          tpl.BusType,
		Price:            tpl.Price,
		TotalSeats:       tpl.TotalSeats,
		SoldSeats:        0,
		Status:           ScheduleStatusNotDeparted,
		IsAlmostSoldOut:  false,
		Archived:         false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	
	s.schedules[key] = schedule
	
	seats := make([]*Seat, tpl.TotalSeats)
	for i := range seats {
		seats[i] = &Seat{
			SeatNo: i + 1,
			IsSold: false,
		}
	}
	s.scheduleSeats[key] = seats
	
	return nil
}

func (s *Store) CreateSchedule(scheduleNo, date, departureStation, arrivalStation, departureTime, arrivalTime, busType string, price, totalSeats int) (*Schedule, error) {
	if err := ValidateTime(departureTime); err != nil {
		return nil, err
	}
	if err := ValidateTime(arrivalTime); err != nil {
		return nil, err
	}
	if err := ValidateDepartureArrival(departureTime, arrivalTime); err != nil {
		return nil, err
	}
	
	key := s.GetScheduleKey(scheduleNo, date)
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.schedules[key]; exists {
		return nil, ErrScheduleAlreadyExist
	}
	
	schedule := &Schedule{
		ScheduleNo:       scheduleNo,
		Date:             date,
		DepartureStation: departureStation,
		ArrivalStation:   arrivalStation,
		DepartureTime:    departureTime,
		ArrivalTime:      arrivalTime,
		BusType:          busType,
		Price:            price,
		TotalSeats:       totalSeats,
		SoldSeats:        0,
		Status:           ScheduleStatusNotDeparted,
		IsAlmostSoldOut:  false,
		Archived:         false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	
	s.schedules[key] = schedule
	
	seats := make([]*Seat, totalSeats)
	for i := range seats {
		seats[i] = &Seat{
			SeatNo: i + 1,
			IsSold: false,
		}
	}
	s.scheduleSeats[key] = seats
	
	return schedule, nil
}

func (s *Store) GetSchedule(scheduleNo, date string) (*Schedule, error) {
	key := s.GetScheduleKey(scheduleNo, date)
	
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	schedule, exists := s.schedules[key]
	if !exists {
		return nil, ErrScheduleNotFound
	}
	
	return schedule, nil
}

func (s *Store) ListSchedules(date string) []*Schedule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := make([]*Schedule, 0)
	for _, schedule := range s.schedules {
		if schedule.Date == date && !schedule.Archived {
			result = append(result, schedule)
		}
	}
	return result
}

func (s *Store) SearchSchedules(departure, arrival, date string) []*Schedule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := make([]*Schedule, 0)
	for _, schedule := range s.schedules {
		if schedule.DepartureStation == departure &&
			schedule.ArrivalStation == arrival &&
			schedule.Date == date &&
			schedule.Status != ScheduleStatusCancelled &&
			!schedule.Archived {
			result = append(result, schedule)
		}
	}
	return result
}

func (s *Store) UpdateScheduleStatus(scheduleNo, date string, status ScheduleStatus) error {
	key := s.GetScheduleKey(scheduleNo, date)
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	schedule, exists := s.schedules[key]
	if !exists {
		return ErrScheduleNotFound
	}
	
	if schedule.Status == ScheduleStatusCancelled && status != ScheduleStatusCancelled {
		return ErrScheduleNotModifiable
	}
	
	schedule.Status = status
	schedule.UpdatedAt = time.Now()
	return nil
}

func (s *Store) CancelSchedule(scheduleNo, date string) ([]*Ticket, error) {
	key := s.GetScheduleKey(scheduleNo, date)
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	schedule, exists := s.schedules[key]
	if !exists {
		return nil, ErrScheduleNotFound
	}
	
	if schedule.Status == ScheduleStatusCancelled {
		return nil, nil
	}
	
	schedule.Status = ScheduleStatusCancelled
	schedule.UpdatedAt = time.Now()
	
	ticketsToRefund := make([]*Ticket, 0)
	for _, ticket := range s.tickets {
		if ticket.ScheduleNo == scheduleNo &&
			ticket.Date == date &&
			ticket.Status == TicketStatusSold {
			ticketsToRefund = append(ticketsToRefund, ticket)
		}
	}
	
	for _, ticket := range ticketsToRefund {
		refundReq := &RefundRequest{
			ID:           s.generateRefundID(),
			TicketNo:     ticket.TicketNo,
			Status:       RefundStatusPending,
			RefundAmount: ticket.Price,
			RequestTime:  time.Now(),
		}
		s.refundRequests[refundReq.ID] = refundReq
	}
	
	return ticketsToRefund, nil
}

func (s *Store) GetAvailableSeats(scheduleNo, date string) ([]int, error) {
	key := s.GetScheduleKey(scheduleNo, date)
	
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	seats, exists := s.scheduleSeats[key]
	if !exists {
		return nil, ErrScheduleNotFound
	}
	
	available := make([]int, 0)
	for _, seat := range seats {
		if !seat.IsSold {
			available = append(available, seat.SeatNo)
		}
	}
	
	return available, nil
}

func (s *Store) GetSeats(scheduleNo, date string) ([]*Seat, error) {
	key := s.GetScheduleKey(scheduleNo, date)
	
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	seats, exists := s.scheduleSeats[key]
	if !exists {
		return nil, ErrScheduleNotFound
	}
	
	result := make([]*Seat, len(seats))
	for i, seat := range seats {
		result[i] = &Seat{
			SeatNo: seat.SeatNo,
			IsSold: seat.IsSold,
		}
	}
	
	return result, nil
}
