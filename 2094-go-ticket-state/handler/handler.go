package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ticket-system/database"
	"ticket-system/models"
	"ticket-system/service"
)

type Handler struct {
	service *service.TicketService
	db      *database.DB
}

func NewHandler(s *service.TicketService, db *database.DB) *Handler {
	return &Handler{service: s, db: db}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, err error, extra ...interface{}) {
	resp := models.ErrorResponse{Error: err.Error()}
	if len(extra) > 0 {
		if allowed, ok := extra[0].([]string); ok {
			resp.AllowedStatuses = allowed
		} else if assignee, ok := extra[0].(string); ok {
			resp.CurrentAssignee = assignee
		}
	}
	respondJSON(w, status, resp)
}

func parseTicketID(r *http.Request) (int64, error) {
	parts := strings.Split(r.URL.Path, "/")
	for i, p := range parts {
		if i > 0 && (parts[i-1] == "tickets" || parts[i-1] == "ticket") {
			id, err := strconv.ParseInt(p, 10, 64)
			if err == nil {
				return id, nil
			}
		}
	}
	return 0, errors.New("invalid ticket id")
}

func (h *Handler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}
	defer r.Body.Close()

	if strings.TrimSpace(req.Title) == "" {
		respondError(w, http.StatusBadRequest, errors.New("title is required"))
		return
	}
	if strings.TrimSpace(req.CreatorID) == "" {
		respondError(w, http.StatusBadRequest, errors.New("creator_id is required"))
		return
	}

	ticket, err := h.service.CreateTicket(r.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPriority):
			respondError(w, http.StatusBadRequest, err)
		case errors.Is(err, service.ErrPriorityNotSpecified):
			respondError(w, http.StatusBadRequest, err)
		case errors.Is(err, service.ErrTicketNotFound):
			respondError(w, http.StatusBadRequest, errors.New("parent ticket not found"))
		default:
			respondError(w, http.StatusInternalServerError, err)
		}
		return
	}

	respondJSON(w, http.StatusCreated, ticket)
}

func (h *Handler) GetTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := parseTicketID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	ticket, err := h.service.GetTicket(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrTicketNotFound) {
			respondError(w, http.StatusNotFound, err)
		} else {
			respondError(w, http.StatusInternalServerError, err)
		}
		return
	}

	respondJSON(w, http.StatusOK, ticket)
}

func (h *Handler) ListTickets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	params := &database.QueryTicketsParams{}

	if v := query.Get("priority"); v != "" {
		v = strings.ToLower(strings.TrimSpace(v))
		if !service.IsValidPriority(v) {
			respondError(w, http.StatusBadRequest, errors.New("invalid priority"))
			return
		}
		params.Priority = &v
	}
	if v := query.Get("status"); v != "" {
		v = strings.ToLower(strings.TrimSpace(v))
		params.Status = &v
	}
	if v := query.Get("assignee"); v != "" {
		params.AssigneeID = &v
	}
	if v := query.Get("creator"); v != "" {
		params.CreatorID = &v
	}
	if v := query.Get("created_from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			respondError(w, http.StatusBadRequest, errors.New("invalid created_from format, use RFC3339"))
			return
		}
		params.CreatedFrom = &t
	}
	if v := query.Get("created_to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			respondError(w, http.StatusBadRequest, errors.New("invalid created_to format, use RFC3339"))
			return
		}
		params.CreatedTo = &t
	}

	tickets, err := h.service.QueryTickets(r.Context(), params)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}

	respondJSON(w, http.StatusOK, tickets)
}

func (h *Handler) AssignTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := parseTicketID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	var req models.AssignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}
	defer r.Body.Close()

	if strings.TrimSpace(req.UserID) == "" {
		respondError(w, http.StatusBadRequest, errors.New("user_id is required"))
		return
	}

	err = h.service.AssignTicket(r.Context(), id, req.UserID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTicketNotFound):
			respondError(w, http.StatusNotFound, err)
		case errors.Is(err, service.ErrTicketClosed):
			respondError(w, http.StatusBadRequest, err)
		default:
			var assignErr *service.AlreadyAssignedError
			if errors.As(err, &assignErr) {
				respondError(w, http.StatusConflict, err, assignErr.Assignee)
			} else {
				respondError(w, http.StatusInternalServerError, err)
			}
		}
		return
	}

	ticket, _ := h.service.GetTicket(r.Context(), id)
	respondJSON(w, http.StatusOK, ticket)
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := parseTicketID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	var req models.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}
	defer r.Body.Close()

	if strings.TrimSpace(req.Status) == "" {
		respondError(w, http.StatusBadRequest, errors.New("status is required"))
		return
	}
	if strings.TrimSpace(req.UserID) == "" {
		respondError(w, http.StatusBadRequest, errors.New("user_id is required"))
		return
	}

	err = h.service.UpdateStatus(r.Context(), id, req.Status, req.UserID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTicketNotFound):
			respondError(w, http.StatusNotFound, err)
		case errors.Is(err, service.ErrTicketClosed):
			respondError(w, http.StatusBadRequest, err)
		default:
			var transErr *service.InvalidTransitionError
			if errors.As(err, &transErr) {
				respondError(w, http.StatusBadRequest, err, transErr.AllowedStatuses)
			} else {
				respondError(w, http.StatusInternalServerError, err)
			}
		}
		return
	}

	ticket, _ := h.service.GetTicket(r.Context(), id)
	respondJSON(w, http.StatusOK, ticket)
}

func (h *Handler) GetTicketHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := parseTicketID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	history, err := h.service.GetTicketHistory(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrTicketNotFound) {
			respondError(w, http.StatusNotFound, err)
		} else {
			respondError(w, http.StatusInternalServerError, err)
		}
		return
	}

	respondJSON(w, http.StatusOK, history)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		respondJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unhealthy", "error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}
