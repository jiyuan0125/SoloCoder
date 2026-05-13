package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"go-cicd-pipeline/internal/amount"
	"go-cicd-pipeline/internal/artifact"
	"go-cicd-pipeline/internal/engine"
	"go-cicd-pipeline/internal/logmanager"
	"go-cicd-pipeline/internal/rollback"
	"go-cicd-pipeline/internal/store"
	"go-cicd-pipeline/internal/types"
	"go-cicd-pipeline/internal/utils"
)

type Server struct {
	store        *store.Store
	engine       *engine.Engine
	logManager   *logmanager.LogManager
	artifactMgr  *artifact.Manager
	amountMgr    *amount.Manager
	rollbackMgr  *rollback.Manager
}

func New(s *store.Store, e *engine.Engine, lm *logmanager.LogManager, am *artifact.Manager) *Server {
	return &Server{
		store:       s,
		engine:      e,
		logManager:  lm,
		artifactMgr: am,
		amountMgr:   amount.New(s),
		rollbackMgr: rollback.New(s, e),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/pipelines", s.handlePipelines)
	mux.HandleFunc("/api/pipelines/", s.handlePipelineByID)

	mux.HandleFunc("/api/executions/", s.handleExecutionByID)

	mux.HandleFunc("/api/tasks/", s.handleTaskByID)

	mux.HandleFunc("/api/artifacts/", s.handleArtifactByID)

	mux.HandleFunc("/api/rollback/", s.handleRollbackByExecutionID)

	return mux
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func readJSON(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.Unmarshal(body, v)
}

func getLastSegment(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func getSegment(path string, index int) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if index < len(parts) {
		return parts[index]
	}
	return ""
}

type CreatePipelineRequest struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Phases      []PhaseRequest      `json:"phases"`
	Triggers    []types.TriggerType `json:"triggers"`
	TotalAmount float64             `json:"total_amount"`
}

type PhaseRequest struct {
	Type          types.PhaseType `json:"type"`
	Name          string          `json:"name"`
	TaskMode      types.TaskMode  `json:"task_mode"`
	Tasks         []TaskRequest   `json:"tasks"`
	PlannedAmount float64         `json:"planned_amount"`
}

type TaskRequest struct {
	Name            string               `json:"name"`
	Script          string               `json:"script"`
	TimeoutSec      int                  `json:"timeout_sec"`
	FailureStrategy types.FailureStrategy `json:"failure_strategy"`
}

type TriggerExecutionRequest struct {
	TriggerType types.TriggerType `json:"trigger_type"`
	TargetEnv   types.TargetEnv   `json:"target_env"`
	TargetPhase types.PhaseType   `json:"target_phase"`
}

type AdjustAmountRequest struct {
	NewTotal float64 `json:"new_total"`
}

func (s *Server) handlePipelines(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		pipelines, err := s.store.ListPipelines()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, pipelines)

	case http.MethodPost:
		var req CreatePipelineRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		pipeline := &types.Pipeline{
			ID:          utils.GenerateID(),
			Name:        req.Name,
			Description: req.Description,
			TotalAmount: req.TotalAmount,
			Triggers:    req.Triggers,
			CreatedAt:   utils.Now(),
			UpdatedAt:   utils.Now(),
		}

		for i, pr := range req.Phases {
			phase := types.Phase{
				ID:            utils.GenerateID(),
				PipelineID:    pipeline.ID,
				Type:          pr.Type,
				Name:          pr.Name,
				TaskMode:      pr.TaskMode,
				PlannedAmount: pr.PlannedAmount,
				OrderIndex:    i,
			}

			for j, tr := range pr.Tasks {
				task := types.Task{
					ID:               utils.GenerateID(),
					PhaseID:          phase.ID,
					Name:             tr.Name,
					Script:           tr.Script,
					TimeoutSec:       tr.TimeoutSec,
					FailureStrategy:  tr.FailureStrategy,
					OrderIndex:       j,
				}
				phase.Tasks = append(phase.Tasks, task)
			}

			pipeline.Phases = append(pipeline.Phases, phase)
		}

		if err := s.store.CreatePipeline(pipeline); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, pipeline)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handlePipelineByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.Contains(path, "/executions") {
		s.handlePipelineExecutions(w, r)
		return
	}

	if strings.Contains(path, "/trigger") {
		s.handlePipelineTrigger(w, r)
		return
	}

	if strings.Contains(path, "/amount") {
		s.handlePipelineAmount(w, r)
		return
	}

	pipelineID := getSegment(path, 2)
	if pipelineID == "" {
		writeError(w, http.StatusBadRequest, "pipeline id required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		pipeline, err := s.store.GetPipeline(pipelineID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if pipeline == nil {
			writeError(w, http.StatusNotFound, "pipeline not found")
			return
		}
		writeJSON(w, http.StatusOK, pipeline)

	case http.MethodPut:
		var req CreatePipelineRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		pipeline, err := s.store.GetPipeline(pipelineID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if pipeline == nil {
			writeError(w, http.StatusNotFound, "pipeline not found")
			return
		}

		pipeline.Name = req.Name
		pipeline.Description = req.Description
		pipeline.TotalAmount = req.TotalAmount
		pipeline.Triggers = req.Triggers
		pipeline.UpdatedAt = utils.Now()
		pipeline.Phases = nil

		for i, pr := range req.Phases {
			phase := types.Phase{
				ID:            utils.GenerateID(),
				PipelineID:    pipelineID,
				Type:          pr.Type,
				Name:          pr.Name,
				TaskMode:      pr.TaskMode,
				PlannedAmount: pr.PlannedAmount,
				OrderIndex:    i,
			}

			for j, tr := range pr.Tasks {
				task := types.Task{
					ID:               utils.GenerateID(),
					PhaseID:          phase.ID,
					Name:             tr.Name,
					Script:           tr.Script,
					TimeoutSec:       tr.TimeoutSec,
					FailureStrategy:  tr.FailureStrategy,
					OrderIndex:       j,
				}
				phase.Tasks = append(phase.Tasks, task)
			}

			pipeline.Phases = append(pipeline.Phases, phase)
		}

		if err := s.store.UpdatePipeline(pipeline); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, pipeline)

	case http.MethodDelete:
		pipeline, err := s.store.GetPipeline(pipelineID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if pipeline == nil {
			writeError(w, http.StatusNotFound, "pipeline not found")
			return
		}

		if err := s.store.DeletePipeline(pipelineID); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handlePipelineExecutions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	pipelineID := getSegment(r.URL.Path, 2)
	if pipelineID == "" {
		writeError(w, http.StatusBadRequest, "pipeline id required")
		return
	}

	pipeline, err := s.store.GetPipeline(pipelineID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if pipeline == nil {
		writeError(w, http.StatusNotFound, "pipeline not found")
		return
	}

	executions, err := s.store.ListExecutions(pipelineID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, executions)
}

func (s *Server) handlePipelineTrigger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	pipelineID := getSegment(r.URL.Path, 2)
	if pipelineID == "" {
		writeError(w, http.StatusBadRequest, "pipeline id required")
		return
	}

	var req TriggerExecutionRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	execution, err := s.engine.CreateExecution(pipelineID, req.TriggerType, req.TargetEnv, req.TargetPhase, false, "")
	if err != nil {
		if err.Error() == "pipeline already has running execution" {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		if err.Error() == "pipeline not found" {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "cannot skip") || strings.Contains(err.Error(), "production deployment") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := s.engine.StartExecution(execution.ID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, execution)
}

func (s *Server) handlePipelineAmount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	pipelineID := getSegment(r.URL.Path, 2)
	if pipelineID == "" {
		writeError(w, http.StatusBadRequest, "pipeline id required")
		return
	}

	var req AdjustAmountRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.amountMgr.AdjustTotalAmount(pipelineID, req.NewTotal); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	pipeline, err := s.store.GetPipeline(pipelineID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, pipeline)
}

func (s *Server) handleExecutionByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasSuffix(path, "/cancel") {
		s.handleExecutionCancel(w, r)
		return
	}

	if strings.HasSuffix(path, "/artifacts") {
		s.handleExecutionArtifacts(w, r)
		return
	}

	if strings.HasSuffix(path, "/upload") {
		s.handleExecutionUpload(w, r)
		return
	}

	if strings.HasSuffix(path, "/cleanup") {
		s.handleExecutionCleanup(w, r)
		return
	}

	executionID := getSegment(path, 2)
	if executionID == "" {
		writeError(w, http.StatusBadRequest, "execution id required")
		return
	}

	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	execution, err := s.store.GetExecution(executionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if execution == nil {
		writeError(w, http.StatusNotFound, "execution not found")
		return
	}

	writeJSON(w, http.StatusOK, execution)
}

func (s *Server) handleExecutionCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	executionID := getSegment(r.URL.Path, 2)
	if executionID == "" {
		writeError(w, http.StatusBadRequest, "execution id required")
		return
	}

	execution, err := s.store.GetExecution(executionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if execution == nil {
		writeError(w, http.StatusNotFound, "execution not found")
		return
	}

	s.engine.CancelExecution(executionID)
	writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

func (s *Server) handleExecutionArtifacts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	executionID := getSegment(r.URL.Path, 2)
	if executionID == "" {
		writeError(w, http.StatusBadRequest, "execution id required")
		return
	}

	artifacts, err := s.artifactMgr.List(executionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, artifacts)
}

func (s *Server) handleExecutionUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	executionID := getSegment(r.URL.Path, 2)
	if executionID == "" {
		writeError(w, http.StatusBadRequest, "execution id required")
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "name parameter required")
		return
	}

	taskResultID := r.URL.Query().Get("task_result_id")

	artifact, err := s.artifactMgr.Upload(executionID, taskResultID, name, r.Body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, artifact)
}

