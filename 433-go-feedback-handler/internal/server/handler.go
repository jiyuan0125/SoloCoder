package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-feedback-handler/pkg/protocol"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/feedbacks", h.createFeedback)
	mux.HandleFunc("GET /api/v1/feedbacks", h.listFeedbacks)
	mux.HandleFunc("GET /api/v1/feedbacks/{id}", h.getFeedback)
	mux.HandleFunc("PUT /api/v1/feedbacks/{id}/status", h.updateStatus)
	mux.HandleFunc("PUT /api/v1/feedbacks/{id}/assign", h.assignHandler)

	mux.HandleFunc("POST /api/v1/comments", h.addComment)

	mux.HandleFunc("GET /api/v1/tags", h.listTags)
	mux.HandleFunc("POST /api/v1/tags", h.createTag)
	mux.HandleFunc("POST /api/v1/feedbacks/{id}/tags", h.addTagToFeedback)
	mux.HandleFunc("DELETE /api/v1/feedbacks/{id}/tags/{tagId}", h.removeTagFromFeedback)

	mux.HandleFunc("GET /api/v1/notifications", h.getNotifications)
	mux.HandleFunc("PUT /api/v1/notifications/{id}/read", h.markNotificationRead)

	mux.HandleFunc("GET /api/v1/users/{id}/limit", h.getUserLimit)

	mux.HandleFunc("GET /api/v1/kpi/{handlerId}", h.getKPIStats)

	mux.HandleFunc("GET /api/v1/reports/monthly", h.getMonthlyReport)
	mux.HandleFunc("POST /api/v1/reports/monthly/generate", h.generateMonthlyReport)

	mux.HandleFunc("GET /health", h.healthCheck)
}

func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, protocol.SuccessResponse{Success: true})
}

func (h *Handler) createFeedback(w http.ResponseWriter, r *http.Request) {
	var req protocol.CreateFeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UserID == "" || req.UserName == "" || req.Content == "" {
		errorResponse(w, http.StatusBadRequest, "user_id, user_name and content are required")
		return
	}

	validTypes := []protocol.FeedbackType{
		protocol.FeedbackTypeBug,
		protocol.FeedbackTypeFeature,
		protocol.FeedbackTypeComplaint,
		protocol.FeedbackTypeConsultation,
	}
	typeValid := false
	for _, t := range validTypes {
		if req.Type == t {
			typeValid = true
			break
		}
	}
	if !typeValid {
		errorResponse(w, http.StatusBadRequest, "invalid feedback type")
		return
	}

	resp, err := h.service.CreateFeedback(&req)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, resp)
}

func (h *Handler) getFeedback(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		errorResponse(w, http.StatusBadRequest, "feedback id is required")
		return
	}

	fb, err := h.service.GetFeedback(id)
	if err != nil {
		if err == ErrFeedbackNotFound {
			errorResponse(w, http.StatusNotFound, "feedback not found")
			return
		}
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, fb)
}

func (h *Handler) listFeedbacks(w http.ResponseWriter, r *http.Request) {
	req := protocol.ListFeedbackRequest{
		Status:   protocol.FeedbackStatus(r.URL.Query().Get("status")),
		Type:     protocol.FeedbackType(r.URL.Query().Get("type")),
		Priority: protocol.FeedbackPriority(r.URL.Query().Get("priority")),
		UserID:   r.URL.Query().Get("user_id"),
		HandlerID: r.URL.Query().Get("handler_id"),
		TagID:    r.URL.Query().Get("tag_id"),
		Page:     parseIntOrDefault(r.URL.Query().Get("page"), protocol.DefaultPage),
		PageSize: parseIntOrDefault(r.URL.Query().Get("page_size"), protocol.DefaultPageSize),
	}

	if inReview := r.URL.Query().Get("in_review"); inReview != "" {
		val := inReview == "true"
		req.InReview = &val
	}

	resp, err := h.service.ListFeedbacks(&req)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		errorResponse(w, http.StatusBadRequest, "feedback id is required")
		return
	}

	var req protocol.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.FeedbackID = id

	if req.Status == "" {
		errorResponse(w, http.StatusBadRequest, "status is required")
		return
	}

	if err := h.service.UpdateStatus(&req); err != nil {
		if err == ErrFeedbackNotFound {
			errorResponse(w, http.StatusNotFound, "feedback not found")
			return
		}
		if err == ErrInvalidStatusTransition {
			errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		if err == ErrFeedbackMerged {
			errorResponse(w, http.StatusBadRequest, "feedback has been merged")
			return
		}
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, protocol.SuccessResponse{Success: true})
}

