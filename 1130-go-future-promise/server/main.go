package main

import (
	"encoding/json"
	"future-promise/api"
	"log"
	"net/http"
	"strings"
)

type Server struct {
	tm *TaskManager
}

func NewServer() *Server {
	return &Server{
		tm: NewTaskManager(),
	}
}

func (s *Server) handleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.SubmitTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	taskID, err := s.tm.SubmitTask(&req)
	if err != nil {
		http.Error(w, "Failed to submit task: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := api.SubmitTaskResponse{
		Success: true,
		TaskID:  taskID,
		Message: "Task submitted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Task ID not provided", http.StatusBadRequest)
		return
	}

	taskID := parts[2]
	task, ok := s.tm.GetTask(taskID)
	if !ok {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	resp := task.ToResponse()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Task ID not provided", http.StatusBadRequest)
		return
	}

	taskID := parts[2]
	success := s.tm.CancelTask(taskID)

	var resp api.CancelTaskResponse
	if success {
		resp = api.CancelTaskResponse{
			Success: true,
			Message: "Task cancelled successfully",
		}
	} else {
		resp = api.CancelTaskResponse{
			Success: false,
			Message: "Task not found or already completed",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logs := s.tm.GetLogs()
	resp := api.GetLogsResponse{
		Success: true,
		Logs:    logs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"service": "future-promise-server",
		"version": "1.0.0",
		"endpoints": "POST /submit, GET /task/{id}, POST /cancel/{id}, GET /logs",
	})
}

func main() {
	server := NewServer()

	http.HandleFunc("/", server.handleRoot)
	http.HandleFunc("/submit", server.handleSubmit)
	http.HandleFunc("/task/", server.handleTask)
	http.HandleFunc("/cancel/", server.handleCancel)
	http.HandleFunc("/logs", server.handleLogs)

	log.Println("Future/Promise Server starting on :8080")
	log.Println("Endpoints:")
	log.Println("  POST   /submit       - Submit a new task")
	log.Println("  GET    /task/{id}    - Get task status and result")
	log.Println("  POST   /cancel/{id}  - Cancel a task")
	log.Println("  GET    /logs         - Get all task logs")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
