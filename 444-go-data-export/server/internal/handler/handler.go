package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"data-export/pkg/common"
	"data-export/server/internal/data"
	"data-export/server/internal/stats"
	"data-export/server/internal/task"
	"data-export/server/internal/template"
)

type Handler struct {
	templates *template.TemplateManager
	scheduler *task.TaskScheduler
	stats     *stats.StatsCollector
	dataSource *data.DataSource
}

func NewHandler(templates *template.TemplateManager, scheduler *task.TaskScheduler,
	stats *stats.StatsCollector, dataSource *data.DataSource) *Handler {
	return &Handler{
		templates:  templates,
		scheduler:  scheduler,
		stats:      stats,
		dataSource: dataSource,
	}
}

func successResponse(data interface{}) common.APIResponse {
	return common.APIResponse{
		Code:    common.ErrCodeSuccess,
		Message: common.ErrorMessage(common.ErrCodeSuccess),
		Data:    data,
	}
}

func errorResponse(err error) common.APIResponse {
	if appErr, ok := err.(*common.AppError); ok {
		return common.APIResponse{
			Code:    appErr.Code,
			Message: appErr.Message,
			Data:    nil,
		}
	}
	return common.APIResponse{
		Code:    common.ErrCodeInternalError,
		Message: common.ErrorMessage(common.ErrCodeInternalError),
		Data:    nil,
	}
}

func writeJSON(w http.ResponseWriter, resp common.APIResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api")

	switch {
	case path == "/templates" && r.Method == "GET":
		h.listTemplates(w, r)
	case path == "/templates" && r.Method == "POST":
		h.createTemplate(w, r)
	case strings.HasPrefix(path, "/templates/") && r.Method == "GET":
		h.getTemplate(w, r)
	case strings.HasPrefix(path, "/templates/") && r.Method == "PUT":
		h.updateTemplate(w, r)
	case strings.HasPrefix(path, "/templates/") && r.Method == "DELETE":
		h.deleteTemplate(w, r)

	case path == "/tasks" && r.Method == "POST":
		h.createTask(w, r)
	case strings.HasPrefix(path, "/tasks/") && strings.HasSuffix(path, "/progress") && r.Method == "GET":
		h.getTaskProgress(w, r)
	case strings.HasPrefix(path, "/tasks/") && strings.HasSuffix(path, "/download") && r.Method == "GET":
		h.downloadTaskFile(w, r)
	case strings.HasPrefix(path, "/tasks/") && r.Method == "GET":
		h.getTask(w, r)

	case path == "/stats" && r.Method == "GET":
		h.getStats(w, r)

	default:
		writeJSON(w, errorResponse(common.NewNotFoundError("endpoint not found")))
	}
}

func (h *Handler) listTemplates(w http.ResponseWriter, r *http.Request) {
	templates := h.templates.List()
	writeJSON(w, successResponse(templates))
}

func (h *Handler) createTemplate(w http.ResponseWriter, r *http.Request) {
	var req common.CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, errorResponse(common.NewInvalidRequestError("invalid request body")))
		return
	}

	tpl, err := h.templates.Create(&req)
	if err != nil {
		writeJSON(w, errorResponse(err))
		return
	}

	writeJSON(w, successResponse(tpl))
}

func extractID(path string, prefix string) string {
	id := strings.TrimPrefix(path, prefix)
	if idx := strings.Index(id, "/"); idx > 0 {
		id = id[:idx]
	}
	return id
}

func (h *Handler) getTemplate(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/templates/")
	if id == "" {
		writeJSON(w, errorResponse(common.NewInvalidRequestError("template id is required")))
		return
	}

	tpl, err := h.templates.GetByID(id)
	if err != nil {
		writeJSON(w, errorResponse(err))
		return
	}

	writeJSON(w, successResponse(tpl))
}

func (h *Handler) updateTemplate(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/templates/")
	if id == "" {
		writeJSON(w, errorResponse(common.NewInvalidRequestError("template id is required")))
		return
	}

	var req common.UpdateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, errorResponse(common.NewInvalidRequestError("invalid request body")))
		return
	}

	tpl, err := h.templates.Update(id, &req)
	if err != nil {
		writeJSON(w, errorResponse(err))
		return
	}

	writeJSON(w, successResponse(tpl))
}

func (h *Handler) deleteTemplate(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/templates/")
	if id == "" {
		writeJSON(w, errorResponse(common.NewInvalidRequestError("template id is required")))
		return
	}

	if err := h.templates.Delete(id); err != nil {
		writeJSON(w, errorResponse(err))
		return
	}

	writeJSON(w, successResponse(nil))
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	var req common.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, errorResponse(common.NewInvalidRequestError("invalid request body")))
		return
	}

	if req.TemplateID == "" {
		writeJSON(w, errorResponse(common.NewInvalidRequestError("template_id is required")))
		return
	}
	if req.UserID == "" {
		writeJSON(w, errorResponse(common.NewInvalidRequestError("user_id is required")))
		return
	}

	task, err := h.scheduler.SubmitTask(&req)
	if err != nil {
		writeJSON(w, errorResponse(err))
		return
	}

	writeJSON(w, successResponse(map[string]interface{}{
		"task_id": task.ID,
		"status":  task.Status,
	}))
}

func (h *Handler) getTask(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/tasks/")
	if id == "" {
		writeJSON(w, errorResponse(common.NewInvalidRequestError("task id is required")))
		return
	}

	task, err := h.scheduler.GetTask(id)
	if err != nil {
		writeJSON(w, errorResponse(err))
		return
	}

	writeJSON(w, successResponse(task))
}

func (h *Handler) getTaskProgress(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/progress")
	id := extractID(path, "/api/tasks/")
	if id == "" {
		writeJSON(w, errorResponse(common.NewInvalidRequestError("task id is required")))
		return
	}

	progress, err := h.scheduler.GetTaskProgress(id)
	if err != nil {
		writeJSON(w, errorResponse(err))
		return
	}

	writeJSON(w, successResponse(progress))
}

func (h *Handler) downloadTaskFile(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/download")
	id := extractID(path, "/api/tasks/")
	if id == "" {
		writeJSON(w, errorResponse(common.NewInvalidRequestError("task id is required")))
		return
	}

	fileData, fileName, err := h.scheduler.GetTaskFile(id)
	if err != nil {
		writeJSON(w, errorResponse(err))
		return
	}

	encodedName := url.PathEscape(fileName)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", encodedName))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.Itoa(len(fileData)))
	w.Write(fileData)
}

func (h *Handler) getStats(w http.ResponseWriter, r *http.Request) {
	days := 7
	topN := 10

	if d := r.URL.Query().Get("days"); d != "" {
		if n, err := strconv.Atoi(d); err == nil && n > 0 {
			days = n
		}
	}
	if t := r.URL.Query().Get("top"); t != "" {
		if n, err := strconv.Atoi(t); err == nil && n > 0 {
			topN = n
		}
	}

	stats := h.stats.GetExportStats(days, topN)
	writeJSON(w, successResponse(stats))
}