func (h *Handler) assignHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		errorResponse(w, http.StatusBadRequest, "feedback id is required")
		return
	}

	var req protocol.AssignHandlerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.FeedbackID = id

	if req.HandlerID == "" {
		errorResponse(w, http.StatusBadRequest, "handler_id is required")
		return
	}

	if err := h.service.AssignHandler(&req); err != nil {
		if err == ErrFeedbackNotFound {
			errorResponse(w, http.StatusNotFound, "feedback not found")
			return
		}
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, protocol.SuccessResponse{Success: true})
}

func (h *Handler) addComment(w http.ResponseWriter, r *http.Request) {
	var req protocol.AddCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.FeedbackID == "" || req.UserID == "" || req.Content == "" {
		errorResponse(w, http.StatusBadRequest, "feedback_id, user_id and content are required")
		return
	}

	if err := h.service.AddComment(&req); err != nil {
		if err == ErrFeedbackNotFound {
			errorResponse(w, http.StatusNotFound, "feedback not found")
			return
		}
		if err == ErrFeedbackMerged {
			errorResponse(w, http.StatusBadRequest, "feedback has been merged")
			return
		}
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, protocol.SuccessResponse{Success: true})
}

func (h *Handler) listTags(w http.ResponseWriter, r *http.Request) {
	tags := h.service.ListTags()
	jsonResponse(w, http.StatusOK, tags)
}

func (h *Handler) createTag(w http.ResponseWriter, r *http.Request) {
	var req protocol.CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		errorResponse(w, http.StatusBadRequest, "tag name is required")
		return
	}

	tag, err := h.service.CreateTag(&req)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, tag)
}

func (h *Handler) addTagToFeedback(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		errorResponse(w, http.StatusBadRequest, "feedback id is required")
		return
	}

	var req protocol.AddTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.FeedbackID = id

	if req.TagID == "" {
		errorResponse(w, http.StatusBadRequest, "tag_id is required")
		return
	}

	if err := h.service.AddTagToFeedback(&req); err != nil {
		if err == ErrFeedbackNotFound {
			errorResponse(w, http.StatusNotFound, "feedback not found")
			return
		}
		if err == ErrTagNotFound {
			errorResponse(w, http.StatusNotFound, "tag not found")
			return
		}
		if err == ErrFeedbackMerged {
			errorResponse(w, http.StatusBadRequest, "feedback has been merged")
			return
		}
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, protocol.SuccessResponse{Success: true})
}

func (h *Handler) removeTagFromFeedback(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tagId := r.PathValue("tagId")

	if id == "" || tagId == "" {
		errorResponse(w, http.StatusBadRequest, "feedback id and tag id are required")
		return
	}

	req := protocol.RemoveTagRequest{
		FeedbackID: id,
		TagID:      tagId,
	}

	if err := h.service.RemoveTagFromFeedback(&req); err != nil {
		if err == ErrFeedbackNotFound {
			errorResponse(w, http.StatusNotFound, "feedback not found")
			return
		}
		if err == ErrFeedbackMerged {
			errorResponse(w, http.StatusBadRequest, "feedback has been merged")
			return
		}
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, protocol.SuccessResponse{Success: true})
}

func (h *Handler) getNotifications(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		errorResponse(w, http.StatusBadRequest, "user_id is required")
		return
	}

	notifications := h.service.GetNotifications(userID)
	jsonResponse(w, http.StatusOK, notifications)
}

func (h *Handler) markNotificationRead(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		errorResponse(w, http.StatusBadRequest, "notification id is required")
		return
	}

	h.service.MarkNotificationRead(id)
	jsonResponse(w, http.StatusOK, protocol.SuccessResponse{Success: true})
}

func (h *Handler) getUserLimit(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		errorResponse(w, http.StatusBadRequest, "user id is required")
		return
	}

	limit := h.service.GetUserLimit(id)
	jsonResponse(w, http.StatusOK, limit)
}

func (h *Handler) getKPIStats(w http.ResponseWriter, r *http.Request) {
	handlerID := r.PathValue("handlerId")
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "month"
	}

	if handlerID == "" {
		errorResponse(w, http.StatusBadRequest, "handler id is required")
		return
	}

	stats := h.service.GetKPIStats(handlerID, period)
	jsonResponse(w, http.StatusOK, stats)
}

func (h *Handler) getMonthlyReport(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	if month == "" {
		month = getMonthKey(now())
	}

	report, err := h.service.GenerateMonthlyReport(month)
	if err != nil {
		errorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, report)
}

func (h *Handler) generateMonthlyReport(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	if month == "" {
		month = getMonthKey(now())
	}

	report, err := h.service.GenerateMonthlyReport(month)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, report)
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func errorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(protocol.ErrorResponse{
		Error: message,
		Code:  status,
	})
}

func parseIntOrDefault(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}

func now() time.Time {
	return time.Now()
}

func parseBool(s string) (bool, error) {
	s = strings.ToLower(s)
	if s == "true" || s == "1" || s == "yes" {
		return true, nil
	}
	if s == "false" || s == "0" || s == "no" {
		return false, nil
	}
	return false, nil
}
