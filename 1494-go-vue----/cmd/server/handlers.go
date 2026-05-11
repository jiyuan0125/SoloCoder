package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"green-care-management/pkg/api"
	"green-care-management/pkg/core"
)

func (s *Server) routes() {
	s.mux.HandleFunc("/api/zones", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.handleGetZones(w, r)
		case http.MethodPost:
			s.handleCreateZone(w, r)
		default:
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	s.mux.HandleFunc("/api/zones/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/zones/")
		if id == "" {
			respondError(w, http.StatusBadRequest, "missing zone id")
			return
		}
		s.handleGetZone(w, r, id)
	})

	s.mux.HandleFunc("/api/plants", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.handleGetPlants(w, r)
		case http.MethodPost:
			s.handleCreatePlant(w, r)
		default:
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	s.mux.HandleFunc("/api/plants/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/plants/")
		if id == "" {
			respondError(w, http.StatusBadRequest, "missing plant id")
			return
		}
		switch r.Method {
		case http.MethodGet:
			s.handleGetPlant(w, r, id)
		case http.MethodPut:
			s.handleUpdatePlant(w, r, id)
		case http.MethodDelete:
			s.handleDeletePlant(w, r, id)
		default:
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	s.mux.HandleFunc("/api/workers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.handleGetWorkers(w, r)
		case http.MethodPost:
			s.handleCreateWorker(w, r)
		default:
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	s.mux.HandleFunc("/api/workers/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/workers/")
		if id == "" {
			respondError(w, http.StatusBadRequest, "missing worker id")
			return
		}
		switch r.Method {
		case http.MethodGet:
			s.handleGetWorker(w, r, id)
		case http.MethodPut:
			s.handleUpdateWorker(w, r, id)
		default:
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	s.mux.HandleFunc("/api/maintenance-plans", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.handleGetMaintenancePlans(w, r)
		case http.MethodPost:
			s.handleCreateMaintenancePlan(w, r)
		default:
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	s.mux.HandleFunc("/api/maintenance-plans/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/maintenance-plans/")
		if id == "" {
			respondError(w, http.StatusBadRequest, "missing plan id")
			return
		}
		s.handleGetMaintenancePlan(w, r, id)
	})

	s.mux.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.handleGetTasks(w, r)
		case http.MethodPost:
			s.handleCreateAdhocTask(w, r)
		default:
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	s.mux.HandleFunc("/api/tasks/generate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		s.handleGenerateDailyTasks(w, r)
	})

	s.mux.HandleFunc("/api/tasks/execute", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		s.handleExecuteTask(w, r)
	})

	s.mux.HandleFunc("/api/tasks/review", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		s.handleReviewTask(w, r)
	})

	s.mux.HandleFunc("/api/tasks/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
		if id == "" {
			respondError(w, http.StatusBadRequest, "missing task id")
			return
		}
		s.handleGetTask(w, r, id)
	})

	s.mux.HandleFunc("/api/cost-report", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		s.handleGenerateCostReport(w, r)
	})
}

func (s *Server) handleCreateZone(w http.ResponseWriter, r *http.Request) {
	var req api.CreateZoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	zone, err := s.service.CreateZone(req.Name, req.Description)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    toZoneResponse(zone),
	})
}

func (s *Server) handleGetZones(w http.ResponseWriter, _ *http.Request) {
	zones := s.service.GetZones()
	res := make([]*api.ZoneResponse, len(zones))
	for i, z := range zones {
		res[i] = toZoneResponse(z)
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    res,
	})
}

func (s *Server) handleGetZone(w http.ResponseWriter, _ *http.Request, id string) {
	zone, err := s.service.GetZone(id)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    toZoneResponse(zone),
	})
}

func (s *Server) handleCreatePlant(w http.ResponseWriter, r *http.Request) {
	var req api.CreatePlantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	plant, err := s.service.CreatePlant(&req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    toPlantResponse(plant),
	})
}

