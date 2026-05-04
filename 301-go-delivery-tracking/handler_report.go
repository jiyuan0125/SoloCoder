package main

import (
	"encoding/json"
	"net/http"
)

type ReportHandler struct {
	store *DataStore
}

type ReportRequest struct {
	OrderID    string  `json:"order_id"`
	StatusName string  `json:"status_name"`
	Longitude  float64 `json:"longitude"`
	Latitude   float64 `json:"latitude"`
	Timestamp  int64   `json:"timestamp"`
}

type ReportResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

func NewReportHandler(store *DataStore) *ReportHandler {
	return &ReportHandler{store: store}
}

func (h *ReportHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, ReportResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if req.OrderID == "" {
		sendJSONResponse(w, http.StatusBadRequest, ReportResponse{
			Success: false,
			Message: "order_id is required",
		})
		return
	}

	if !IsValidStatusName(req.StatusName) {
		sendJSONResponse(w, http.StatusBadRequest, ReportResponse{
			Success: false,
			Message: "Invalid status_name: " + req.StatusName,
		})
		return
	}

	if !IsValidCoordinate(req.Longitude, req.Latitude) {
		sendJSONResponse(w, http.StatusBadRequest, ReportResponse{
			Success: false,
			Message: "Invalid coordinates",
		})
		return
	}

	if req.Timestamp <= 0 {
		sendJSONResponse(w, http.StatusBadRequest, ReportResponse{
			Success: false,
			Message: "timestamp is required",
		})
		return
	}

	status := DeliveryStatus{
		OrderID:    req.OrderID,
		StatusName: req.StatusName,
		Longitude:  req.Longitude,
		Latitude:   req.Latitude,
		Timestamp:  req.Timestamp,
	}

	added := h.store.AddStatus(status)

	if added {
		sendJSONResponse(w, http.StatusOK, ReportResponse{
			Success: true,
			Message: "Status reported",
		})
	} else {
		sendJSONResponse(w, http.StatusOK, ReportResponse{
			Success: true,
			Message: "Duplicate status ignored",
		})
	}
}

func sendJSONResponse(w http.ResponseWriter, statusCode int, response ReportResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
