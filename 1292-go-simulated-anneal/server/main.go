package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/tsp-simulated-anneal/annealing"
	"github.com/tsp-simulated-anneal/common"
)

var (
	taskManager *TaskManager
	port        string
)

func main() {
	flag.StringVar(&port, "port", "", "Server port (e.g., 8080)")
	flag.Parse()

	if port == "" {
		port = os.Getenv("TSP_SERVER_PORT")
		if port == "" {
			port = "8101"
		}
	}

	taskManager = NewTaskManager()

	http.HandleFunc("/solve", handleSolve)
	http.HandleFunc("/progress", handleProgress)
	http.HandleFunc("/result", handleResult)
	http.HandleFunc("/download", handleDownload)

	fmt.Printf("TSP Solver server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}

func handleSolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.SolveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Cities) == 0 {
		writeJSONError(w, "No cities provided", http.StatusBadRequest)
		return
	}

	taskID := uuid.New().String()
	config := common.SolverConfig{}
	if req.Config != nil {
		config = *req.Config
	}

	task := taskManager.CreateTask(taskID, req.Cities, config)

	go runSolver(task)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.SolveResponse{TaskID: taskID})
}

func handleProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := r.URL.Query().Get("taskId")
	if taskID == "" {
		writeJSONError(w, "taskId parameter is required", http.StatusBadRequest)
		return
	}

	task, exists := taskManager.GetTask(taskID)
	if !exists {
		writeJSONError(w, "Task not found", http.StatusNotFound)
		return
	}

	response := buildProgressResponse(task)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := r.URL.Query().Get("taskId")
	if taskID == "" {
		writeJSONError(w, "taskId parameter is required", http.StatusBadRequest)
		return
	}

	task, exists := taskManager.GetTask(taskID)
	if !exists {
		writeJSONError(w, "Task not found", http.StatusNotFound)
		return
	}

	task.Mu.RLock()
	status := task.Status
	result := task.Result
	err := task.Error
	task.Mu.RUnlock()

	if status == TaskStatusRunning {
		response := buildProgressResponse(task)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	if status == TaskStatusFailed {
		writeJSONError(w, fmt.Sprintf("Solver failed: %v", err), http.StatusInternalServerError)
		return
	}

	response := buildResultResponse(task, result)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := r.URL.Query().Get("taskId")
	if taskID == "" {
		writeJSONError(w, "taskId parameter is required", http.StatusBadRequest)
		return
	}

	task, exists := taskManager.GetTask(taskID)
	if !exists {
		writeJSONError(w, "Task not found", http.StatusNotFound)
		return
	}

	task.Mu.RLock()
	status := task.Status
	result := task.Result
	cities := task.Cities
	task.Mu.RUnlock()

	if status != TaskStatusCompleted || result == nil {
		writeJSONError(w, "Task not completed yet", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=tsp_result_%s.csv", taskID))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	writer.Write([]string{"Step", "City", "Latitude", "Longitude", "X", "Y"})
	for i, idx := range result.Path {
		city := cities[idx]
		writer.Write([]string{
			strconv.Itoa(i + 1),
			city.Name,
			formatFloat(city.Latitude),
			formatFloat(city.Longitude),
			formatFloat(city.X),
			formatFloat(city.Y),
		})
	}

	if len(result.Path) > 0 {
		firstCity := cities[result.Path[0]]
		writer.Write([]string{
			strconv.Itoa(len(result.Path) + 1),
			firstCity.Name + " (Return)",
			formatFloat(firstCity.Latitude),
			formatFloat(firstCity.Longitude),
			formatFloat(firstCity.X),
			formatFloat(firstCity.Y),
		})
	}

	writer.Write([]string{""})
	writer.Write([]string{"Total Distance:", formatFloat(result.TotalDistance)})
	writer.Write([]string{"Initial Distance:", formatFloat(result.InitialDistance)})
	writer.Write([]string{"Improvement:", fmt.Sprintf("%.2f%%", result.ImprovementPct)})
	writer.Write([]string{"Duration:", result.Duration.String()})
	writer.Write([]string{"Total Iterations:", strconv.Itoa(result.TotalIterations)})
	writer.Write([]string{"Random Seed:", strconv.FormatInt(result.RandomSeed, 10)})
}

func runSolver(task *Task) {
	annealingCities := make([]annealing.City, len(task.Cities))
	for i, c := range task.Cities {
		annealingCities[i] = annealing.City{
			Name:           c.Name,
			Latitude:       c.Latitude,
			Longitude:      c.Longitude,
			X:              c.X,
			Y:              c.Y,
			CoordinateType: annealing.CoordinateType(c.CoordinateType),
		}
	}

	solverConfig := annealing.SolverConfig{
		InitialTemperature:       task.Config.InitialTemperature,
		FinalTemperature:         task.Config.FinalTemperature,
		CoolingRate:              task.Config.CoolingRate,
		IterationsPerTemperature: task.Config.IterationsPerTemperature,
		AdaptiveCooling:          task.Config.AdaptiveCooling,
		AdaptiveThreshold:        task.Config.AdaptiveThreshold,
		RandomSeed:               task.Config.RandomSeed,
	}

	solver, err := annealing.NewSolver(annealingCities, solverConfig)
	if err != nil {
		taskManager.SetTaskError(task.ID, err)
		return
	}

	task.Mu.Lock()
	task.Solver = solver
	task.Mu.Unlock()

	result, err := solver.Solve()
	if err != nil {
		taskManager.SetTaskError(task.ID, err)
		return
	}

	taskManager.SetTaskResult(task.ID, result)
}

func buildProgressResponse(task *Task) common.ProgressResponse {
	task.Mu.RLock()
	defer task.Mu.RUnlock()

	response := common.ProgressResponse{
		TaskID: task.ID,
		Status: string(task.Status),
	}

	if task.Solver != nil {
		progress := task.Solver.GetProgress()
		response.CurrentTemperature = progress.CurrentTemperature
		response.CurrentIteration = progress.CurrentIteration
		response.CurrentDistance = progress.CurrentDistance
		response.BestDistance = progress.BestDistance
		response.InitialDistance = progress.InitialDistance
		response.ImprovementPct = progress.ImprovementPct
	}

	if task.Status == TaskStatusCompleted && task.Result != nil {
		response = fillResultFields(response, task.Cities, task.Result)
	}

	return response
}

func buildResultResponse(task *Task, result *annealing.SolverResult) common.ProgressResponse {
	response := common.ProgressResponse{
		TaskID: task.ID,
		Status: string(TaskStatusCompleted),
	}
	return fillResultFields(response, task.Cities, result)
}

func fillResultFields(response common.ProgressResponse, cities []common.City, result *annealing.SolverResult) common.ProgressResponse {
	response.Path = result.Path
	response.CityNames = make([]string, len(result.Path))
	for i, idx := range result.Path {
		response.CityNames[i] = cities[idx].Name
	}
	response.TotalDistance = result.TotalDistance
	response.InitialDistance = result.InitialDistance
	response.ImprovementPct = result.ImprovementPct
	response.Duration = result.Duration.Milliseconds()
	response.DurationStr = result.Duration.String()
	response.TotalIterations = result.TotalIterations
	response.RandomSeed = result.RandomSeed

	convergence := make([]common.ConvergencePoint, len(result.Convergence))
	for i, cp := range result.Convergence {
		convergence[i] = common.ConvergencePoint{
			Temperature:    cp.Temperature,
			ObjectiveValue: cp.ObjectiveValue,
		}
	}
	response.Convergence = convergence

	return response
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(common.ErrorResponse{Error: message})
}

func formatFloat(f float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.6f", f), "0"), ".")
}
