package main

import (
	"cleaning-service/internal/common"
	"cleaning-service/internal/core"
	"encoding/json"
	"net/http"
	"strings"
)

type Handler struct {
	service *core.Service
}

func NewHandler(service *core.Service) *Handler {
	return &Handler{service: service}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(common.Response{
		Code:    status,
		Message: "success",
		Data:    data,
	})
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(common.Response{
		Code:    status,
		Message: message,
	})
}

func (h *Handler) HandleClients(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req common.CreateClientRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		client, err := h.service.CreateClient(&req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, client)
	case http.MethodGet:
		clients := h.service.ListClients()
		writeJSON(w, http.StatusOK, clients)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) HandleClientByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/clients/")
	client, err := h.service.GetClient(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, client)
}

func (h *Handler) HandleZones(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req common.CreateZoneRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		zone, err := h.service.CreateZone(&req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, zone)
	case http.MethodGet:
		zones := h.service.ListZones()
		writeJSON(w, http.StatusOK, zones)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) HandleServiceAreas(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req common.CreateServiceAreaRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		area, err := h.service.CreateServiceArea(&req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, area)
	case http.MethodGet:
		areas := h.service.ListServiceAreas()
		writeJSON(w, http.StatusOK, areas)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) HandleTeams(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req common.CreateTeamRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		team, err := h.service.CreateTeam(&req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, team)
	case http.MethodGet:
		teams := h.service.ListTeams()
		writeJSON(w, http.StatusOK, teams)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) HandleCleaners(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req common.CreateCleanerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		cleaner, err := h.service.CreateCleaner(&req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, cleaner)
	case http.MethodGet:
		cleaners := h.service.ListCleaners()
		writeJSON(w, http.StatusOK, cleaners)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) HandleSchedules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req common.CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	schedule, err := h.service.CreateSchedule(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, schedule)
}

func (h *Handler) HandleLeaves(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req common.CreateLeaveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		leave, err := h.service.CreateLeaveRequest(&req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, leave)
	case http.MethodGet:
		leaves := h.service.ListLeaves()
		writeJSON(w, http.StatusOK, leaves)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) HandleLeaveByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/leaves/")
	var req common.ApproveLeaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	leave, err := h.service.ApproveLeave(id, req.Approved)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, leave)
}

func (h *Handler) HandleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	tasks := h.service.ListTasks()
	writeJSON(w, http.StatusOK, tasks)
}

func (h *Handler) HandleGenerateTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req common.GenerateTasksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := h.service.GenerateWeeklyTasks(req.WeekStart, req.WeekEnd)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) HandleCompleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req common.CompleteTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	task, err := h.service.CompleteTask(req.TaskID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *Handler) HandleQualityChecks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req common.CreateQualityCheckRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		qc, err := h.service.CreateQualityCheck(&req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, qc)
	case http.MethodGet:
		qcs := h.service.ListQualityChecks()
		writeJSON(w, http.StatusOK, qcs)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) HandleTodos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	todos := h.service.ListTodos()
	writeJSON(w, http.StatusOK, todos)
}

func (h *Handler) HandleTodoByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/todos/")
	todo, err := h.service.CompleteTodo(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, todo)
}
