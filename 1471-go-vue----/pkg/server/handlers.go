package server

import (
	"bus-station/internal/api"
	"bus-station/internal/core"
	"net/http"
	"strings"
)

func (s *Server) createScheduleHandler(w http.ResponseWriter, r *http.Request) {
	var req api.CreateScheduleRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.NewErrorResponse(400, "无效的请求体"))
		return
	}

	schedule, err := s.store.CreateSchedule(
		req.ScheduleNo,
		req.Date,
		req.DepartureStation,
		req.ArrivalStation,
		req.DepartureTime,
		req.ArrivalTime,
		req.BusType,
		req.Price,
		req.TotalSeats,
	)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.NewErrorResponse(400, err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, api.NewSuccessResponse(convertSchedule(schedule)))
}

func (s *Server) getScheduleHandler(w http.ResponseWriter, r *http.Request) {
	scheduleNo := r.URL.Query().Get("schedule_no")
	date := r.URL.Query().Get("date")

	if scheduleNo == "" || date == "" {
		writeJSON(w, http.StatusBadRequest, api.NewErrorResponse(400, "缺少 schedule_no 或 date 参数"))
		return
	}

	schedule, err := s.store.GetSchedule(scheduleNo, date)
	if err != nil {
		writeJSON(w, http.StatusNotFound, api.NewErrorResponse(404, err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, api.NewSuccessResponse(convertSchedule(schedule)))
}

func (s *Server) listSchedulesHandler(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		writeJSON(w, http.StatusBadRequest, api.NewErrorResponse(400, "缺少 date 参数"))
		return
	}

	schedules := s.store.ListSchedules(date)
	result := make([]*api.Schedule, len(schedules))
	for i, sch := range schedules {
		result[i] = convertSchedule(sch)
	}

	writeJSON(w, http.StatusOK, api.NewSuccessResponse(result))
}

func (s *Server) searchSchedulesHandler(w http.ResponseWriter, r *http.Request) {
	var req api.SearchSchedulesRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.NewErrorResponse(400, "无效的请求体"))
		return
	}

	schedules := s.store.SearchSchedules(req.Departure, req.Arrival, req.Date)
	result := make([]*api.Schedule, len(schedules))
	for i, sch := range schedules {
		result[i] = convertSchedule(sch)
	}

	writeJSON(w, http.StatusOK, api.NewSuccessResponse(result))
}

func (s *Server) updateScheduleStatusHandler(w http.ResponseWriter, r *http.Request) {
	var req api.UpdateScheduleStatusRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.NewErrorResponse(400, "无效的请求体"))
		return
	}

	status := core.ScheduleStatus(strings.ToLower(req.Status))
	err := s.store.UpdateScheduleStatus(req.ScheduleNo, req.Date, status)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.NewErrorResponse(400, err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, api.NewSuccessResponse(nil))
}

func (s *Server) cancelScheduleHandler(w http.ResponseWriter, r *http.Request) {
	var req api.CancelScheduleRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.NewErrorResponse(400, "无效的请求体"))
		return
	}

	tickets, err := s.store.CancelSchedule(req.ScheduleNo, req.Date)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.NewErrorResponse(400, err.Error()))
		return
	}

	result := make([]*api.Ticket, len(tickets))
	for i, ticket := range tickets {
		result[i] = convertTicket(ticket)
	}

	writeJSON(w, http.StatusOK, api.NewSuccessResponse(map[string]interface{}{
		"refund_tickets": result,
	}))
}

func (s *Server) getSeatsHandler(w http.ResponseWriter, r *http.Request) {
	scheduleNo := r.URL.Query().Get("schedule_no")
	date := r.URL.Query().Get("date")

	if scheduleNo == "" || date == "" {
		writeJSON(w, http.StatusBadRequest, api.NewErrorResponse(400, "缺少 schedule_no 或 date 参数"))
		return
	}

	seats, err := s.store.GetSeats(scheduleNo, date)
	if err != nil {
		writeJSON(w, http.StatusNotFound, api.NewErrorResponse(404, err.Error()))
		return
	}

	result := make([]*api.Seat, len(seats))
	for i, seat := range seats {
		result[i] = &api.Seat{
			SeatNo: seat.SeatNo,
			IsSold: seat.IsSold,
		}
	}

	writeJSON(w, http.StatusOK, api.NewSuccessResponse(result))
}

