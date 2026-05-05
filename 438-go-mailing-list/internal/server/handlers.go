package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"go-mailing-list/pkg/protocol"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func errorResponse(w http.ResponseWriter, status int, err error) {
	response := map[string]interface{}{
		"success": false,
		"error":   err.Error(),
	}
	jsonResponse(w, status, response)
}

func (h *Handler) CreateMailingList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	var req protocol.CreateMailingListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, err)
		return
	}

	list, err := h.service.CreateMailingList(req.Name, req.Description)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	jsonResponse(w, http.StatusCreated, protocol.CreateMailingListResponse{
		Success:     true,
		MailingList: list,
	})
}

func (h *Handler) GetMailingList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	listID := parts[len(parts)-1]

	list, err := h.service.GetMailingList(listID)
	if err != nil {
		errorResponse(w, http.StatusNotFound, err)
		return
	}

	jsonResponse(w, http.StatusOK, protocol.GetMailingListResponse{
		Success:     true,
		MailingList: list,
	})
}

func (h *Handler) ListMailingLists(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	lists := h.service.ListMailingLists()

	jsonResponse(w, http.StatusOK, protocol.ListMailingListsResponse{
		Success:      true,
		MailingLists: lists,
	})
}

func (h *Handler) PauseList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	var req protocol.PauseListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.PauseList(req.ListID)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	jsonResponse(w, http.StatusOK, protocol.PauseListResponse{
		Success: true,
	})
}

func (h *Handler) ResumeList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	var req protocol.ResumeListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.ResumeList(req.ListID)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	jsonResponse(w, http.StatusOK, protocol.ResumeListResponse{
		Success: true,
	})
}

func (h *Handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	listID := parts[len(parts)-2]

	var req protocol.SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, err)
		return
	}

	sub, err := h.service.Subscribe(listID, req.Email, req.Name, req.CustomFields)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	jsonResponse(w, http.StatusCreated, protocol.SubscribeResponse{
		Success:    true,
		Subscriber: sub,
	})
}

func (h *Handler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	listID := parts[len(parts)-2]

	var req protocol.UnsubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.Unsubscribe(listID, req.Email)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	jsonResponse(w, http.StatusOK, protocol.UnsubscribeResponse{
		Success: true,
	})
}

func (h *Handler) ListSubscribers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	listID := parts[len(parts)-2]

	subscribers, err := h.service.ListSubscribers(listID)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	jsonResponse(w, http.StatusOK, protocol.ListSubscribersResponse{
		Success:     true,
		Subscribers: subscribers,
	})
}

func (h *Handler) ImportSubscribers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	listID := parts[len(parts)-2]

	var req protocol.ImportSubscribersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, err)
		return
	}

	total, imported, skipped, err := h.service.ImportSubscribers(listID, req.FilePath)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	jsonResponse(w, http.StatusOK, protocol.ImportSubscribersResponse{
		Success:  true,
		Total:    total,
		Imported: imported,
		Skipped:  skipped,
	})
}

func (h *Handler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	var req protocol.CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, err)
		return
	}

	template, err := h.service.CreateTemplate(
		req.Name, req.Subject, req.HTMLBody, req.TextBody,
		req.TrackOpens, req.TrackClicks, req.Variables,
	)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	jsonResponse(w, http.StatusCreated, protocol.CreateTemplateResponse{
		Success:  true,
		Template: template,
	})
}

func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	templates := h.service.ListTemplates()

	jsonResponse(w, http.StatusOK, protocol.ListTemplatesResponse{
		Success:   true,
		Templates: templates,
	})
}

func (h *Handler) SendCampaign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	var req protocol.SendCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, err)
		return
	}

	task, err := h.service.SendCampaign(&req)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	jsonResponse(w, http.StatusCreated, protocol.SendCampaignResponse{
		Success: true,
		Task:    task,
	})
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	taskID := parts[len(parts)-1]

	task, err := h.service.GetTask(taskID)
	if err != nil {
		errorResponse(w, http.StatusNotFound, err)
		return
	}

	jsonResponse(w, http.StatusOK, protocol.GetTaskResponse{
		Success: true,
		Task:    task,
	})
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	tasks := h.service.ListTasks()

	jsonResponse(w, http.StatusOK, protocol.ListTasksResponse{
		Success: true,
		Tasks:   tasks,
	})
}

