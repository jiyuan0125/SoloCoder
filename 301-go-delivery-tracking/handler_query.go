package main

import (
	"encoding/json"
	"net/http"
)

type QueryHandler struct {
	store *DataStore
}

type QueryResponse struct {
	OrderID  string           `json:"order_id"`
	Statuses []StatusResponse `json:"statuses"`
}

type StatusResponse struct {
	StatusName string  `json:"status_name"`
	Longitude  float64 `json:"longitude"`
	Latitude   float64 `json:"latitude"`
	Timestamp  int64   `json:"timestamp"`
}

func NewQueryHandler(store *DataStore) *QueryHandler {
	return &QueryHandler{store: store}
}

func (h *QueryHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orderID := r.URL.Query().Get("order_id")
	if orderID == "" {
		sendQueryError(w, http.StatusBadRequest, "order_id is required")
		return
	}

	statuses := h.store.GetStatuses(orderID)
	if statuses == nil {
		sendQueryError(w, http.StatusNotFound, "Order not found")
		return
	}

	response := QueryResponse{
		OrderID:  orderID,
		Statuses: make([]StatusResponse, len(statuses)),
	}

	for i, s := range statuses {
		response.Statuses[i] = StatusResponse{
			StatusName: s.StatusName,
			Longitude:  s.Longitude,
			Latitude:   s.Latitude,
			Timestamp:  s.Timestamp,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func sendQueryError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	type ErrorResponse struct {
		Error string `json:"error"`
	}
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
