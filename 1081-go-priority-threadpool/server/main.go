package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/solocoder/priority-threadpool/common"
	"github.com/solocoder/priority-threadpool/pool"
)

type server struct {
	pool *pool.Pool
	mux  *http.ServeMux
}

func newServer(workerCount int, starvationThreshold time.Duration) (*server, error) {
	p, err := pool.New(workerCount, starvationThreshold)
	if err != nil {
		return nil, err
	}

	s := &server{
		pool: p,
		mux:  http.NewServeMux(),
	}

	s.mux.HandleFunc("/submit", s.handleSubmit)
	s.mux.HandleFunc("/status", s.handleStatus)
	s.mux.HandleFunc("/shutdown", s.handleShutdown)

	return s, nil
}

func (s *server) handleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	priority := convertPriority(req.Priority)
	err := s.pool.Submit(priority, func() {
		fmt.Printf("Executing task: %s\n", req.Task)
		time.Sleep(100 * time.Millisecond)
	})

	resp := common.SubmitResponse{}
	if err != nil {
		resp.Success = false
		resp.Error = err.Error()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp.Success = true
	resp.Message = "Task submitted successfully"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := s.pool.Status()
	resp := common.StatusResponse{
		WorkerCount:         status.WorkerCount,
		IdleWorkerCount:   status.IdleWorkerCount,
		QueueHighPriority:  status.QueueHighPriority,
		QueueMediumPriority: status.QueueMediumPriority,
		QueueLowPriority:     status.QueueLowPriority,
		StarvedTasksCount:     status.StarvedTasksCount,
		Success:             true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *server) handleShutdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	go func() {
		s.pool.Shutdown()
	}()

	resp := common.ShutdownResponse{
		Success: true,
		Message: "Shutdown initiated, waiting for all tasks to complete",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func convertPriority(p common.Priority) pool.Priority {
	switch p {
	case common.High:
		return pool.High
	case common.Medium:
		return pool.Medium
	case common.Low:
		return pool.Low
	default:
		return pool.Medium
	}
}

func main() {
	workerCount := flag.Int("workers", 4, "Number of worker goroutines")
	starvationThreshold := flag.Duration("threshold", 30*time.Second, "Starvation threshold for low priority tasks")
	port := flag.String("port", "8080", "Port to listen on")
	flag.Parse()

	srv, err := newServer(*workerCount, *starvationThreshold)
	if err != nil {
		fmt.Printf("Failed to create server: %v\n", err)
		return
	}

	fmt.Printf("Server starting on port %s with %d workers and starvation threshold of %s\n", *port, *workerCount, *starvationThreshold)
	http.ListenAndServe(":"+*port, srv.mux)
}