func (s *Server) purchaseTicketsHandler(w http.ResponseWriter, r *http.Request) {
	var req api.PurchaseTicketsRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.NewErrorResponse(400, "无效的请求体"))
		return
	}

	coreReq := &core.TicketPurchaseRequest{
		ScheduleNo:    req.ScheduleNo,
		Date:          req.Date,
		SeatNos:       req.SeatNos,
		PassengerName: req.PassengerName,
	}

	result, err := s.store.PurchaseTickets(coreReq)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.NewErrorResponse(400, err.Error()))
		return
	}

	tickets := make([]*api.Ticket, len(result.Tickets))
	for i, ticket := range result.Tickets {
		tickets[i] = convertTicket(ticket)
	}

	writeJSON(w, http.StatusOK, api.NewSuccessResponse(&api.PurchaseTicketsResponse{
		Tickets:     tickets,
		TotalAmount: result.TotalAmount,
	}))
}

func (s *Server) getTicketHandler(w http.ResponseWriter, r *http.Request) {
	ticketNo := r.URL.Query().Get("ticket_no")
	if ticketNo == "" {
		writeJSON(w, http.StatusBadRequest, api.NewErrorResponse(400, "缺少 ticket_no 参数"))
		return
	}

	ticket, err := s.store.GetTicket(ticketNo)
	if err != nil {
		writeJSON(w, http.StatusNotFound, api.NewErrorResponse(404, err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, api.NewSuccessResponse(convertTicket(ticket)))
}

func (s *Server) checkInHandler(w http.ResponseWriter, r *http.Request) {
	var req api.CheckInRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.NewErrorResponse(400, "无效的请求体"))
		return
	}

	record, err := s.store.CheckIn(req.TicketNo)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.NewErrorResponse(400, err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, api.NewSuccessResponse(map[string]interface{}{
		"check_in_id": record.ID,
		"ticket_no":   record.TicketNo,
		"checked_at":  record.CheckedAt,
	}))
}

func (s *Server) requestRefundHandler(w http.ResponseWriter, r *http.Request) {
	var req api.RefundRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.NewErrorResponse(400, "无效的请求体"))
		return
	}

	result, err := s.store.RequestRefund(req.TicketNo)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.NewErrorResponse(400, err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, api.NewSuccessResponse(&api.RefundResult{
		RequestID:    result.RequestID,
		TicketNo:     result.TicketNo,
		RefundAmount: result.RefundAmount,
		Status:       string(result.Status),
		Message:      result.Message,
	}))
}

func (s *Server) processRefundHandler(w http.ResponseWriter, r *http.Request) {
	var req api.ProcessRefundRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.NewErrorResponse(400, "无效的请求体"))
		return
	}

	result, err := s.store.ProcessRefund(req.RequestID, req.Approved, req.Reason)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.NewErrorResponse(400, err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, api.NewSuccessResponse(&api.RefundResult{
		RequestID:    result.RequestID,
		TicketNo:     result.TicketNo,
		RefundAmount: result.RefundAmount,
		Status:       string(result.Status),
		Message:      result.Message,
	}))
}

func (s *Server) listPendingRefundsHandler(w http.ResponseWriter, r *http.Request) {
	requests := s.store.ListPendingRefunds()

	type RefundReqDTO struct {
		ID           string `json:"id"`
		TicketNo     string `json:"ticket_no"`
		RefundAmount int    `json:"refund_amount"`
		RequestTime  string `json:"request_time"`
	}

	result := make([]*RefundReqDTO, len(requests))
	for i, req := range requests {
		result[i] = &RefundReqDTO{
			ID:           req.ID,
			TicketNo:     req.TicketNo,
			RefundAmount: req.RefundAmount,
			RequestTime:  req.RequestTime.Format("2006-01-02 15:04:05"),
		}
	}

	writeJSON(w, http.StatusOK, api.NewSuccessResponse(result))
}

func convertSchedule(s *core.Schedule) *api.Schedule {
	return &api.Schedule{
		ScheduleNo:       s.ScheduleNo,
		Date:             s.Date,
		DepartureStation: s.DepartureStation,
		ArrivalStation:   s.ArrivalStation,
		DepartureTime:    s.DepartureTime,
		ArrivalTime:      s.ArrivalTime,
		BusType:          s.BusType,
		Price:            s.Price,
		TotalSeats:       s.TotalSeats,
		SoldSeats:        s.SoldSeats,
		Status:           string(s.Status),
		IsAlmostSoldOut:  s.IsAlmostSoldOut,
	}
}

func convertTicket(t *core.Ticket) *api.Ticket {
	return &api.Ticket{
		TicketNo:      t.TicketNo,
		ScheduleNo:    t.ScheduleNo,
		Date:          t.Date,
		SeatNo:        t.SeatNo,
		PassengerName: t.PassengerName,
		Price:         t.Price,
		Status:        string(t.Status),
	}
}
