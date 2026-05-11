package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"bikeshare/common"
	"bikeshare/core"
)

var (
	store      *core.Store
	billingSvc *core.BillingService
	dispatchSvc *core.DispatchService
)

func main() {
	store = core.NewStore()
	billingSvc = core.NewBillingService(store)
	dispatchSvc = core.NewDispatchService(store)

	port := getPort()

	mux := http.NewServeMux()

	mux.HandleFunc("/fences", handleFences)
	mux.HandleFunc("/fences/", handleFenceByID)

	mux.HandleFunc("/bikes", handleBikes)
	mux.HandleFunc("/bikes/", handleBikeByID)

	mux.HandleFunc("/rides", handleRides)
	mux.HandleFunc("/rides/start", handleStartRide)
	mux.HandleFunc("/rides/end", handleEndRide)

	mux.HandleFunc("/dispatch/analyze", handleAnalyzeCapacity)
	mux.HandleFunc("/dispatch/tasks", handleDispatchTasks)
	mux.HandleFunc("/dispatch/tasks/", handleDispatchTaskByID)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func getPort() string {
	port := flag.String("port", "", "Server port")
	flag.Parse()

	if *port != "" {
		return *port
	}

	if envPort := os.Getenv("PORT"); envPort != "" {
		return envPort
	}

	return "8903"
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(common.Response{Success: true, Data: data})
}

func writeError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(common.Response{Success: false, Error: err.Error()})
}

func handleFences(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listFences(w, r)
	case http.MethodPost:
		createFence(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func listFences(w http.ResponseWriter, r *http.Request) {
	fences := store.GetFences()
	resp := make([]*common.FenceResponse, 0, len(fences))
	for _, f := range fences {
		resp = append(resp, common.ToFenceResponse(f))
	}
	writeJSON(w, http.StatusOK, resp)
}

func createFence(w http.ResponseWriter, r *http.Request) {
	var req common.CreateFenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	fenceType := core.FenceType(req.Type)
	if fenceType != core.FenceTypeRecommended &&
		fenceType != core.FenceTypeForbidden &&
		fenceType != core.FenceTypeDispatch {
		writeError(w, http.StatusBadRequest, errInvalidFenceType)
		return
	}

	fence, err := store.CreateFence(req.Name, fenceType, req.MinLat, req.MaxLat, req.MinLng, req.MaxLng, req.Capacity)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusCreated, common.ToFenceResponse(fence))
}

func handleFenceByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/fences/")
	if id == "" {
		writeError(w, http.StatusBadRequest, errInvalidFenceID)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getFence(w, r, id)
	case http.MethodDelete:
		deleteFence(w, r, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func getFence(w http.ResponseWriter, r *http.Request, id string) {
	fence := store.GetFence(id)
	if fence == nil {
		writeError(w, http.StatusNotFound, core.ErrFenceNotFound)
		return
	}
	writeJSON(w, http.StatusOK, common.ToFenceResponse(fence))
}

func deleteFence(w http.ResponseWriter, r *http.Request, id string) {
	if err := store.DeleteFence(id); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

func handleBikes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listBikes(w, r)
	case http.MethodPost:
		addBike(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func listBikes(w http.ResponseWriter, r *http.Request) {
	bikes := store.GetBikes()
	resp := make([]*common.BikeResponse, 0, len(bikes))
	for _, b := range bikes {
		resp = append(resp, common.ToBikeResponse(b))
	}
	writeJSON(w, http.StatusOK, resp)
}

func addBike(w http.ResponseWriter, r *http.Request) {
	var req common.AddBikeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	bike, err := store.AddBike(req.Lat, req.Lng)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusCreated, common.ToBikeResponse(bike))
}

func handleBikeByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/bikes/")
	if id == "" {
		writeError(w, http.StatusBadRequest, errInvalidBikeID)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getBike(w, r, id)
	case http.MethodPut:
		updateBikeLocation(w, r, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func getBike(w http.ResponseWriter, r *http.Request, id string) {
	bike := store.GetBike(id)
	if bike == nil {
		writeError(w, http.StatusNotFound, core.ErrBikeNotFound)
		return
	}
	writeJSON(w, http.StatusOK, common.ToBikeResponse(bike))
}

func updateBikeLocation(w http.ResponseWriter, r *http.Request, id string) {
	var req common.UpdateBikeLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	bike, err := store.UpdateBikeLocation(id, req.Lat, req.Lng)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, common.ToBikeResponse(bike))
}

func handleRides(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	activeOnly := r.URL.Query().Get("active") == "true"

	var rides []*core.Ride
	if activeOnly {
		rides = billingSvc.GetActiveRides()
	} else {
		rides = store.GetRides()
	}

	resp := make([]*common.RideResponse, 0, len(rides))
	for _, ride := range rides {
		resp = append(resp, common.ToRideResponse(ride))
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleStartRide(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.StartRideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	ride, err := billingSvc.StartRide(req.BikeID, req.StartLat, req.StartLng)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusCreated, common.ToRideResponse(ride))
}

func handleEndRide(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.EndRideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	ride, err := billingSvc.EndRide(req.RideID, req.EndLat, req.EndLng)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, common.ToRideResponse(ride))
}

func handleAnalyzeCapacity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	statuses := dispatchSvc.AnalyzeCapacity()
	resp := make([]*common.FenceCapacityResponse, 0, len(statuses))
	for _, s := range statuses {
		resp = append(resp, common.ToFenceCapacityResponse(s))
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleDispatchTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listDispatchTasks(w, r)
	case http.MethodPost:
		generateDispatchTasks(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func listDispatchTasks(w http.ResponseWriter, r *http.Request) {
	pendingOnly := r.URL.Query().Get("pending") == "true"

	var tasks []*core.DispatchTask
	if pendingOnly {
		tasks = dispatchSvc.GetPendingTasks()
	} else {
		tasks = store.GetTasks()
	}

	resp := make([]*common.DispatchTaskResponse, 0, len(tasks))
	for _, t := range tasks {
		resp = append(resp, common.ToDispatchTaskResponse(t))
	}
	writeJSON(w, http.StatusOK, resp)
}

func generateDispatchTasks(w http.ResponseWriter, r *http.Request) {
	tasks := dispatchSvc.GenerateDispatchTasks()
	resp := make([]*common.DispatchTaskResponse, 0, len(tasks))
	for _, t := range tasks {
		resp = append(resp, common.ToDispatchTaskResponse(t))
	}
	writeJSON(w, http.StatusCreated, resp)
}

func handleDispatchTaskByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/dispatch/tasks/")
	if id == "" {
		writeError(w, http.StatusBadRequest, errInvalidTaskID)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	action := r.URL.Query().Get("action")
	if action != "complete" {
		writeError(w, http.StatusBadRequest, errInvalidAction)
		return
	}

	task, err := dispatchSvc.CompleteTask(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, common.ToDispatchTaskResponse(task))
}

func atoi(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}
