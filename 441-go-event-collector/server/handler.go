package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"event-collector/common"
)

type Handler struct {
	storage      *MemoryStorage
	validator    *Validator
	deduplicator *Deduplicator
	reporter     *DailyReporter
	aggregator   *Aggregator
}

func NewHandler(storage *MemoryStorage, validator *Validator, deduplicator *Deduplicator, reporter *DailyReporter, aggregator *Aggregator) *Handler {
	return &Handler{
		storage:      storage,
		validator:    validator,
		deduplicator: deduplicator,
		reporter:     reporter,
		aggregator:   aggregator,
	}
}

func (h *Handler) Track(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, common.ErrCodeInvalidRequest, "Method not allowed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, common.ErrCodeInvalidRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	var req common.TrackRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, common.ErrCodeInvalidRequest, "Invalid JSON format")
		return
	}

	if err := h.validator.ValidateEvent(&req.Event); err != nil {
		if apiErr, ok := err.(*common.APIError); ok {
			h.writeError(w, http.StatusBadRequest, apiErr.Code, apiErr.Message)
		} else {
			h.writeError(w, http.StatusBadRequest, common.ErrCodeInvalidRequest, err.Error())
		}
		return
	}

	dedupeKey := h.deduplicator.GetKey(req.Event.Name, req.Event.DeviceID)
	eventTime := time.Unix(req.Event.Timestamp, 0)
	if h.deduplicator.IsDuplicate(dedupeKey, eventTime) {
		h.writeError(w, http.StatusTooManyRequests, common.ErrCodeDuplicateEvent, "Duplicate event")
		return
	}

	event := req.Event
	h.storage.Store(&event)
	h.deduplicator.Record(dedupeKey, eventTime)

	h.writeSuccess(w, "Event tracked successfully")
}

func (h *Handler) Query(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, common.ErrCodeInvalidRequest, "Method not allowed")
		return
	}

	req := common.QueryRequest{}
	req.EventName = r.URL.Query().Get("event_name")
	req.UserID = r.URL.Query().Get("user_id")

	if startStr := r.URL.Query().Get("start_time"); startStr != "" {
		startTime, err := strconv.ParseInt(startStr, 10, 64)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, common.ErrCodeInvalidRequest, "Invalid start_time")
			return
		}
		req.StartTime = startTime
	}

	if endStr := r.URL.Query().Get("end_time"); endStr != "" {
		endTime, err := strconv.ParseInt(endStr, 10, 64)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, common.ErrCodeInvalidRequest, "Invalid end_time")
			return
		}
		req.EndTime = endTime
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, common.ErrCodeInvalidRequest, "Invalid page")
			return
		}
		req.Page = page
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, common.ErrCodeInvalidRequest, "Invalid page_size")
			return
		}
		req.PageSize = pageSize
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > common.MaxPageSize {
		req.PageSize = common.DefaultPageSize
	}

	events, total := h.storage.Query(req.EventName, req.UserID, req.StartTime, req.EndTime, req.Page, req.PageSize)

	resp := common.QueryResponse{
		Success: true,
		Data:    events,
		Total:   total,
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) Aggregate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, common.ErrCodeInvalidRequest, "Method not allowed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, common.ErrCodeInvalidRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	var req common.AggregationRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, common.ErrCodeInvalidRequest, "Invalid JSON format")
		return
	}

	if len(req.EventNames) == 0 {
		h.writeError(w, http.StatusBadRequest, common.ErrCodeInvalidRequest, "At least one event name required")
		return
	}

	result, err := h.aggregator.Aggregate(req.EventNames, req.StartTime, req.EndTime)
	if err != nil {
		if apiErr, ok := err.(*common.APIError); ok {
			h.writeError(w, http.StatusBadRequest, apiErr.Code, apiErr.Message)
		} else {
			h.writeError(w, http.StatusInternalServerError, common.ErrCodeInternalError, err.Error())
		}
		return
	}

	resp := common.AggregationResponse{
		Success: true,
		Data:    *result,
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetReports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, common.ErrCodeInvalidRequest, "Method not allowed")
		return
	}

	reports := h.reporter.GetReports()

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    reports,
	})
}

func (h *Handler) writeError(w http.ResponseWriter, status int, code common.ErrorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	})
}

func (h *Handler) writeSuccess(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": message,
	})
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
