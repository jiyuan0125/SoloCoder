package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"progress-tracker/common"
)

type Handler struct {
	store *ProgressStore
}

func NewHandler(store *ProgressStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) CreateProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.CreateProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	id := h.store.Create(req.Total, req.Description)

	resp := common.CreateProgressResponse{
		ID:     id,
		Status: "created",
	}

	h.sendJSON(w, resp, http.StatusCreated)
}

func (h *Handler) GetProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := h.extractID(r.URL.Path)
	if id == "" {
		h.listProgresses(w, r)
		return
	}

	status, exists := h.store.Get(id)
	if !exists {
		h.sendError(w, "Progress not found", http.StatusNotFound)
		return
	}

	h.sendJSON(w, status, http.StatusOK)
}

func (h *Handler) listProgresses(w http.ResponseWriter, r *http.Request) {
	statuses := h.store.List()
	resp := common.ListProgressResponse{
		Progresses: statuses,
	}
	h.sendJSON(w, resp, http.StatusOK)
}

func (h *Handler) UpdateProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := h.extractID(r.URL.Path)
	if id == "" {
		h.sendError(w, "Progress ID required", http.StatusBadRequest)
		return
	}

	var req common.UpdateProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !h.store.Update(id, req.Delta) {
		h.sendError(w, "Progress not found", http.StatusNotFound)
		return
	}

	status, _ := h.store.Get(id)
	h.sendJSON(w, status, http.StatusOK)
}

func (h *Handler) SetProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := h.extractID(r.URL.Path)
	if id == "" {
		h.sendError(w, "Progress ID required", http.StatusBadRequest)
		return
	}

	var req common.SetProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !h.store.Set(id, req.Current) {
		h.sendError(w, "Progress not found", http.StatusNotFound)
		return
	}

	status, _ := h.store.Get(id)
	h.sendJSON(w, status, http.StatusOK)
}

func (h *Handler) CancelProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := h.extractID(r.URL.Path)
	if id == "" {
		h.sendError(w, "Progress ID required", http.StatusBadRequest)
		return
	}

	if !h.store.Cancel(id) {
		h.sendError(w, "Progress not found", http.StatusNotFound)
		return
	}

	status, _ := h.store.Get(id)
	h.sendJSON(w, status, http.StatusOK)
}

func (h *Handler) CloseProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := h.extractID(r.URL.Path)
	if id == "" {
		h.sendError(w, "Progress ID required", http.StatusBadRequest)
		return
	}

	if !h.store.Close(id) {
		h.sendError(w, "Progress not found", http.StatusNotFound)
		return
	}

	h.sendJSON(w, map[string]string{"status": "closed"}, http.StatusOK)
}

func (h *Handler) AddSubTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.AddSubTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ParentID == "" {
		h.sendError(w, "Parent ID required", http.StatusBadRequest)
		return
	}

	id, success := h.store.AddSubTask(req.ParentID, req.Total, req.Weight)
	if !success {
		h.sendError(w, "Parent progress not found", http.StatusNotFound)
		return
	}

	resp := common.AddSubTaskResponse{
		ID:     id,
		Status: "created",
	}

	h.sendJSON(w, resp, http.StatusCreated)
}

func (h *Handler) extractID(path string) string {
	parts := strings.Split(path, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" && strings.HasPrefix(parts[i], "p_") {
			return parts[i]
		}
	}
	return ""
}

func (h *Handler) sendJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) sendError(w http.ResponseWriter, message string, status int) {
	resp := common.ErrorResponse{Error: message}
	h.sendJSON(w, resp, status)
}
