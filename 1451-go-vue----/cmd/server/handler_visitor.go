package main

import (
	"encoding/json"
	"net/http"

	"smart-park/common"
)

func (h *Handler) handleVisitorReserve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Success: false, Message: "method not allowed"})
		return
	}

	var req common.CreateVisitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: "invalid request body"})
		return
	}

	reservation, err := h.visitorService.CreateReservation(
		req.EmployeeID,
		req.VisitorName,
		req.VisitorPhone,
		req.Purpose,
		req.ExpectedArrival,
	)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, common.Response{Success: true, Data: reservation})
}

func (h *Handler) handleVisitorReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Success: false, Message: "method not allowed"})
		return
	}

	var req common.ReviewVisitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: "invalid request body"})
		return
	}

	reservation, err := h.visitorService.ReviewReservation(req.ReservationID, req.Approve)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, common.Response{Success: true, Data: reservation})
}

func (h *Handler) handleVisitorCheckIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Success: false, Message: "method not allowed"})
		return
	}

	var req common.CheckInVisitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: "invalid request body"})
		return
	}

	reservation, err := h.visitorService.CheckInVisitor(req.VisitorPhone)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, common.Response{Success: true, Data: reservation})
}

func (h *Handler) handleVisitorList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Success: false, Message: "method not allowed"})
		return
	}
	reservations := h.visitorService.GetAllReservations()
	writeJSON(w, http.StatusOK, common.Response{Success: true, Data: reservations})
}