func (s *Server) handleExecutionCleanup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	executionID := getSegment(r.URL.Path, 2)
	if executionID == "" {
		writeError(w, http.StatusBadRequest, "execution id required")
		return
	}

	records, err := s.store.ListCleanupRecords(executionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (s *Server) handleTaskByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	taskResultID := getSegment(r.URL.Path, 2)
	if taskResultID == "" {
		writeError(w, http.StatusBadRequest, "task result id required")
		return
	}

	if strings.HasSuffix(r.URL.Path, "/log") {
		s.handleTaskLog(w, r, taskResultID)
		return
	}

	writeError(w, http.StatusNotFound, "not found")
}

func (s *Server) handleTaskLog(w http.ResponseWriter, r *http.Request, taskResultID string) {
	db := s.store.DB()

	var logPath sql.NullString
	err := db.QueryRow(`
		SELECT log_path FROM task_results WHERE id = ?
	`, taskResultID).Scan(&logPath)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "task result not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !logPath.Valid || logPath.String == "" {
		writeError(w, http.StatusNotFound, "log not found")
		return
	}

	logContent, err := s.logManager.ReadLog(logPath.String)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%s", filepath.Base(logPath.String)))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(logContent))
}

func (s *Server) handleArtifactByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	artifactID := getSegment(r.URL.Path, 2)
	if artifactID == "" {
		writeError(w, http.StatusBadRequest, "artifact id required")
		return
	}

	reader, artifact, err := s.artifactMgr.Download(artifactID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if artifact == nil {
		writeError(w, http.StatusNotFound, "artifact not found")
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", artifact.Name))
	w.WriteHeader(http.StatusOK)
	io.Copy(w, reader)
}

func (s *Server) handleRollbackByExecutionID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	executionID := getSegment(r.URL.Path, 2)
	if executionID == "" {
		writeError(w, http.StatusBadRequest, "execution id required")
		return
	}

	rollbackExec, err := s.rollbackMgr.RollbackTo(executionID)
	if err != nil {
		if err.Error() == "pipeline already has running execution" {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := s.engine.StartExecution(rollbackExec.ID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, rollbackExec)
}
