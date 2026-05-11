package main

import (
	"encoding/json"
	"net/http"
	"time"

	"delayqueue/pkg/api"
	"delayqueue/pkg/queue"
)

type Server struct {
	dq *queue.DelayQueue
}

func NewServer() *Server {
	return &Server{
		dq: queue.New(),
	}
}

func (s *Server) handleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.SubmitResponse{Success: false, Error: err.Error()})
		return
	}

	if req.Priority == 0 {
		req.Priority = api.PriorityNormal
	}

	task := &queue.Task{
		ID:        req.ID,
		Payload:   []byte(req.Payload),
		ExecuteAt: req.ExecuteAt,
		Priority:  queue.Priority(req.Priority),
	}

	if err := s.dq.Push(task); err != nil {
		writeJSON(w, http.StatusConflict, api.SubmitResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, api.SubmitResponse{Success: true})
}

func (s *Server) handleModify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.ModifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ModifyResponse{Success: false, Error: err.Error()})
		return
	}

	if err := s.dq.ModifyDelay(req.ID, req.ExecuteAt); err != nil {
		writeJSON(w, http.StatusNotFound, api.ModifyResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, api.ModifyResponse{Success: true})
}

func (s *Server) handlePromote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.PromoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.PromoteResponse{Success: false, Error: err.Error()})
		return
	}

	if err := s.dq.Promote(req.ID, queue.Priority(req.Priority)); err != nil {
		writeJSON(w, http.StatusNotFound, api.PromoteResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, api.PromoteResponse{Success: true})
}

func (s *Server) handlePoll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	task := s.dq.Poll(time.Now())
	resp := api.PollResponse{}
	if task != nil {
		resp.Task = &api.Task{
			ID:        task.ID,
			Payload:   string(task.Payload),
			ExecuteAt: task.ExecuteAt,
			Priority:  api.Priority(task.Priority),
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handlePeek(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	task := s.dq.Peek()
	resp := api.PeekResponse{}
	if task != nil {
		resp.Task = &api.Task{
			ID:        task.ID,
			Payload:   string(task.Payload),
			ExecuteAt: task.ExecuteAt,
			Priority:  api.Priority(task.Priority),
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleDrain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tasks := s.dq.Drain(time.Now())
	apiTasks := make([]*api.Task, 0, len(tasks))
	for _, task := range tasks {
		apiTasks = append(apiTasks, &api.Task{
			ID:        task.ID,
			Payload:   string(task.Payload),
			ExecuteAt: task.ExecuteAt,
			Priority:  api.Priority(task.Priority),
		})
	}

	writeJSON(w, http.StatusOK, api.DrainResponse{Tasks: apiTasks})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, api.StatusResponse{Queued: s.dq.Len()})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func main() {
	s := NewServer()

	http.HandleFunc("/submit", s.handleSubmit)
	http.HandleFunc("/modify", s.handleModify)
	http.HandleFunc("/promote", s.handlePromote)
	http.HandleFunc("/poll", s.handlePoll)
	http.HandleFunc("/peek", s.handlePeek)
	http.HandleFunc("/drain", s.handleDrain)
	http.HandleFunc("/status", s.handleStatus)

	_ = http.ListenAndServe(":8080", nil)
}
