package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"dag-topo-sort/pkg/api"
	"dag-topo-sort/pkg/toposort"
)

type Handler struct {
	graph *toposort.Graph
	mu    sync.RWMutex
}

func NewHandler() *Handler {
	return &Handler{
		graph: toposort.NewGraph(),
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.mu.RLock()
	tasks := h.graph.GetAllTasks()
	h.mu.RUnlock()

	apiTasks := make([]api.Task, len(tasks))
	for i, task := range tasks {
		apiTasks[i] = api.Task{
			ID:           task.ID,
			Name:         task.Name,
			Dependencies: task.Dependencies,
		}
	}

	writeJSON(w, http.StatusOK, api.ListTasksResponse{
		Success: true,
		Tasks:   apiTasks,
	})
}

func (h *Handler) AddTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.AddTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.AddTaskResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	h.mu.Lock()
	err := h.graph.AddTask(toposort.Task{
		ID:           req.Task.ID,
		Name:         req.Task.Name,
		Dependencies: req.Task.Dependencies,
	})
	h.mu.Unlock()

	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.AddTaskResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, api.AddTaskResponse{
		Success: true,
		Message: "Task added successfully",
	})
}

func (h *Handler) BatchAddTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.BatchAddTasksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.BatchAddTasksResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	h.mu.Lock()
	added := 0
	for _, t := range req.Tasks {
		err := h.graph.AddTask(toposort.Task{
			ID:           t.ID,
			Name:         t.Name,
			Dependencies: t.Dependencies,
		})
		if err == nil {
			added++
		}
	}
	h.mu.Unlock()

	writeJSON(w, http.StatusOK, api.BatchAddTasksResponse{
		Success: true,
		Added:   added,
		Message: "Batch operation completed",
	})
}

func (h *Handler) RemoveTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.RemoveTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.RemoveTaskResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	h.mu.Lock()
	h.graph.RemoveTask(req.TaskID)
	h.mu.Unlock()

	writeJSON(w, http.StatusOK, api.RemoveTaskResponse{
		Success: true,
		Message: "Task removed successfully",
	})
}

func (h *Handler) UpdateDependencies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.UpdateDependenciesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.UpdateDependenciesResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	h.mu.Lock()
	err := h.graph.UpdateDependencies(req.TaskID, req.Dependencies)
	h.mu.Unlock()

	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, toposort.ErrTaskNotFound{}) {
			status = http.StatusNotFound
		}
		writeJSON(w, status, api.UpdateDependenciesResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, api.UpdateDependenciesResponse{
		Success: true,
		Message: "Dependencies updated successfully",
	})
}

func (h *Handler) ClearTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.mu.Lock()
	h.graph = toposort.NewGraph()
	h.mu.Unlock()

	writeJSON(w, http.StatusOK, api.ClearTasksResponse{
		Success: true,
		Message: "All tasks cleared",
	})
}

func (h *Handler) Sort(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.mu.RLock()
	result, err := h.graph.Sort()
	h.mu.RUnlock()

	if err != nil {
		resp := api.SortResponse{
			Success: false,
			Error:   err.Error(),
		}

		var cyclicErr toposort.ErrCyclicDependency
		if errors.As(err, &cyclicErr) {
			resp.ErrorType = "cyclic"
			resp.CyclePath = cyclicErr.Path
		} else if errors.Is(err, toposort.ErrSelfDependency{}) {
			resp.ErrorType = "self_dependency"
		} else if errors.Is(err, toposort.ErrDependencyNotFound{}) {
			resp.ErrorType = "dependency_not_found"
		} else {
			resp.ErrorType = "unknown"
		}

		writeJSON(w, http.StatusBadRequest, resp)
		return
	}

	writeJSON(w, http.StatusOK, api.SortResponse{
		Success:           true,
		Order:             result.Order,
		MaxParallelism:    result.MaxParallelism,
		CriticalPath:      result.CriticalPath,
		MinCompletionTime: result.MinCompletionTime,
	})
}

func (h *Handler) GenerateDot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.mu.RLock()
	dot := h.graph.GenerateDot()
	h.mu.RUnlock()

	writeJSON(w, http.StatusOK, api.DotResponse{
		Success: true,
		Dot:     dot,
	})
}

func (h *Handler) GenerateDotWithCriticalPath(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.mu.RLock()
	result, err := h.graph.Sort()
	if err != nil {
		h.mu.RUnlock()
		writeJSON(w, http.StatusBadRequest, api.DotResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	dot := h.graph.GenerateDotWithCriticalPath(result.CriticalPath)
	h.mu.RUnlock()

	writeJSON(w, http.StatusOK, api.DotResponse{
		Success: true,
		Dot:     dot,
	})
}
