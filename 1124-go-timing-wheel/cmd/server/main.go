package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/timingwheel/pkg/api"
	"github.com/timingwheel/pkg/timingwheel"
)

type Server struct {
	tw *timingwheel.TimingWheel
}

func NewServer() *Server {
	return &Server{
		tw: timingwheel.New(),
	}
}

func (s *Server) handleAdd(w http.ResponseWriter, r *http.Request) {
	var req api.AddTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, &api.ErrorResponse{Success: false, Error: err.Error()})
		return
	}

	err := s.tw.Add(req.ID, req.Callback, req.Delay)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &api.AddTaskResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, &api.AddTaskResponse{Success: true})
}

func (s *Server) handleCancel(w http.ResponseWriter, r *http.Request) {
	var req api.CancelTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, &api.ErrorResponse{Success: false, Error: err.Error()})
		return
	}

	success, err := s.tw.Cancel(req.ID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, &api.CancelTaskResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, &api.CancelTaskResponse{Success: success})
}

func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	var req api.ResetTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, &api.ErrorResponse{Success: false, Error: err.Error()})
		return
	}

	err := s.tw.Reset(req.ID, req.NewDelay)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &api.ResetTaskResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, &api.ResetTaskResponse{Success: true})
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, &api.ErrorResponse{Success: false, Error: "missing id parameter"})
		return
	}

	task, err := s.tw.Get(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, &api.GetTaskResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, &api.GetTaskResponse{
		Success: true,
		Task:    taskToInfo(task),
	})
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	tasks := s.tw.List()
	taskInfos := make([]*api.TaskInfo, 0, len(tasks))
	for _, t := range tasks {
		taskInfos = append(taskInfos, taskToInfo(t))
	}

	writeJSON(w, http.StatusOK, &api.ListTasksResponse{
		Success: true,
		Tasks:   taskInfos,
	})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := s.tw.Stats()
	writeJSON(w, http.StatusOK, &api.StatsResponse{
		Success:        true,
		TotalPending:   stats.TotalPending,
		ExecutedCount:  stats.ExecutedCount,
		CancelledCount: stats.CancelledCount,
		HourTasks:      stats.HourTasks,
		MinuteTasks:    stats.MinuteTasks,
		SecondTasks:    stats.SecondTasks,
		HourTick:       stats.HourTick,
		MinuteTick:     stats.MinuteTick,
		SecondTick:     stats.SecondTick,
	})
}

func taskToInfo(t *timingwheel.Task) *api.TaskInfo {
	statusStr := "pending"
	switch t.Status() {
	case timingwheel.TaskStatusRunning:
		statusStr = "running"
	case timingwheel.TaskStatusCancelled:
		statusStr = "cancelled"
	case timingwheel.TaskStatusExecuted:
		statusStr = "executed"
	}

	return &api.TaskInfo{
		ID:        t.ID,
		Callback:  t.Callback,
		Delay:     t.Delay,
		CreatedAt: t.CreatedAt,
		ExpireAt:  t.ExpireAt,
		Remaining: t.Remaining(),
		Status:    statusStr,
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func main() {
	s := NewServer()
	s.tw.Start()

	mux := http.NewServeMux()
	mux.HandleFunc("/tasks/add", s.handleAdd)
	mux.HandleFunc("/tasks/cancel", s.handleCancel)
	mux.HandleFunc("/tasks/reset", s.handleReset)
	mux.HandleFunc("/tasks/get", s.handleGet)
	mux.HandleFunc("/tasks/list", s.handleList)
	mux.HandleFunc("/stats", s.handleStats)

	addr := ":8080"
	fmt.Printf("TimingWheel server starting on %s...\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
