package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"moving-platform/pkg/common"
	"moving-platform/pkg/core"
)

type Server struct {
	manager *core.BookingManager
	port    int
}

func NewServer(manager *core.BookingManager, port int) *Server {
	return &Server{
		manager: manager,
		port:    port,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/estimate", s.handleEstimate)
	mux.HandleFunc("/api/bookings", s.handleBookings)
	mux.HandleFunc("/api/bookings/", s.handleBookingByID)
	mux.HandleFunc("/api/availability", s.handleAvailability)

	addr := fmt.Sprintf(":%d", s.port)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) handleEstimate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req common.EstimateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	estimate := core.EstimatePrice(req.Items, req.DistanceKM)
	writeJSON(w, http.StatusOK, common.EstimateResponse{Estimate: estimate})
}

func (s *Server) handleBookings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleCreateBooking(w, r)
	case http.MethodGet:
		s.handleListBookings(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *Server) handleCreateBooking(w http.ResponseWriter, r *http.Request) {
	var req common.CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	estimate := core.EstimatePrice(req.Items, req.DistanceKM)

	booking := &core.Booking{
		CustomerName: req.CustomerName,
		Phone:        req.Phone,
		MoveDate:     req.MoveDate,
		TimeSlot:     req.TimeSlot,
		FromAddress:  req.FromAddress,
		ToAddress:    req.ToAddress,
		DistanceKM:   req.DistanceKM,
		Items:        req.Items,
		Estimate:     estimate,
	}

	if err := s.manager.CreateBooking(booking); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, common.CreateBookingResponse{
		BookingID: booking.ID,
		Estimate:  booking.Estimate,
	})
}

func (s *Server) handleListBookings(w http.ResponseWriter, r *http.Request) {
	bookings := s.manager.ListBookings()
	writeJSON(w, http.StatusOK, common.ListBookingsResponse{Bookings: bookings})
}

func (s *Server) handleBookingByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/bookings/"):]
	if id == "" {
		writeError(w, http.StatusBadRequest, "Booking ID required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetBooking(w, r, id)
	case http.MethodDelete:
		s.handleCancelBooking(w, r, id)
	case http.MethodPatch:
		s.handleBookingAction(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *Server) handleGetBooking(w http.ResponseWriter, r *http.Request, id string) {
	booking, err := s.manager.GetBooking(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, common.GetBookingResponse{Booking: booking})
}

func (s *Server) handleCancelBooking(w http.ResponseWriter, r *http.Request, id string) {
	fee, err := s.manager.CancelBooking(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, common.CancelBookingResponse{CancellationFee: fee})
}

func (s *Server) handleBookingAction(w http.ResponseWriter, r *http.Request, id string) {
	action := r.URL.Query().Get("action")
	switch action {
	case "complete":
		s.handleCompleteBooking(w, r, id)
	case "settle":
		s.handleSettleBooking(w, r, id)
	case "review":
		s.handleAddReview(w, r, id)
	default:
		writeError(w, http.StatusBadRequest, "Invalid action")
	}
}

func (s *Server) handleCompleteBooking(w http.ResponseWriter, r *http.Request, id string) {
	if err := s.manager.CompleteBooking(id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleSettleBooking(w http.ResponseWriter, r *http.Request, id string) {
	var req common.SettleBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	amount, err := s.manager.SettleBooking(id, req.FinalItems, req.FinalDistanceKM)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, common.SettleBookingResponse{FinalAmount: amount})
}

func (s *Server) handleAddReview(w http.ResponseWriter, r *http.Request, id string) {
	var req common.AddReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	review := &core.Review{
		Rating:  req.Rating,
		Comment: req.Comment,
	}

	if err := s.manager.AddReview(id, review); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAvailability(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		writeError(w, http.StatusBadRequest, "Date parameter required")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid date format")
		return
	}

	slots := s.manager.GetAvailableSlots(date)
	writeJSON(w, http.StatusOK, common.CheckAvailabilityResponse{AvailableSlots: slots})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, common.ErrorResponse{Error: message})
}
