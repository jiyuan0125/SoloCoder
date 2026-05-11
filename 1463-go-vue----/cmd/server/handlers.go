package main

import (
	"encoding/json"
	"net/http"

	"safetymanager/internal/api"
	"safetymanager/internal/core"
)

type Server struct {
	service *core.Service
}

func NewServer() *Server {
	return &Server{
		service: core.NewService(),
	}
}

func writeSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(api.Response{
		Success: true,
		Data:    data,
	})
}

func writeError(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(api.Response{
		Success: false,
		Error:   err.Error(),
	})
}

func (s *Server) handleCreateZone(w http.ResponseWriter, r *http.Request) {
	var req api.CreateZoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	zone, err := s.service.CreateZone(req.Name, req.ParentID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeSuccess(w, zone)
}

func (s *Server) handleListZones(w http.ResponseWriter, r *http.Request) {
	zones, err := s.service.GetAllZones()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeSuccess(w, zones)
}

func (s *Server) handleCreatePlan(w http.ResponseWriter, r *http.Request) {
	var req api.CreateInspectionPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	plan, err := s.service.CreateInspectionPlan(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeSuccess(w, plan)
}

func (s *Server) handleListPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := s.service.GetAllInspectionPlans()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeSuccess(w, plans)
}

func (s *Server) handleGenerateTasks(w http.ResponseWriter, r *http.Request) {
	if err := s.service.GenerateTasks(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeSuccess(w, map[string]string{"status": "tasks_generated"})
}

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.service.GetAllInspectionTasks()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeSuccess(w, tasks)
}

func (s *Server) handleSubmitInspection(w http.ResponseWriter, r *http.Request) {
	var req api.SubmitInspectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	hazards, err := s.service.SubmitInspection(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeSuccess(w, hazards)
}

func (s *Server) handleListHazards(w http.ResponseWriter, r *http.Request) {
	hazards, err := s.service.GetAllHazards()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeSuccess(w, hazards)
}

func (s *Server) handleSubmitRemediation(w http.ResponseWriter, r *http.Request) {
	var req api.SubmitRemediationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := s.service.SubmitRemediation(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeSuccess(w, map[string]string{"status": "remediation_submitted"})
}

func (s *Server) handleReviewRemediation(w http.ResponseWriter, r *http.Request) {
	var req api.ReviewRemediationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := s.service.ReviewRemediation(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeSuccess(w, map[string]string{"status": "review_completed"})
}

func (s *Server) handleRequestLevelChange(w http.ResponseWriter, r *http.Request) {
	var req api.RequestLevelChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	levelReq, err := s.service.RequestLevelChange(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeSuccess(w, levelReq)
}

func (s *Server) handleReviewLevelChange(w http.ResponseWriter, r *http.Request) {
	var req api.ReviewLevelChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := s.service.ReviewLevelChange(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeSuccess(w, map[string]string{"status": "level_change_reviewed"})
}

func (s *Server) handleCheckEscalations(w http.ResponseWriter, r *http.Request) {
	escalated := s.service.CheckEscalations()
	writeSuccess(w, escalated)
}

func (s *Server) handleListAuditLogs(w http.ResponseWriter, r *http.Request) {
	logs := s.service.GetAllAuditLogs()
	writeSuccess(w, logs)
}

func (s *Server) handleExportCriticalLogs(w http.ResponseWriter, r *http.Request) {
	logs := s.service.ExportCriticalHazardLogs()
	writeSuccess(w, logs)
}
