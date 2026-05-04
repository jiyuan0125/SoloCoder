package main

import (
	"batch-processor/batch"
	"batch-processor/protocol"
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type ServerHandler struct {
	processor *batch.BatchProcessor[string]
	stopped   bool
	mu        sync.Mutex
}

func NewServerHandler(processor *batch.BatchProcessor[string]) *ServerHandler {
	return &ServerHandler{
		processor: processor,
		stopped:   false,
	}
}

func (h *ServerHandler) SubmitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Data) == 0 {
		h.writeJSON(w, protocol.SubmitResponse{
			Success: true,
			Message: "no data provided",
		})
		return
	}

	h.mu.Lock()
	if h.stopped {
		h.mu.Unlock()
		h.writeError(w, "server is shutting down", http.StatusServiceUnavailable)
		return
	}
	h.mu.Unlock()

	h.processor.AddAll(req.Data)

	h.writeJSON(w, protocol.SubmitResponse{
		Success: true,
		Message: "data submitted",
	})
}

func (h *ServerHandler) FlushHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.mu.Lock()
	if h.stopped {
		h.mu.Unlock()
		h.writeError(w, "server is shutting down", http.StatusServiceUnavailable)
		return
	}
	h.mu.Unlock()

	h.processor.Flush()

	h.writeJSON(w, protocol.FlushResponse{
		Success: true,
		Message: "flush requested",
	})
}

func (h *ServerHandler) StatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	size := h.processor.Size()

	h.writeJSON(w, protocol.StatsResponse{
		Success:    true,
		BufferSize: size,
	})
}

func (h *ServerHandler) ShutdownHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.mu.Lock()
	if h.stopped {
		h.mu.Unlock()
		h.writeJSON(w, protocol.ShutdownResponse{
			Success: true,
			Message: "already shutting down",
		})
		return
	}
	h.stopped = true
	h.mu.Unlock()

	go func() {
		log.Println("Closing batch processor...")
		h.processor.Close()
		log.Println("Batch processor closed")
	}()

	h.writeJSON(w, protocol.ShutdownResponse{
		Success: true,
		Message: "shutdown initiated",
	})
}

func (h *ServerHandler) writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("Failed to write JSON response: %v", err)
	}
}

func (h *ServerHandler) writeError(w http.ResponseWriter, err string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(protocol.ErrorResponse{
		Success: false,
		Error:   err,
	})
}