func (s *Server) handleGetPlants(w http.ResponseWriter, r *http.Request) {
	zoneID := r.URL.Query().Get("zone_id")
	var plants []*core.Plant
	if zoneID != "" {
		plants = s.service.GetPlantsByZone(zoneID)
	} else {
		plants = s.service.GetPlants()
	}

	res := make([]*api.PlantResponse, len(plants))
	for i, p := range plants {
		res[i] = toPlantResponse(p)
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    res,
	})
}

func (s *Server) handleGetPlant(w http.ResponseWriter, _ *http.Request, id string) {
	plant, err := s.service.GetPlant(id)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    toPlantResponse(plant),
	})
}

func (s *Server) handleUpdatePlant(w http.ResponseWriter, r *http.Request, id string) {
	var req api.UpdatePlantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	plant, err := s.service.UpdatePlant(id, &req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    toPlantResponse(plant),
	})
}

func (s *Server) handleDeletePlant(w http.ResponseWriter, _ *http.Request, id string) {
	if err := s.service.DeletePlant(id); err != nil {
		handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "plant deleted",
	})
}

func (s *Server) handleCreateWorker(w http.ResponseWriter, r *http.Request) {
	var req api.CreateWorkerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	worker, err := s.service.CreateWorker(&req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    toWorkerResponse(worker),
	})
}

func (s *Server) handleGetWorkers(w http.ResponseWriter, r *http.Request) {
	zoneID := r.URL.Query().Get("zone_id")
	var workers []*core.Worker
	if zoneID != "" {
		workers = s.service.GetWorkers()
		var filtered []*core.Worker
		for _, w := range workers {
			if w.ZoneID == zoneID {
				filtered = append(filtered, w)
			}
		}
		workers = filtered
	} else {
		workers = s.service.GetWorkers()
	}

	res := make([]*api.WorkerResponse, len(workers))
	for i, wk := range workers {
		res[i] = toWorkerResponse(wk)
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    res,
	})
}

func (s *Server) handleGetWorker(w http.ResponseWriter, _ *http.Request, id string) {
	worker, err := s.service.GetWorker(id)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    toWorkerResponse(worker),
	})
}

func (s *Server) handleUpdateWorker(w http.ResponseWriter, r *http.Request, id string) {
	var req api.UpdateWorkerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	worker, err := s.service.UpdateWorker(id, &req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    toWorkerResponse(worker),
	})
}

func (s *Server) handleCreateMaintenancePlan(w http.ResponseWriter, r *http.Request) {
	var req api.CreateMaintenancePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	plan, err := s.service.CreateMaintenancePlan(&req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    toMaintenancePlanResponse(plan),
	})
}

func (s *Server) handleGetMaintenancePlans(w http.ResponseWriter, _ *http.Request) {
	plans := s.service.GetMaintenancePlans()
	res := make([]*api.MaintenancePlanResponse, len(plans))
	for i, p := range plans {
		res[i] = toMaintenancePlanResponse(p)
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    res,
	})
}

func (s *Server) handleGetMaintenancePlan(w http.ResponseWriter, _ *http.Request, id string) {
	plan, err := s.service.GetMaintenancePlan(id)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    toMaintenancePlanResponse(plan),
	})
}

func (s *Server) handleCreateAdhocTask(w http.ResponseWriter, r *http.Request) {
	var req api.CreateAdhocTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := s.service.CreateAdhocTask(&req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    toTaskResponse(task),
	})
}

func (s *Server) handleGetTasks(w http.ResponseWriter, r *http.Request) {
	zoneID := r.URL.Query().Get("zone_id")
	workerID := r.URL.Query().Get("worker_id")

	var tasks []*core.Task
	if workerID != "" {
		tasks = s.service.GetTasksByWorker(workerID)
	} else if zoneID != "" {
		tasks = s.service.GetTasksByZone(zoneID)
	} else {
		tasks = s.service.GetTasks()
	}

	res := make([]*api.TaskResponse, len(tasks))
	for i, t := range tasks {
		res[i] = toTaskResponse(t)
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    res,
	})
}

