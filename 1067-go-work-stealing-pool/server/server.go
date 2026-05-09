package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/workstealing/pool/common"
	"github.com/workstealing/pool/pool"
)

type TaskServer struct {
	pool      *pool.Pool
	futures   map[string]*pool.Future
	futuresMu sync.RWMutex
	nextID    int
	idMu      sync.Mutex
}

func NewTaskServer(workerCount int) *TaskServer {
	return &TaskServer{
		pool:    pool.New(workerCount),
		futures: make(map[string]*pool.Future),
	}
}

func (s *TaskServer) generateTaskID() string {
	s.idMu.Lock()
	id := s.nextID
	s.nextID++
	s.idMu.Unlock()
	return fmt.Sprintf("task-%d", id)
}

func (s *TaskServer) storeFuture(taskID string, f *pool.Future) {
	s.futuresMu.Lock()
	s.futures[taskID] = f
	s.futuresMu.Unlock()
}

func (s *TaskServer) getFuture(taskID string) (*pool.Future, bool) {
	s.futuresMu.RLock()
	f, ok := s.futures[taskID]
	s.futuresMu.RUnlock()
	return f, ok
}

func (s *TaskServer) SubmitHandler(w http.ResponseWriter, r *http.Request) {
	var req common.SubmitTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	taskID := s.generateTaskID()
	resp := common.SubmitTaskResponse{
		TaskID: taskID,
	}

	taskFn := func() {
		switch req.TaskType {
		case "fib":
			s.fibTask(req.Payload)
		case "echo":
			s.echoTask(req.Payload)
		case "sleep":
			s.sleepTask(req.Payload)
		case "panic":
			s.panicTask(req.Payload)
		default:
			log.Printf("Unknown task type: %s", req.TaskType)
		}
	}

	var future *pool.Future
	var err error
	if req.Timeout > 0 {
		future, err = s.pool.SubmitWithTimeout(taskFn, req.Timeout)
	} else {
		future, err = s.pool.Submit(taskFn)
	}

	if err != nil {
		resp.Success = false
		resp.Error = err.Error()
		json.NewEncoder(w).Encode(resp)
		return
	}

	s.storeFuture(taskID, future)
	resp.Success = true
	json.NewEncoder(w).Encode(resp)
}

func (s *TaskServer) fibTask(payload interface{}) {
	n, ok := payload.(float64)
	if !ok {
		log.Printf("Invalid fib payload: %v", payload)
		return
	}
	fib(int(n))
}

func (s *TaskServer) echoTask(payload interface{}) {
	log.Printf("Echo: %v", payload)
}

func (s *TaskServer) sleepTask(payload interface{}) {
	d, ok := payload.(float64)
	if !ok {
		log.Printf("Invalid sleep payload: %v", payload)
		return
	}
	time.Sleep(time.Duration(d) * time.Millisecond)
}

func (s *TaskServer) panicTask(payload interface{}) {
	panic(fmt.Sprintf("intentional panic: %v", payload))
}

func fib(n int) int {
	if n <= 1 {
		return n
	}
	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

func (s *TaskServer) GetResultHandler(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	if taskID == "" {
		http.Error(w, "missing task_id", http.StatusBadRequest)
		return
	}

	f, ok := s.getFuture(taskID)
	if !ok {
		json.NewEncoder(w).Encode(common.GetTaskResultResponse{
			TaskID: taskID,
			Status: "not_found",
			Done:   true,
		})
		return
	}

	resp := common.GetTaskResultResponse{
		TaskID: taskID,
	}

	select {
	case <-f.Done():
		result, err := f.Get()
		resp.Done = true
		if err != nil {
			resp.Status = "error"
			resp.Error = err.Error()
		} else {
			resp.Status = "completed"
			resp.Result = result
		}
	default:
		resp.Status = "running"
		resp.Done = false
	}

	json.NewEncoder(w).Encode(resp)
}

func (s *TaskServer) StatsHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(common.PoolStatsResponse{
		WorkerCount: s.pool.Size(),
		Status:      "running",
	})
}

func (s *TaskServer) AdjustWorkersHandler(w http.ResponseWriter, r *http.Request) {
	var req common.AdjustWorkersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Add > 0 {
		s.pool.AddWorkers(req.Add)
	}
	if req.Remove > 0 {
		s.pool.RemoveWorkers(req.Remove)
	}

	json.NewEncoder(w).Encode(common.PoolStatsResponse{
		WorkerCount: s.pool.Size(),
		Status:      "running",
	})
}

func (s *TaskServer) ShutdownHandler(w http.ResponseWriter, r *http.Request) {
	var req common.ShutdownRequest
	json.NewDecoder(r.Body).Decode(&req)

	go func() {
		if req.Force {
			s.pool.ShutdownNow()
		} else {
			s.pool.Shutdown()
		}
	}()

	json.NewEncoder(w).Encode(common.ShutdownResponse{
		Success: true,
		Message: "Shutdown initiated",
	})
}
