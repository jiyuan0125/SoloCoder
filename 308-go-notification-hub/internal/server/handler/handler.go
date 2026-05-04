package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"notification-hub/internal/server/service"
	"notification-hub/pkg/models"
)

type Handler struct {
	service *service.NotificationService
}

func NewHandler(s *service.NotificationService) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/notifications", h.CreateNotification)
	mux.HandleFunc("GET /api/statistics", h.GetStatistics)
	mux.HandleFunc("GET /api/users/{user_id}/notifications", h.GetUserNotifications)
	mux.HandleFunc("PUT /api/users/{user_id}/notifications/{delivery_id}/read", h.MarkAsRead)
}

func (h *Handler) CreateNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.service.CreateNotification(&req)
	if err != nil {
		h.writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.writeJSON(w, resp, http.StatusCreated)
}

func (h *Handler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := h.service.GetStatistics()
	resp := models.StatisticsResponse{
		Statistics: stats,
	}

	h.writeJSON(w, resp, http.StatusOK)
}

func (h *Handler) GetUserNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.PathValue("user_id")
	if userID == "" {
		h.writeError(w, "user_id is required", http.StatusBadRequest)
		return
	}

	var isRead *bool
	isReadStr := r.URL.Query().Get("is_read")
	if isReadStr != "" {
		val, err := strconv.ParseBool(isReadStr)
		if err != nil {
			h.writeError(w, "invalid is_read parameter", http.StatusBadRequest)
			return
		}
		isRead = &val
	}

	notifications, err := h.service.GetUserNotifications(userID, isRead)
	if err != nil {
		h.writeError(w, err.Error(), http.StatusNotFound)
		return
	}

	resp := models.UserListNotificationsResponse{
		Notifications: notifications,
	}

	h.writeJSON(w, resp, http.StatusOK)
}

func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.PathValue("user_id")
	deliveryID := r.PathValue("delivery_id")

	if userID == "" || deliveryID == "" {
		h.writeError(w, "user_id and delivery_id are required", http.StatusBadRequest)
		return
	}

	err := h.service.MarkAsRead(userID, deliveryID)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		} else if strings.Contains(err.Error(), "only insite") || strings.Contains(err.Error(), "unauthorized") {
			status = http.StatusBadRequest
		}
		h.writeError(w, err.Error(), status)
		return
	}

	resp := models.MarkAsReadResponse{
		Success: true,
	}

	h.writeJSON(w, resp, http.StatusOK)
}

func (h *Handler) writeJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.ErrorResponse{
		Error: message,
	})
}
