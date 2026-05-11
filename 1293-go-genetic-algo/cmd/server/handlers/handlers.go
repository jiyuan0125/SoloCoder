package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"genetic-algo/cmd/server/task"
	"genetic-algo/pkg/api"
	"genetic-algo/pkg/testfuncs"
)

type Handler struct {
	tasks *task.Manager
}

func Register(mux *http.ServeMux, tm *task.Manager) {
	h := &Handler{tasks: tm}

	mux.HandleFunc("/api/v1/functions", h.listFunctions)
	mux.HandleFunc("/api/v1/optimize", h.optimize)
	mux.HandleFunc("/api/v1/tasks", h.listTasks)
	mux.HandleFunc("/api/v1/tasks/", h.taskEndpoint)
}

func (h *Handler) listFunctions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	names := testfuncs.ListFunctions()
	functions := make([]api.FunctionInfo, 0, len(names))

	for _, name := range names {
		fn, err := testfuncs.GetFunction(name)
		if err != nil {
			continue
		}
		info := api.FunctionInfo{
			Name:        name,
			Description: fn.Description,
			Minimize:    fn.Minimize,
			DefaultLower: fn.DefaultLower[:min(3, len(fn.DefaultLower))],
			DefaultUpper: fn.DefaultUpper[:min(3, len(fn.DefaultUpper))],
		}
		functions = append(functions, info)
	}

	writeJSON(w, http.StatusOK, api.ListFunctionsResponse{Functions: functions})
}

func (h *Handler) optimize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.OptimizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if req.Problem == nil {
		writeError(w, http.StatusBadRequest, "problem is required")
		return
	}

	t, err := h.tasks.CreateTask(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := api.TaskResponse{
		TaskID:    t.ID,
		Status:    t.Status,
		CreatedAt: t.CreatedAt.Format(timeFormat),
	}

	writeJSON(w, http.StatusAccepted, resp)
}

func (h *Handler) listTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	tasks := h.tasks.ListTasks()
	infos := make([]api.TaskInfo, 0, len(tasks))

	for _, t := range tasks {
		info := api.TaskInfo{
			ID:         t.ID,
			Status:     t.Status,
			Function:   getFunctionName(t.Request),
			CurrentGen: t.CurrentGen,
			BestFitness: t.BestFitness,
			CreatedAt:  t.CreatedAt.Format(timeFormat),
		}
		infos = append(infos, info)
	}

	writeJSON(w, http.StatusOK, api.TaskListResponse{Tasks: infos})
}

func (h *Handler) taskEndpoint(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/tasks/")
	parts := strings.SplitN(path, "/", 2)

	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "task id required")
		return
	}

	taskID := parts[0]
	t, err := h.tasks.GetTask(taskID)
	if err != nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}

	if len(parts) > 1 && parts[1] == "statistics" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.getStatistics(w, t)
		return
	}

	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	h.getTask(w, t)
}

func (h *Handler) getTask(w http.ResponseWriter, t *task.Task) {
	result := api.TaskResult{
		TaskID:      t.ID,
		Status:      t.Status,
		CurrentGen:  t.CurrentGen,
		TotalGens:   t.TotalGens,
		BestFitness: t.BestFitness,
		BestGenes:   t.BestGenes,
		Error:       t.Error,
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) getStatistics(w http.ResponseWriter, t *task.Task) {
	if t.Algorithm == nil {
		writeJSON(w, http.StatusOK, []api.Statistics{})
		return
	}

	stats := make([]api.Statistics, len(t.Algorithm.Statistics))
	for i, s := range t.Algorithm.Statistics {
		stats[i] = api.Statistics{
			Generation:   s.Generation,
			BestFitness:  s.BestFitness,
			MeanFitness:  s.MeanFitness,
			StdDeviation: s.StdDeviation,
			BestGenes:    s.BestGenes,
		}
	}

	writeJSON(w, http.StatusOK, stats)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, api.ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}

func getFunctionName(req *api.OptimizeRequest) string {
	if req == nil || req.Problem == nil {
		return ""
	}
	if req.Problem.CustomExpr != "" {
		return "custom"
	}
	return req.Problem.Function
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

const timeFormat = "2006-01-02T15:04:05.000Z07:00"