func (h *Handler) GetSendRecords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	taskID := parts[len(parts)-1]

	records := h.service.GetSendRecords(taskID)

	jsonResponse(w, http.StatusOK, protocol.GetSendRecordsResponse{
		Success:     true,
		SendRecords: records,
	})
}

func (h *Handler) TrackOpen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	var req protocol.TrackOpenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.TrackOpen(req.TaskID, req.Email)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	jsonResponse(w, http.StatusOK, protocol.TrackResponse{
		Success: true,
	})
}

func (h *Handler) TrackClick(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	var req protocol.TrackClickRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.TrackClick(req.TaskID, req.Email, req.URL)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}

	jsonResponse(w, http.StatusOK, protocol.TrackResponse{
		Success: true,
	})
}

func (h *Handler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	alerts := h.service.GetAlerts()

	jsonResponse(w, http.StatusOK, protocol.GetAlertsResponse{
		Success: true,
		Alerts:  alerts,
	})
}

func (h *Handler) ResolveAlert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	var req protocol.ResolveAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, err)
		return
	}

	success := h.service.ResolveAlert(req.AlertID)
	if !success {
		errorResponse(w, http.StatusNotFound, ErrListNotFound)
		return
	}

	jsonResponse(w, http.StatusOK, protocol.ResolveAlertResponse{
		Success: true,
	})
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorResponse(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	totalLists, totalSubs, totalTasks, totalSent, totalBounced, activeAlerts := h.service.GetStats()

	response := protocol.GetStatsResponse{}
	response.Success = true
	response.Stats.TotalLists = totalLists
	response.Stats.TotalSubscribers = totalSubs
	response.Stats.TotalTasks = totalTasks
	response.Stats.TotalSent = totalSent
	response.Stats.TotalBounced = totalBounced
	response.Stats.ActiveAlerts = activeAlerts

	jsonResponse(w, http.StatusOK, response)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case path == "/api/lists" && r.Method == http.MethodGet:
		h.ListMailingLists(w, r)
	case path == "/api/lists" && r.Method == http.MethodPost:
		h.CreateMailingList(w, r)
	case strings.HasPrefix(path, "/api/lists/") && strings.HasSuffix(path, "/subscribers") && r.Method == http.MethodGet:
		h.ListSubscribers(w, r)
	case strings.HasPrefix(path, "/api/lists/") && strings.HasSuffix(path, "/subscribe") && r.Method == http.MethodPost:
		h.Subscribe(w, r)
	case strings.HasPrefix(path, "/api/lists/") && strings.HasSuffix(path, "/unsubscribe") && r.Method == http.MethodPost:
		h.Unsubscribe(w, r)
	case strings.HasPrefix(path, "/api/lists/") && strings.HasSuffix(path, "/import") && r.Method == http.MethodPost:
		h.ImportSubscribers(w, r)
	case path == "/api/lists/pause" && r.Method == http.MethodPost:
		h.PauseList(w, r)
	case path == "/api/lists/resume" && r.Method == http.MethodPost:
		h.ResumeList(w, r)
	case strings.HasPrefix(path, "/api/lists/") && r.Method == http.MethodGet:
		h.GetMailingList(w, r)
	case path == "/api/templates" && r.Method == http.MethodGet:
		h.ListTemplates(w, r)
	case path == "/api/templates" && r.Method == http.MethodPost:
		h.CreateTemplate(w, r)
	case path == "/api/campaigns" && r.Method == http.MethodPost:
		h.SendCampaign(w, r)
	case path == "/api/tasks" && r.Method == http.MethodGet:
		h.ListTasks(w, r)
	case strings.HasPrefix(path, "/api/tasks/") && strings.HasSuffix(path, "/records") && r.Method == http.MethodGet:
		h.GetSendRecords(w, r)
	case strings.HasPrefix(path, "/api/tasks/") && r.Method == http.MethodGet:
		h.GetTask(w, r)
	case path == "/api/track/open" && r.Method == http.MethodPost:
		h.TrackOpen(w, r)
	case path == "/api/track/click" && r.Method == http.MethodPost:
		h.TrackClick(w, r)
	case path == "/api/alerts" && r.Method == http.MethodGet:
		h.GetAlerts(w, r)
	case path == "/api/alerts/resolve" && r.Method == http.MethodPost:
		h.ResolveAlert(w, r)
	case path == "/api/stats" && r.Method == http.MethodGet:
		h.GetStats(w, r)
	default:
		http.NotFound(w, r)
	}
}
