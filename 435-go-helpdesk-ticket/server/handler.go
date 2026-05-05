package main

import (
	"encoding/json"
	"go-helpdesk-ticket/common"
	"net/http"
	"strings"
)

type Handler struct {
	service *TicketService
}

func NewHandler(service *TicketService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RespondJSON(w http.ResponseWriter, status int, data interface{}, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := common.APIResponse{
		Success: err == nil,
		Data:    data,
	}

	if err != nil {
		resp.Error = err.Error()
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondJSON(w, http.StatusMethodNotAllowed, nil, nil)
		return
	}

	var req common.CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, nil, err)
		return
	}

	ticket, err := h.service.CreateTicket(&req)
	if err != nil {
		h.RespondJSON(w, http.StatusBadRequest, nil, err)
		return
	}

	h.RespondJSON(w, http.StatusCreated, ticket, nil)
}

func (h *Handler) GetTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondJSON(w, http.StatusMethodNotAllowed, nil, nil)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/tickets/")
	if id == "" {
		h.RespondJSON(w, http.StatusBadRequest, nil, nil)
		return
	}

	ticket, err := h.service.GetTicket(id)
	if err != nil {
		h.RespondJSON(w, http.StatusNotFound, nil, err)
		return
	}

	h.RespondJSON(w, http.StatusOK, ticket, nil)
}

func (h *Handler) ListTickets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondJSON(w, http.StatusMethodNotAllowed, nil, nil)
		return
	}

	tickets := h.service.GetAllTickets()
	h.RespondJSON(w, http.StatusOK, tickets, nil)
}

func (h *Handler) ReassignTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondJSON(w, http.StatusMethodNotAllowed, nil, nil)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/tickets/")
	id = strings.TrimSuffix(id, "/reassign")
	if id == "" {
		h.RespondJSON(w, http.StatusBadRequest, nil, nil)
		return
	}

	var req common.ReassignTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, nil, err)
		return
	}

	if err := h.service.ReassignTicket(id, &req); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, nil, err)
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"}, nil)
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondJSON(w, http.StatusMethodNotAllowed, nil, nil)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/tickets/")
	id = strings.TrimSuffix(id, "/status")
	if id == "" {
		h.RespondJSON(w, http.StatusBadRequest, nil, nil)
		return
	}

	var req common.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, nil, err)
		return
	}

	if err := h.service.UpdateStatus(id, &req); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, nil, err)
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"}, nil)
}

func (h *Handler) RateTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondJSON(w, http.StatusMethodNotAllowed, nil, nil)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/tickets/")
	id = strings.TrimSuffix(id, "/rate")
	if id == "" {
		h.RespondJSON(w, http.StatusBadRequest, nil, nil)
		return
	}

	var req common.RateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, nil, err)
		return
	}

	if err := h.service.RateTicket(id, &req); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, nil, err)
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"}, nil)
}

func (h *Handler) LinkTickets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondJSON(w, http.StatusMethodNotAllowed, nil, nil)
		return
	}

	var req common.LinkTicketsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, nil, err)
		return
	}

	if err := h.service.LinkTickets(&req); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, nil, err)
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"}, nil)
}

func (h *Handler) TransferTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondJSON(w, http.StatusMethodNotAllowed, nil, nil)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/tickets/")
	id = strings.TrimSuffix(id, "/transfer")
	if id == "" {
		h.RespondJSON(w, http.StatusBadRequest, nil, nil)
		return
	}

	var req common.TransferTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, nil, err)
		return
	}

	if err := h.service.TransferTicket(id, &req); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, nil, err)
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"}, nil)
}

func (h *Handler) GetLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondJSON(w, http.StatusMethodNotAllowed, nil, nil)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/tickets/")
	id = strings.TrimSuffix(id, "/logs")
	if id == "" {
		h.RespondJSON(w, http.StatusBadRequest, nil, nil)
		return
	}

	logs := h.service.GetOperationLogs(id)
	h.RespondJSON(w, http.StatusOK, logs, nil)
}

func (h *Handler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondJSON(w, http.StatusMethodNotAllowed, nil, nil)
		return
	}

	stats := h.service.GetStatistics()
	h.RespondJSON(w, http.StatusOK, stats, nil)
}

func (h *Handler) GetHandlers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondJSON(w, http.StatusMethodNotAllowed, nil, nil)
		return
	}

	handlers := h.service.GetHandlers()
	h.RespondJSON(w, http.StatusOK, handlers, nil)
}

func (h *Handler) GetTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondJSON(w, http.StatusMethodNotAllowed, nil, nil)
		return
	}

	templates := h.service.GetTemplates()
	h.RespondJSON(w, http.StatusOK, templates, nil)
}

func (h *Handler) AutoClassify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondJSON(w, http.StatusMethodNotAllowed, nil, nil)
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, nil, err)
		return
	}

	category := h.service.AutoClassify(req.Title, req.Description)
	h.RespondJSON(w, http.StatusOK, map[string]string{"category": string(category)}, nil)
}
