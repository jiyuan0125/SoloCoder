package handler

import (
	"datacleanser/internal/common"
	"datacleanser/internal/server/service"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Handler struct {
	service *service.CleanerService
}

func NewHandler(s *service.CleanerService) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/clean", h.handleClean)
	mux.HandleFunc("/api/v1/templates", h.handleTemplates)
	mux.HandleFunc("/api/v1/templates/", h.handleTemplate)
	mux.HandleFunc("/api/v1/tasks/", h.handleTask)
	mux.HandleFunc("/api/v1/errors", h.handleErrors)
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func errorResponse(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]interface{}{
		"success": false,
		"message": message,
	})
}

func (h *Handler) handleClean(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.CleanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Records) == 0 {
		errorResponse(w, http.StatusBadRequest, "no records provided")
		return
	}

	if len(req.Records) > common.MaxBatchSize {
		errorResponse(w, http.StatusBadRequest, fmt.Sprintf("batch size exceeds limit: %d (max %d)", len(req.Records), common.MaxBatchSize))
		return
	}

	var rule common.CleanRule
	if req.TemplateName != "" {
		template, exists := h.service.GetTemplate(req.TemplateName)
		if !exists {
			errorResponse(w, http.StatusBadRequest, fmt.Sprintf("template not found: %s", req.TemplateName))
			return
		}
		rule = template.Rule
	} else if req.Rule != nil {
		rule = *req.Rule
	} else {
		errorResponse(w, http.StatusBadRequest, "either rule or template_name must be provided")
		return
	}

	if req.Async {
		taskID, err := h.service.CleanAsync(req)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusAccepted, common.CleanResponse{
			Success: true,
			TaskID:  taskID,
		})
		return
	}

	result, err := h.service.Clean(req.Records, rule)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, common.CleanResponse{
		Success: true,
		Data:    result,
	})
}

func (h *Handler) handleTemplates(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		templates := h.service.ListTemplates()
		jsonResponse(w, http.StatusOK, common.ListTemplatesResponse{
			Success: true,
			Data:    templates,
		})
	case http.MethodPost:
		var req common.CreateTemplateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if strings.TrimSpace(req.Name) == "" {
			errorResponse(w, http.StatusBadRequest, "template name is required")
			return
		}

		template, err := h.service.CreateTemplate(req.Name, req.Rule)
		if err != nil {
			errorResponse(w, http.StatusConflict, err.Error())
			return
		}

		jsonResponse(w, http.StatusCreated, common.TemplateResponse{
			Success: true,
			Data:    template,
		})
	default:
		errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) handleTemplate(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/templates/")
	parts := strings.SplitN(path, "/", 2)
	name := parts[0]

	if name == "" {
		errorResponse(w, http.StatusBadRequest, "template name is required")
		return
	}

	if len(parts) == 2 && parts[1] == "rollback" {
		h.handleTemplateRollback(w, r, name)
		return
	}

	switch r.Method {
	case http.MethodGet:
		template, exists := h.service.GetTemplate(name)
		if !exists {
			errorResponse(w, http.StatusNotFound, "template not found")
			return
		}
		jsonResponse(w, http.StatusOK, common.TemplateResponse{
			Success: true,
			Data:    template,
		})
	case http.MethodPut:
		var req common.UpdateTemplateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorResponse(w, http.StatusBadRequest, "invalid request body")
			return
		}

		template, err := h.service.UpdateTemplate(name, req.Rule)
		if err != nil {
			errorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		jsonResponse(w, http.StatusOK, common.TemplateResponse{
			Success: true,
			Data:    template,
		})
	default:
		errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) handleTemplateRollback(w http.ResponseWriter, r *http.Request, name string) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Version int `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	template, err := h.service.RollbackTemplate(name, req.Version)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, common.TemplateResponse{
		Success: true,
		Data:    template,
	})
}

func (h *Handler) handleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	taskID := strings.TrimPrefix(r.URL.Path, "/api/v1/tasks/")
	if taskID == "" {
		errorResponse(w, http.StatusBadRequest, "task id is required")
		return
	}

	task, exists := h.service.GetTaskStatus(taskID)
	if !exists {
		errorResponse(w, http.StatusNotFound, "task not found")
		return
	}

	jsonResponse(w, http.StatusOK, common.TaskResponse{
		Success: true,
		Data:    task,
	})
}

func (h *Handler) handleErrors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	samples := h.service.GetErrorSamples()
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    samples,
	})
}
