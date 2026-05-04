package handler

import (
	"billing-service/service"
	"encoding/json"
	"net/http"
	"strconv"
)

type UsageHandler struct {
	usageService *service.UsageService
}

func NewUsageHandler(usageService *service.UsageService) *UsageHandler {
	return &UsageHandler{usageService: usageService}
}

type RecordSmsRequest struct {
	CustomerID uint `json:"customer_id"`
	SmsCount   int  `json:"sms_count"`
}

type RecordStorageRequest struct {
	CustomerID uint    `json:"customer_id"`
	StorageGB  float64 `json:"storage_gb"`
}

type GetUsageRequest struct {
	CustomerID uint `json:"customer_id"`
	Year       int  `json:"year"`
	Month      int  `json:"month"`
}

func (h *UsageHandler) HandleUsages(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getUsage(w, r)
	case http.MethodPost:
		h.recordUsage(w, r)
	default:
		RespondError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *UsageHandler) getUsage(w http.ResponseWriter, r *http.Request) {
	customerIDStr := r.URL.Query().Get("customer_id")
	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")

	if customerIDStr == "" || yearStr == "" || monthStr == "" {
		RespondError(w, http.StatusBadRequest, "customer_id, year, and month are required")
		return
	}

	customerID, err := strconv.ParseUint(customerIDStr, 10, 64)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "invalid customer_id")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "invalid year")
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "invalid month")
		return
	}

	usage, err := h.usageService.GetUsage(uint(customerID), year, month)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondSuccess(w, usage)
}

func (h *UsageHandler) recordUsage(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("action")

	switch action {
	case "sms":
		h.recordSmsUsage(w, r)
	case "storage":
		h.recordStorageUsage(w, r)
	default:
		RespondError(w, http.StatusBadRequest, "invalid action, use 'sms' or 'storage'")
	}
}

func (h *UsageHandler) recordSmsUsage(w http.ResponseWriter, r *http.Request) {
	var req RecordSmsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.CustomerID == 0 {
		RespondError(w, http.StatusBadRequest, "customer_id is required")
		return
	}

	if req.SmsCount <= 0 {
		RespondError(w, http.StatusBadRequest, "sms_count must be positive")
		return
	}

	if err := h.usageService.RecordSmsUsage(req.CustomerID, req.SmsCount); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondSuccess(w, map[string]string{"message": "sms usage recorded successfully"})
}

func (h *UsageHandler) recordStorageUsage(w http.ResponseWriter, r *http.Request) {
	var req RecordStorageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.CustomerID == 0 {
		RespondError(w, http.StatusBadRequest, "customer_id is required")
		return
	}

	if req.StorageGB < 0 {
		RespondError(w, http.StatusBadRequest, "storage_gb must be non-negative")
		return
	}

	if err := h.usageService.RecordStorageUsage(req.CustomerID, req.StorageGB); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondSuccess(w, map[string]string{"message": "storage usage recorded successfully"})
}
