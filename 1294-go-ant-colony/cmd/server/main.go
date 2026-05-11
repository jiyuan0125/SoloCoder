package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/example/antcolony/pkg/antcolony"
	"github.com/example/antcolony/pkg/common"
)

type task struct {
	solver *antcolony.Solver
	mu     sync.Mutex
}

var (
	tasks   = make(map[string]*task)
	tasksMu sync.RWMutex
)

func generateTaskID() string {
	return fmt.Sprintf("%d", len(tasks)+1)
}

func mergeParams(commonParams common.Parameters) antcolony.Parameters {
	def := antcolony.DefaultParameters()
	result := def

	if commonParams.AntCount > 0 {
		result.AntCount = commonParams.AntCount
	}
	if commonParams.InitialPheromone > 0 {
		result.InitialPheromone = commonParams.InitialPheromone
	}
	if commonParams.EvaporationRate != 0 {
		result.EvaporationRate = commonParams.EvaporationRate
	}
	if commonParams.Alpha != 0 {
		result.Alpha = commonParams.Alpha
	}
	if commonParams.Beta != 0 {
		result.Beta = commonParams.Beta
	}
	if commonParams.MaxIterations > 0 {
		result.MaxIterations = commonParams.MaxIterations
	}
	if commonParams.ConvergenceThreshold != 0 {
		result.ConvergenceThreshold = commonParams.ConvergenceThreshold
	}
	if commonParams.EliteWeight != 0 {
		result.EliteWeight = commonParams.EliteWeight
	}

	return result
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	acGraph := antcolony.Graph{
		Nodes: make([]antcolony.Node, len(req.Graph.Nodes)),
		Edges: make([]antcolony.Edge, len(req.Graph.Edges)),
	}
	for i, n := range req.Graph.Nodes {
		acGraph.Nodes[i] = antcolony.Node{ID: n.ID}
	}
	for i, e := range req.Graph.Edges {
		acGraph.Edges[i] = antcolony.Edge{From: e.From, To: e.To, Weight: e.Weight}
	}

	params := mergeParams(req.Parameters)

	solver, err := antcolony.NewSolver(acGraph, req.Start, req.End, params)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		resp := common.SubmitResponse{
			Success: false,
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	taskID := generateTaskID()
	t := &task{solver: solver}

	tasksMu.Lock()
	tasks[taskID] = t
	tasksMu.Unlock()

	isLarge := solver.IsLargeGraph()
	msg := "Task submitted successfully"
	if isLarge {
		msg += " (warning: large graph detected, computation may be slow)"
	}

	w.Header().Set("Content-Type", "application/json")
	resp := common.SubmitResponse{
		TaskID:  taskID,
		Success: true,
		Message: msg,
		IsLarge: isLarge,
	}
	json.NewEncoder(w).Encode(resp)
}

func startHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := r.URL.Query().Get("taskId")
	if taskID == "" {
		http.Error(w, "Missing taskId parameter", http.StatusBadRequest)
		return
	}

	tasksMu.RLock()
	t, ok := tasks[taskID]
	tasksMu.RUnlock()

	if !ok {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	status := t.solver.GetStatus()
	if status.Status == antcolony.StatusRunning {
		http.Error(w, "Task already running", http.StatusConflict)
		return
	}
	if status.Status == antcolony.StatusCompleted {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Task already completed",
		})
		return
	}

	go func() {
		t.solver.Solve()
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Task started",
	})
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := r.URL.Query().Get("taskId")
	if taskID == "" {
		http.Error(w, "Missing taskId parameter", http.StatusBadRequest)
		return
	}

	tasksMu.RLock()
	t, ok := tasks[taskID]
	tasksMu.RUnlock()

	if !ok {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	status := t.solver.GetStatus()

	w.Header().Set("Content-Type", "application/json")
	resp := common.StatusResponse{
		TaskID:           taskID,
		Status:           status.Status,
		Progress:         status.Progress,
		CurrentIteration: status.CurrentIteration,
		BestLength:       status.BestLength,
		Completed:        status.Status == antcolony.StatusCompleted,
	}
	json.NewEncoder(w).Encode(resp)
}

func resultHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := r.URL.Query().Get("taskId")
	if taskID == "" {
		http.Error(w, "Missing taskId parameter", http.StatusBadRequest)
		return
	}

	tasksMu.RLock()
	t, ok := tasks[taskID]
	tasksMu.RUnlock()

	if !ok {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	status := t.solver.GetStatus()
	if status.Result == nil {
		if status.Status == antcolony.StatusRunning {
			http.Error(w, "Task still running", http.StatusAccepted)
		} else if status.Status == antcolony.StatusIdle {
			http.Error(w, "Task not started", http.StatusBadRequest)
		} else {
			http.Error(w, "Result not available", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	result := status.Result
	resp := common.ResultResponse{
		TaskID:            taskID,
		Success:           result.Success,
		Error:             result.Error,
		Iterations:        result.Iterations,
		Converged:         result.Converged,
		BestLengthHistory: result.BestLengthHistory,
	}
	if result.Success {
		resp.Path = common.Path{
			Nodes:  result.Path.Nodes,
			Length: result.Path.Length,
		}
	}
	json.NewEncoder(w).Encode(resp)
}

func main() {
	portFlag := flag.String("port", "", "Port to listen on")
	flag.Parse()

	port := os.Getenv("PORT")
	if *portFlag != "" {
		port = *portFlag
	}
	if port == "" {
		port = "8103"
	}

	http.HandleFunc("/submit", submitHandler)
	http.HandleFunc("/start", startHandler)
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/result", resultHandler)

	log.Printf("Ant Colony Server starting on port %s", port)
	log.Printf("Endpoints:")
	log.Printf("  POST /submit    - Submit a graph for solving")
	log.Printf("  POST /start     - Start solving a submitted task")
	log.Printf("  GET  /status    - Query task status")
	log.Printf("  GET  /result    - Get final result")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