func (s *Server) handleGetTask(w http.ResponseWriter, _ *http.Request, id string) {
	task, err := s.service.GetTask(id)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    toTaskResponse(task),
	})
}

func (s *Server) handleGenerateDailyTasks(w http.ResponseWriter, r *http.Request) {
	var req api.GenerateTasksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	date := req.Date
	if date.IsZero() {
		date = time.Now()
	}

	tasks, err := s.service.GenerateDailyTasks(date)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	res := make([]*api.TaskResponse, len(tasks))
	for i, t := range tasks {
		res[i] = toTaskResponse(t)
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    res,
	})
}

func (s *Server) handleExecuteTask(w http.ResponseWriter, r *http.Request) {
	var req api.ExecuteTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := s.service.ExecuteTask(&req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    toTaskResponse(task),
	})
}

func (s *Server) handleReviewTask(w http.ResponseWriter, r *http.Request) {
	var req api.ReviewTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := s.service.ReviewTask(&req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    toTaskResponse(task),
	})
}

func (s *Server) handleGenerateCostReport(w http.ResponseWriter, r *http.Request) {
	var req api.CostReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	year := req.Year
	month := req.Month
	if year == 0 {
		year = time.Now().Year()
	}
	if month == 0 {
		month = int(time.Now().Month())
	}

	report, err := s.service.GenerateCostReport(year, month)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    report,
	})
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, core.ErrNotFound):
		respondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, core.ErrInvalidRequest):
		respondError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, core.ErrAlreadyExists):
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, core.ErrCapacityExceeded):
		respondError(w, http.StatusServiceUnavailable, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, err.Error())
	}
}

func toZoneResponse(z *core.Zone) *api.ZoneResponse {
	return &api.ZoneResponse{
		ID:          z.ID,
		Name:        z.Name,
		Description: z.Description,
		CreatedAt:   z.CreatedAt,
		UpdatedAt:   z.UpdatedAt,
	}
}

func toPlantResponse(p *core.Plant) *api.PlantResponse {
	return &api.PlantResponse{
		ID:           p.ID,
		ZoneID:       p.ZoneID,
		Name:         p.Name,
		Variety:      p.Variety,
		PlantType:    p.PlantType,
		PlantingDate: p.PlantingDate,
		Location:     p.Location,
		HealthStatus: p.HealthStatus,
		Description:  p.Description,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

func toWorkerResponse(w *core.Worker) *api.WorkerResponse {
	return &api.WorkerResponse{
		ID:         w.ID,
		EmployeeID: w.EmployeeID,
		Name:       w.Name,
		SkillLevel: w.SkillLevel,
		ZoneID:     w.ZoneID,
		CreatedAt:  w.CreatedAt,
		UpdatedAt:  w.UpdatedAt,
	}
}

func toMaintenancePlanResponse(p *core.MaintenancePlan) *api.MaintenancePlanResponse {
	return &api.MaintenancePlanResponse{
		ID:        p.ID,
		ZoneID:    p.ZoneID,
		PlantType: p.PlantType,
		Items:     p.Items,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func toTaskResponse(t *core.Task) *api.TaskResponse {
	return &api.TaskResponse{
		ID:               t.ID,
		TaskType:         t.TaskType,
		ZoneID:           t.ZoneID,
		PlantIDs:         t.PlantIDs,
		MaintenanceType:  t.MaintenanceType,
		Title:            t.Title,
		Description:      t.Description,
		Status:           t.Status,
		EstimatedHours:   t.EstimatedHours,
		AssignedWorkerID: t.AssignedWorkerID,
		ActualHours:      t.ActualHours,
		MaterialCost:     t.MaterialCost,
		WorkerNotes:      t.WorkerNotes,
		RejectCount:      t.RejectCount,
		RejectReasons:    t.RejectReasons,
		Priority:         t.Priority,
		DueDate:          t.DueDate,
		CreatedAt:        t.CreatedAt,
		UpdatedAt:        t.UpdatedAt,
	}
}
