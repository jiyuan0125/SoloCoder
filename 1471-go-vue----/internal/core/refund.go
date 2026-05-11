package core

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

type RefundResult struct {
	RequestID    string
	TicketNo     string
	RefundAmount int
	Status       RefundStatus
	Message      string
}

func parseDateString(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

func parseTimeString(timeStr string) (int, int, error) {
	if err := ValidateTime(timeStr); err != nil {
		return 0, 0, err
	}
	hour, _ := strconv.Atoi(timeStr[:2])
	min, _ := strconv.Atoi(timeStr[3:])
	return hour, min, nil
}

func getDepartureDateTime(dateStr, timeStr string) (time.Time, error) {
	date, err := parseDateString(dateStr)
	if err != nil {
		return time.Time{}, err
	}
	hour, min, err := parseTimeString(timeStr)
	if err != nil {
		return time.Time{}, err
	}
	
	return time.Date(date.Year(), date.Month(), date.Day(), hour, min, 0, 0, time.Local), nil
}

func calculateRefundAmount(price int, departureTime time.Time, requestTime time.Time) (int, string, error) {
	diff := departureTime.Sub(requestTime)
	
	hours2 := 2 * time.Hour
	minutes30 := 30 * time.Minute
	
	if diff >= hours2 {
		return price, "全额退款", nil
	} else if diff >= minutes30 {
		fee := int(float64(price) * 0.2)
		refund := price - fee
		return refund, fmt.Sprintf("扣20%%手续费，退款%d元", refund), nil
	} else {
		return 0, "发车前30分钟内或发车后不允许退票", errors.New("不允许退票")
	}
}

func (s *Store) RequestRefund(ticketNo string) (*RefundResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	ticket, exists := s.tickets[ticketNo]
	if !exists {
		return nil, ErrTicketNotFound
	}
	
	if ticket.Status == TicketStatusRefunded {
		return nil, ErrAlreadyRefunded
	}
	
	if ticket.Status == TicketStatusChecked {
		return nil, errors.New("已检票的车票不能退票")
	}
	
	if ticket.Status == TicketStatusNotBoarded {
		return nil, errors.New("未乘车的车票不能退票")
	}
	
	key := s.GetScheduleKey(ticket.ScheduleNo, ticket.Date)
	schedule, exists := s.schedules[key]
	if !exists {
		return nil, ErrScheduleNotFound
	}
	
	departureTime, err := getDepartureDateTime(ticket.Date, schedule.DepartureTime)
	if err != nil {
		return nil, err
	}
	
	requestTime := time.Now()
	refundAmount, message, calcErr := calculateRefundAmount(ticket.Price, departureTime, requestTime)
	if calcErr != nil {
		return nil, calcErr
	}
	
	refundReq := &RefundRequest{
		ID:           s.generateRefundID(),
		TicketNo:     ticketNo,
		Status:       RefundStatusPending,
		RefundAmount: refundAmount,
		RequestTime:  requestTime,
	}
	s.refundRequests[refundReq.ID] = refundReq
	
	return &RefundResult{
		RequestID:    refundReq.ID,
		TicketNo:     ticketNo,
		RefundAmount: refundAmount,
		Status:       RefundStatusPending,
		Message:      message,
	}, nil
}

func (s *Store) ProcessRefund(requestID string, approved bool, reason string) (*RefundResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	refundReq, exists := s.refundRequests[requestID]
	if !exists {
		return nil, errors.New("退款请求不存在")
	}
	
	if refundReq.Status != RefundStatusPending {
		return nil, errors.New("退款请求已处理")
	}
	
	ticket, exists := s.tickets[refundReq.TicketNo]
	if !exists {
		return nil, ErrTicketNotFound
	}
	
	now := time.Now()
	refundReq.ProcessedAt = &now
	
	if approved {
		refundReq.Status = RefundStatusSuccess
		ticket.Status = TicketStatusRefunded
		ticket.CancelledRefundAt = &now
		
		key := s.GetScheduleKey(ticket.ScheduleNo, ticket.Date)
		if seats, seatsExist := s.scheduleSeats[key]; seatsExist {
			seat := seats[ticket.SeatNo-1]
			seat.IsSold = false
			seat.TicketID = ""
		}
		
		if schedule, schExist := s.schedules[key]; schExist {
			if schedule.SoldSeats > 0 {
				schedule.SoldSeats--
				s.updateAlmostSoldOut(schedule)
			}
			schedule.UpdatedAt = now
		}
		
		return &RefundResult{
			RequestID:    requestID,
			TicketNo:     refundReq.TicketNo,
			RefundAmount: refundReq.RefundAmount,
			Status:       RefundStatusSuccess,
			Message:      "退款成功",
		}, nil
	} else {
		refundReq.Status = RefundStatusReject
		refundReq.RejectReason = reason
		
		return &RefundResult{
			RequestID:    requestID,
			TicketNo:     refundReq.TicketNo,
			RefundAmount: 0,
			Status:       RefundStatusReject,
			Message:      fmt.Sprintf("退款被拒绝：%s", reason),
		}, nil
	}
}

func (s *Store) GetRefundRequest(requestID string) (*RefundRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	req, exists := s.refundRequests[requestID]
	if !exists {
		return nil, errors.New("退款请求不存在")
	}
	
	return req, nil
}

func (s *Store) ListPendingRefunds() []*RefundRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := make([]*RefundRequest, 0)
	for _, req := range s.refundRequests {
		if req.Status == RefundStatusPending {
			result = append(result, req)
		}
	}
	return result
}
