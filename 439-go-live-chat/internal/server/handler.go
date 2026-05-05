package server

import (
	"encoding/json"
	"go-live-chat/common"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, common.ErrorResponse{
		Error: message,
	})
}

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var request common.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if request.UserID == "" {
		h.respondError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	response, err := h.service.CreateSession(request.UserID, request.Username)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var request common.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if request.SessionID == "" || request.SenderID == "" || request.Content == "" {
		h.respondError(w, http.StatusBadRequest, "session_id, sender_id and content are required")
		return
	}
	response, err := h.service.SendMessage(request)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) CloseSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var request common.CloseSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if request.SessionID == "" {
		h.respondError(w, http.StatusBadRequest, "session_id is required")
		return
	}
	response, err := h.service.CloseSession(request)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) SubmitSatisfaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var request common.SubmitSatisfactionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if request.SessionID == "" || request.UserID == "" {
		h.respondError(w, http.StatusBadRequest, "session_id and user_id are required")
		return
	}
	response, err := h.service.SubmitSatisfaction(request)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) TransferSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var request common.TransferSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if request.SessionID == "" || request.FromAgentID == "" || request.ToAgentID == "" {
		h.respondError(w, http.StatusBadRequest, "session_id, from_agent_id and to_agent_id are required")
		return
	}
	response, err := h.service.TransferSession(request)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) CreateQuickReply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var request common.CreateQuickReplyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if request.AgentID == "" || request.Title == "" || request.Content == "" {
		h.respondError(w, http.StatusBadRequest, "agent_id, title and content are required")
		return
	}
	response, err := h.service.CreateQuickReply(request)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) GetQuickReplies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		h.respondError(w, http.StatusBadRequest, "agent_id is required")
		return
	}
	response, err := h.service.GetQuickReplies(agentID)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) AddToBlacklist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var request common.AddToBlacklistRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if request.UserID == "" {
		h.respondError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	response, err := h.service.AddToBlacklist(request)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) RemoveFromBlacklist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.respondError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	response, err := h.service.RemoveFromBlacklist(userID)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) AgentLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var request common.AgentLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if request.AgentID == "" {
		h.respondError(w, http.StatusBadRequest, "agent_id is required")
		return
	}
	response, err := h.service.AgentLogin(request)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) AgentLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var request common.AgentLogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if request.AgentID == "" {
		h.respondError(w, http.StatusBadRequest, "agent_id is required")
		return
	}
	response, err := h.service.AgentLogout(request.AgentID)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) GetAgentSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		h.respondError(w, http.StatusBadRequest, "agent_id is required")
		return
	}
	response, err := h.service.GetAgentSessions(agentID)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) GetSessionHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		h.respondError(w, http.StatusBadRequest, "session_id is required")
		return
	}
	response, err := h.service.GetSessionHistory(sessionID)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	response, err := h.service.GetStatistics()
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) AddSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var request common.AddScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if request.AgentID == "" {
		h.respondError(w, http.StatusBadRequest, "agent_id is required")
		return
	}
	if request.StartTime.IsZero() || request.EndTime.IsZero() {
		h.respondError(w, http.StatusBadRequest, "start_time and end_time are required")
		return
	}
	response, err := h.service.AddSchedule(request)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) GetSchedules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		h.respondError(w, http.StatusBadRequest, "agent_id is required")
		return
	}
	response, err := h.service.GetSchedules(agentID)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/sessions", h.CreateSession)
	mux.HandleFunc("/api/sessions/close", h.CloseSession)
	mux.HandleFunc("/api/sessions/transfer", h.TransferSession)
	mux.HandleFunc("/api/sessions/history", h.GetSessionHistory)
	mux.HandleFunc("/api/messages", h.SendMessage)
	mux.HandleFunc("/api/satisfaction", h.SubmitSatisfaction)
	mux.HandleFunc("/api/quick-replies", h.GetQuickReplies)
	mux.HandleFunc("/api/quick-replies/create", h.CreateQuickReply)
	mux.HandleFunc("/api/blacklist", h.AddToBlacklist)
	mux.HandleFunc("/api/blacklist/remove", h.RemoveFromBlacklist)
	mux.HandleFunc("/api/agents/login", h.AgentLogin)
	mux.HandleFunc("/api/agents/logout", h.AgentLogout)
	mux.HandleFunc("/api/agents/sessions", h.GetAgentSessions)
	mux.HandleFunc("/api/agents/schedules", h.GetSchedules)
	mux.HandleFunc("/api/agents/schedules/create", h.AddSchedule)
	mux.HandleFunc("/api/statistics", h.GetStatistics)
}
