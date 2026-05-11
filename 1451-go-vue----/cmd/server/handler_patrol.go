package main

import (
	"encoding/json"
	"net/http"

	"smart-park/common"
)

func (h *Handler) handlePatrolPoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Success: false, Message: "method not allowed"})
		return
	}

	var req common.PatrolPoint
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: "invalid request body"})
		return
	}

	point, err := h.patrolService.CreatePatrolPoint(req.ID, req.Name)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, common.Response{Success: true, Data: point})
}

func (h *Handler) handlePatrolRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Success: false, Message: "method not allowed"})
		return
	}

	var req common.CreatePatrolRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: "invalid request body"})
		return
	}

	route, err := h.patrolService.CreatePatrolRoute(req.Name, req.PointIDs)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, common.Response{Success: true, Data: route})
}

func (h *Handler) handlePatrolRoutes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Success: false, Message: "method not allowed"})
		return
	}
	routes := h.patrolService.GetAllRoutes()
	writeJSON(w, http.StatusOK, common.Response{Success: true, Data: routes})
}

func (h *Handler) handlePatrolTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Success: false, Message: "method not allowed"})
		return
	}

	var req common.CreatePatrolTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: "invalid request body"})
		return
	}

	task, err := h.patrolService.CreatePatrolTask(req.RouteID, req.AssigneeID, req.Frequency, req.StartTime)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, common.Response{Success: true, Data: task})
}

func (h *Handler) handlePatrolTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Success: false, Message: "method not allowed"})
		return
	}
	tasks := h.patrolService.GetAllTasks()
	writeJSON(w, http.StatusOK, common.Response{Success: true, Data: tasks})
}

func (h *Handler) handlePatrolStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Success: false, Message: "method not allowed"})
		return
	}

	var req common.StartPatrolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: "invalid request body"})
		return
	}

	execution, err := h.patrolService.StartPatrolExecution(req.TaskID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, common.Response{Success: true, Data: execution})
}

func (h *Handler) handlePatrolCheckIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Success: false, Message: "method not allowed"})
		return
	}

	var req common.PatrolCheckInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: "invalid request body"})
		return
	}

	checkIn, err := h.patrolService.CheckIn(
		req.ExecutionID,
		req.PointID,
		req.HasAnomaly,
		req.AnomalyDesc,
		req.RelatedAccessPointID,
	)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, common.Response{Success: true, Data: checkIn})
}

func (h *Handler) handlePatrolCheckIns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Success: false, Message: "method not allowed"})
		return
	}

	executionID := r.URL.Query().Get("execution_id")
	if executionID == "" {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: "execution_id is required"})
		return
	}

	checkIns := h.patrolService.GetCheckIns(executionID)
	writeJSON(w, http.StatusOK, common.Response{Success: true, Data: checkIns})
}
