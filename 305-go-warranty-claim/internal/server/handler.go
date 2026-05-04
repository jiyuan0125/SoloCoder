package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"warranty-claim/pkg/common"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) SubmitApplication(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.SubmitApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp := h.service.SubmitApplication(req)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetApplication(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/applications/")
	if id == "" {
		http.Error(w, "Application ID required", http.StatusBadRequest)
		return
	}

	app, exists := h.service.GetApplication(id)
	if !exists {
		http.Error(w, "Application not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(app)
}

func (h *Handler) GetUserApplications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	apps := h.service.GetUserApplications(userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.GetApplicationsResponse{
		Success:      true,
		Applications: apps,
	})
}

func (h *Handler) GetAllApplications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apps := h.service.GetAllApplications()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.GetApplicationsResponse{
		Success:      true,
		Applications: apps,
	})
}

func (h *Handler) GetPendingApplications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apps := h.service.GetPendingApplications()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.GetApplicationsResponse{
		Success:      true,
		Applications: apps,
	})
}

func (h *Handler) ReviewApplication(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.ReviewApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	success, message := h.service.ReviewApplication(req)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": success,
		"message": message,
	})
}

func (h *Handler) SubmitAppeal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.SubmitAppealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "X-User-ID header required", http.StatusBadRequest)
		return
	}

	success, message := h.service.SubmitAppeal(req, userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": success,
		"message": message,
	})
}

func (h *Handler) GetAppeals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	appID := r.URL.Query().Get("application_id")
	if appID == "" {
		appeals := h.service.GetPendingAppeals()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(common.GetAppealsResponse{
			Success: true,
			Appeals: appeals,
		})
		return
	}

	appeals := h.service.GetAppealsByApplication(appID)
	var withApp []common.AppealWithApplication
	if app, exists := h.service.GetApplication(appID); exists {
		for _, a := range appeals {
			withApp = append(withApp, common.AppealWithApplication{
				Appeal:      a,
				Application: app,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.GetAppealsResponse{
		Success: true,
		Appeals: withApp,
	})
}

func (h *Handler) ResolveAppeal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AppealID   string `json:"appeal_id"`
		Resolved   bool   `json:"resolved"`
		AdminNote  string `json:"admin_note"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	success, message := h.service.ResolveAppeal(req.AppealID, req.Resolved, req.AdminNote)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": success,
		"message": message,
	})
}

func (h *Handler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := h.service.GetStatistics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *Handler) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/submit", h.SubmitApplication)
	mux.HandleFunc("/applications/", h.GetApplication)
	mux.HandleFunc("/my-applications", h.GetUserApplications)
	mux.HandleFunc("/all-applications", h.GetAllApplications)
	mux.HandleFunc("/pending-applications", h.GetPendingApplications)
	mux.HandleFunc("/review", h.ReviewApplication)
	mux.HandleFunc("/appeal", h.SubmitAppeal)
	mux.HandleFunc("/appeals", h.GetAppeals)
	mux.HandleFunc("/resolve-appeal", h.ResolveAppeal)
	mux.HandleFunc("/statistics", h.GetStatistics)
}
