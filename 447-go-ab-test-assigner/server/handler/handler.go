package handler

import (
	"abtest/api"
	"abtest/server/service"
	"encoding/json"
	"net/http"
	"strings"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateExperiment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.CreateExperimentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		h.sendError(w, "experiment name is required", http.StatusBadRequest)
		return
	}

	if req.Type != api.ExperimentTypeNormal && req.Type != api.ExperimentTypeGray {
		h.sendError(w, "invalid experiment type, must be 'normal' or 'gray'", http.StatusBadRequest)
		return
	}

	exp, err := h.svc.CreateExperiment(req)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.sendJSON(w, http.StatusCreated, api.CreateExperimentResponse{
		Success:    true,
		Experiment: *exp,
	})
}

func (h *Handler) StartExperiment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	experimentID := h.getPathParam(r.URL.Path, "/experiments/start/")
	if experimentID == "" {
		var req api.StartExperimentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.sendError(w, "experiment_id is required", http.StatusBadRequest)
			return
		}
		experimentID = req.ExperimentID
	}

	if experimentID == "" {
		h.sendError(w, "experiment_id is required", http.StatusBadRequest)
		return
	}

	exp, err := h.svc.StartExperiment(experimentID)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.sendJSON(w, http.StatusOK, api.StartExperimentResponse{
		Success:    true,
		Experiment: *exp,
	})
}

func (h *Handler) EndExperiment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	experimentID := h.getPathParam(r.URL.Path, "/experiments/end/")
	if experimentID == "" {
		var req api.EndExperimentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.sendError(w, "experiment_id is required", http.StatusBadRequest)
			return
		}
		experimentID = req.ExperimentID
	}

	if experimentID == "" {
		h.sendError(w, "experiment_id is required", http.StatusBadRequest)
		return
	}

	exp, err := h.svc.EndExperiment(experimentID)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.sendJSON(w, http.StatusOK, api.EndExperimentResponse{
		Success:    true,
		Experiment: *exp,
	})
}

func (h *Handler) ArchiveExperiment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	experimentID := h.getPathParam(r.URL.Path, "/experiments/archive/")
	if experimentID == "" {
		var req api.ArchiveExperimentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.sendError(w, "experiment_id is required", http.StatusBadRequest)
			return
		}
		experimentID = req.ExperimentID
	}

	if experimentID == "" {
		h.sendError(w, "experiment_id is required", http.StatusBadRequest)
		return
	}

	exp, err := h.svc.ArchiveExperiment(experimentID)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.sendJSON(w, http.StatusOK, api.ArchiveExperimentResponse{
		Success:    true,
		Experiment: *exp,
	})
}

func (h *Handler) UpdateTraffic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.UpdateTrafficRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.ExperimentID == "" {
		h.sendError(w, "experiment_id is required", http.StatusBadRequest)
		return
	}

	exp, err := h.svc.UpdateTraffic(req.ExperimentID, req.ControlPercentage, req.TreatmentPercentage)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.sendJSON(w, http.StatusOK, api.UpdateTrafficResponse{
		Success:    true,
		Experiment: *exp,
	})
}

func (h *Handler) AssignUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.AssignUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.ExperimentID == "" || req.UserID == "" {
		h.sendError(w, "experiment_id and user_id are required", http.StatusBadRequest)
		return
	}

	group, err := h.svc.AssignUser(req.ExperimentID, req.UserID)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.sendJSON(w, http.StatusOK, api.AssignUserResponse{
		Success:      true,
		Group:        group,
		ExperimentID: req.ExperimentID,
		UserID:       req.UserID,
	})
}

func (h *Handler) RecordMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.RecordMetricsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.ExperimentID == "" || req.UserID == "" {
		h.sendError(w, "experiment_id and user_id are required", http.StatusBadRequest)
		return
	}

	if err := h.svc.RecordMetrics(req.ExperimentID, req.UserID, req.Converted, req.StayDuration); err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.sendJSON(w, http.StatusOK, api.RecordMetricsResponse{
		Success: true,
		Message: "metrics recorded",
	})
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	experimentID := h.getPathParam(r.URL.Path, "/experiments/stats/")
	if experimentID == "" {
		experimentID = r.URL.Query().Get("experiment_id")
	}

	if experimentID == "" {
		h.sendError(w, "experiment_id is required", http.StatusBadRequest)
		return
	}

	stats, err := h.svc.GetStats(experimentID)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.sendJSON(w, http.StatusOK, stats)
}

func (h *Handler) GetUserAssignments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := h.getPathParam(r.URL.Path, "/users/assignments/")
	if userID == "" {
		userID = r.URL.Query().Get("user_id")
	}

	if userID == "" {
		h.sendError(w, "user_id is required", http.StatusBadRequest)
		return
	}

	assignments, err := h.svc.GetUserAssignments(userID)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.sendJSON(w, http.StatusOK, api.GetUserAssignmentsResponse{
		Success:     true,
		UserID:      userID,
		Assignments: assignments,
	})
}

func (h *Handler) ListExperiments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var status *api.ExperimentStatus
	statusStr := r.URL.Query().Get("status")
	if statusStr != "" {
		s := api.ExperimentStatus(statusStr)
		status = &s
	}

	experiments := h.svc.ListExperiments(status)

	h.sendJSON(w, http.StatusOK, api.ListExperimentsResponse{
		Success:     true,
		Experiments: experiments,
	})
}

func (h *Handler) GetExperiment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	experimentID := h.getPathParam(r.URL.Path, "/experiments/")
	if experimentID == "" {
		experimentID = r.URL.Query().Get("experiment_id")
	}

	if experimentID == "" {
		h.sendError(w, "experiment_id is required", http.StatusBadRequest)
		return
	}

	exp, exists := h.svc.GetExperiment(experimentID)
	if !exists {
		h.sendError(w, "experiment not found", http.StatusNotFound)
		return
	}

	h.sendJSON(w, http.StatusOK, api.GetExperimentResponse{
		Success:    true,
		Experiment: *exp,
	})
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.sendJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (h *Handler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) sendError(w http.ResponseWriter, message string, status int) {
	h.sendJSON(w, status, map[string]interface{}{
		"success": false,
		"message": message,
	})
}

func (h *Handler) getPathParam(path, prefix string) string {
	if strings.HasPrefix(path, prefix) {
		param := strings.TrimPrefix(path, prefix)
		if param != "" && !strings.Contains(param, "/") {
			return param
		}
	}
	return ""
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.HealthCheck)

	mux.HandleFunc("/experiments", h.ListExperiments)
	mux.HandleFunc("/experiments/create", h.CreateExperiment)
	mux.HandleFunc("/experiments/start/", h.StartExperiment)
	mux.HandleFunc("/experiments/end/", h.EndExperiment)
	mux.HandleFunc("/experiments/archive/", h.ArchiveExperiment)
	mux.HandleFunc("/experiments/traffic", h.UpdateTraffic)
	mux.HandleFunc("/experiments/stats/", h.GetStats)
	mux.HandleFunc("/experiments/", h.GetExperiment)

	mux.HandleFunc("/assign", h.AssignUser)
	mux.HandleFunc("/metrics", h.RecordMetrics)

	mux.HandleFunc("/users/assignments/", h.GetUserAssignments)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			h.sendJSON(w, http.StatusOK, map[string]string{
				"message": "AB Test API",
				"version": "1.0",
			})
			return
		}
		http.NotFound(w, r)
	})
}
