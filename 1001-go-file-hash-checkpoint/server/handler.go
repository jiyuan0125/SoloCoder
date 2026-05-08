package main

import (
	"encoding/json"
	"net/http"
	"time"

	"hash-checkpoint/common"
	"hash-checkpoint/core"
)

type Handler struct {
	manager *core.CheckpointManager
}

func NewHandler(manager *core.CheckpointManager) *Handler {
	return &Handler{manager: manager}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, common.ErrorResponse{
		Success: false,
		Error:   message,
	})
}

func (h *Handler) StartCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.StartCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ChunkSize <= 0 {
		req.ChunkSize = common.DefaultChunkSize
	}

	progress, nextIndex := h.manager.StartOrResume(
		req.FilePath,
		req.ModTime,
		req.FileSize,
		req.Algorithm,
		req.ChunkSize,
	)

	var message string
	if nextIndex == 0 {
		message = "started new check"
	} else {
		message = "resumed from existing progress"
	}

	writeJSON(w, http.StatusOK, common.StartCheckResponse{
		Success:        true,
		Message:        message,
		NextChunkIndex: nextIndex,
		Progress:       progress,
	})
}

func (h *Handler) SubmitChunk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.SubmitChunkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	progress, nextIndex, err := h.manager.SubmitChunk(
		req.FilePath,
		req.ModTime,
		req.ChunkIndex,
		req.ChunkHash,
		req.ChunkLen,
	)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if progress == nil {
		writeError(w, http.StatusNotFound, "checkpoint not found, file may have been modified")
		return
	}

	var message string
	if progress.Status == common.StatusChunkMismatch {
		message = "chunk hash mismatch, restarting from this chunk"
	} else {
		message = "chunk submitted"
	}

	writeJSON(w, http.StatusOK, common.SubmitChunkResponse{
		Success:        true,
		Message:        message,
		NextChunkIndex: nextIndex,
		Progress:       progress,
	})
}

func (h *Handler) GetProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	filePath := r.URL.Query().Get("file_path")
	modTimeStr := r.URL.Query().Get("mod_time")

	if filePath == "" || modTimeStr == "" {
		writeError(w, http.StatusBadRequest, "missing file_path or mod_time")
		return
	}

	modTime, err := time.Parse(time.RFC3339Nano, modTimeStr)
	if err != nil {
		modTime, err = time.Parse(time.RFC3339, modTimeStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid mod_time format")
			return
		}
	}

	progress := h.manager.GetProgress(filePath, modTime)
	if progress == nil {
		writeError(w, http.StatusNotFound, "progress not found")
		return
	}

	writeJSON(w, http.StatusOK, common.GetProgressResponse{
		Success:  true,
		Message:  "ok",
		Progress: progress,
	})
}

func (h *Handler) CompleteCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.CompleteCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	progress, finalHash, err := h.manager.CompleteCheck(req.FilePath, req.ModTime)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if progress == nil {
		writeError(w, http.StatusNotFound, "checkpoint not found")
		return
	}

	if progress.CompletedChunks != progress.TotalChunks {
		writeError(w, http.StatusBadRequest, "check not complete, missing chunks")
		return
	}

	writeJSON(w, http.StatusOK, common.CompleteCheckResponse{
		Success:   true,
		Message:   "check completed",
		FinalHash: finalHash,
	})
}
