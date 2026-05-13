package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"ssh-tunnel-manager/database"
	"ssh-tunnel-manager/manager"
	"ssh-tunnel-manager/models"
	"ssh-tunnel-manager/tunnel"
	"ssh-tunnel-manager/workflow"
)

type Server struct {
	db       *database.DB
	manager  *manager.Manager
	workflow *workflow.Service
}

func NewServer(db *database.DB, mgr *manager.Manager, wf *workflow.Service) *Server {
	return &Server{
		db:       db,
		manager:  mgr,
		workflow: wf,
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

type CreateTunnelRequest struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	SSHServer   string `json:"ssh_server"`
	SSHUser     string `json:"ssh_user"`
	AuthType    string `json:"auth_type"`
	AuthData    string `json:"auth_data"`
	LocalPort   int    `json:"local_port"`
	RemoteHost  string `json:"remote_host"`
	RemotePort  int    `json:"remote_port"`
	Operator    string `json:"operator"`
}

func (s *Server) CreateTunnel(w http.ResponseWriter, r *http.Request) {
	var req CreateTunnelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.SSHServer == "" || req.SSHUser == "" || req.RemoteHost == "" {
		writeError(w, http.StatusBadRequest, "missing required fields")
		return
	}

	if req.LocalPort <= 0 || req.LocalPort > 65535 || req.RemotePort <= 0 || req.RemotePort > 65535 {
		writeError(w, http.StatusBadRequest, "invalid port number")
		return
	}

	tunnelType := models.TunnelType(req.Type)
	if tunnelType != models.TunnelTypeLocal && tunnelType != models.TunnelTypeRemote {
		writeError(w, http.StatusBadRequest, "invalid tunnel type, must be 'local' or 'remote'")
		return
	}

	authType := models.AuthType(req.AuthType)
	if authType != models.AuthTypePassword && authType != models.AuthTypeKey {
		writeError(w, http.StatusBadRequest, "invalid auth type, must be 'password' or 'private_key'")
		return
	}

	config := &models.TunnelConfig{
		Name:        req.Name,
		Type:        tunnelType,
		SSHServer:   req.SSHServer,
		SSHUser:     req.SSHUser,
		AuthType:    authType,
		AuthData:    req.AuthData,
		LocalPort:   req.LocalPort,
		RemoteHost:  req.RemoteHost,
		RemotePort:  req.RemotePort,
	}

	configID, err := s.db.CreateTunnelConfig(config)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create tunnel config")
		return
	}
	config.ID = configID

	operator := req.Operator
	if operator == "" {
		operator = "api"
	}
	recordID, err := s.workflow.CreateRecord(configID, operator)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create workflow record")
		return
	}

	response := map[string]interface{}{
		"id":          configID,
		"config":      config,
		"record_id":   recordID,
	}
	writeJSON(w, http.StatusCreated, response)
}

func (s *Server) ListTunnels(w http.ResponseWriter, r *http.Request) {
	configs, err := s.db.ListTunnelConfigs()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tunnels")
		return
	}

	type TunnelInfo struct {
		*models.TunnelConfig
		Status string `json:"status,omitempty"`
	}

	result := make([]TunnelInfo, len(configs))
	for i, cfg := range configs {
		state, _, _ := s.manager.GetTunnelStatus(cfg.ID)
		status := string(models.TunnelStatusStopped)
		if state != nil {
			status = string(state.Status)
		}
		result[i] = TunnelInfo{
			TunnelConfig: cfg,
			Status:       status,
		}
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) GetTunnel(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/tunnels/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tunnel ID")
		return
	}

	config, err := s.db.GetTunnelConfig(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get tunnel")
		return
	}
	if config == nil {
		writeError(w, http.StatusNotFound, "tunnel not found")
		return
	}

	state, stats, _ := s.manager.GetTunnelStatus(id)
	response := map[string]interface{}{
		"config": config,
		"state":  state,
		"stats":  stats,
	}
	writeJSON(w, http.StatusOK, response)
}

type StartTunnelRequest struct {
	Operator string `json:"operator"`
}

func (s *Server) StartTunnel(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/tunnels/"), "/start")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tunnel ID")
		return
	}

	config, err := s.db.GetTunnelConfig(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check tunnel")
		return
	}
	if config == nil {
		writeError(w, http.StatusNotFound, "tunnel not found")
		return
	}

	inUse, pid, err := tunnel.IsPortInUse(config.LocalPort)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check port")
		return
	}
	if inUse {
		response := map[string]interface{}{
			"error":      fmt.Sprintf("port %d already in use", config.LocalPort),
			"local_port": config.LocalPort,
			"process_id": pid,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(response)
		return
	}

	err = s.manager.StartTunnel(id)
	if err != nil {
		if strings.Contains(err.Error(), "port") {
			response := map[string]interface{}{
				"error":      err.Error(),
				"local_port": config.LocalPort,
				"process_id": 0,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(response)
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var req StartTunnelRequest
	json.NewDecoder(r.Body).Decode(&req)
	operator := req.Operator
	if operator == "" {
		operator = "api"
	}

	records, _ := s.workflow.ListRecords()
	for _, rec := range records {
		if rec.TunnelConfigID == id {
			if rec.Status == models.ApprovalApproved {
				s.workflow.StartExecution(rec.ID, operator)
			}
			break
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "starting"})
}

func (s *Server) StopTunnel(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/tunnels/"), "/stop")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tunnel ID")
		return
	}

	config, err := s.db.GetTunnelConfig(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check tunnel")
		return
	}
	if config == nil {
		writeError(w, http.StatusNotFound, "tunnel not found")
		return
	}

	alreadyStopped, err := s.manager.StopTunnel(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if alreadyStopped {
		writeJSON(w, http.StatusOK, map[string]string{"status": "already stopped"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func (s *Server) GetConnectionLogs(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/tunnels/"), "/logs")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tunnel ID")
		return
	}

	config, err := s.db.GetTunnelConfig(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check tunnel")
		return
	}
	if config == nil {
		writeError(w, http.StatusNotFound, "tunnel not found")
		return
	}

	logs, err := s.manager.GetConnectionLogs(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get logs")
		return
	}

	writeJSON(w, http.StatusOK, logs)
}

func (s *Server) ListRecords(w http.ResponseWriter, r *http.Request) {
	records, err := s.workflow.ListRecords()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list records")
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (s *Server) GetRecord(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/records/")
	if strings.Contains(idStr, "/history") {
		s.GetRecordHistory(w, r)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid record ID")
		return
	}

	record, err := s.workflow.GetRecord(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get record")
		return
	}
	if record == nil {
		writeError(w, http.StatusNotFound, "record not found")
		return
	}

	config, _ := s.db.GetTunnelConfig(record.TunnelConfigID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"record": record,
		"config": config,
	})
}

func (s *Server) GetRecordHistory(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/records/")
	parts := strings.Split(path, "/")
	if len(parts) < 1 {
		writeError(w, http.StatusBadRequest, "invalid record ID")
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid record ID")
		return
	}

	record, err := s.workflow.GetRecord(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check record")
		return
	}
	if record == nil {
		writeError(w, http.StatusNotFound, "record not found")
		return
	}

	histories, err := s.workflow.GetOperationHistories(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get history")
		return
	}

	writeJSON(w, http.StatusOK, histories)
}

type AddNoteRequest struct {
	Note     string `json:"note"`
	HistoryID int64  `json:"history_id"`
}

func (s *Server) AddHistoryNote(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/records/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 || parts[1] != "history" {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	recordID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid record ID")
		return
	}

	historyID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid history ID")
		return
	}

	record, err := s.workflow.GetRecord(recordID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check record")
		return
	}
	if record == nil {
		writeError(w, http.StatusNotFound, "record not found")
		return
	}

	var req AddNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Note == "" {
		writeError(w, http.StatusBadRequest, "note is required")
		return
	}

	histories, _ := s.workflow.GetOperationHistories(recordID)
	found := false
	for _, h := range histories {
		if h.ID == historyID {
			found = true
			break
		}
	}
	if !found {
		writeError(w, http.StatusNotFound, "history not found for this record")
		return
	}

	if err := s.workflow.AddNote(historyID, req.Note); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type WorkflowActionRequest struct {
	Operator string `json:"operator"`
	Reason   string `json:"reason"`
}

func (s *Server) SubmitRecord(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/records/"), "/submit")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid record ID")
		return
	}

	var req WorkflowActionRequest
	json.NewDecoder(r.Body).Decode(&req)
	operator := req.Operator
	if operator == "" {
		operator = "api"
	}

	if err := s.workflow.SubmitForReview(id, operator); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "submitted"})
}

func (s *Server) ApproveRecord(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/records/"), "/approve")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid record ID")
		return
	}

	var req WorkflowActionRequest
	json.NewDecoder(r.Body).Decode(&req)
	operator := req.Operator
	if operator == "" {
		operator = "api"
	}

	if err := s.workflow.Review(id, operator, true, req.Reason); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "approved"})
}

func (s *Server) RejectRecord(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/records/"), "/reject")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid record ID")
		return
	}

	var req WorkflowActionRequest
	json.NewDecoder(r.Body).Decode(&req)
	operator := req.Operator
	if operator == "" {
		operator = "api"
	}

	if req.Reason == "" {
		writeError(w, http.StatusBadRequest, "reject reason is required")
		return
	}

	if err := s.workflow.Review(id, operator, false, req.Reason); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}

func (s *Server) ExecuteRecord(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/records/"), "/execute")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid record ID")
		return
	}

	var req WorkflowActionRequest
	json.NewDecoder(r.Body).Decode(&req)
	operator := req.Operator
	if operator == "" {
		operator = "api"
	}

	if err := s.workflow.StartExecution(id, operator); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "executing"})
}

func (s *Server) CompleteRecord(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/records/"), "/complete")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid record ID")
		return
	}

	var req WorkflowActionRequest
	json.NewDecoder(r.Body).Decode(&req)
	operator := req.Operator
	if operator == "" {
		operator = "api"
	}

	if err := s.workflow.Complete(id, operator); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "completed"})
}

func (s *Server) Register(mux *http.ServeMux) {
	mux.HandleFunc("/tunnels", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.ListTunnels(w, r)
		case http.MethodPost:
			s.CreateTunnel(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	mux.HandleFunc("/tunnels/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/start") {
			if r.Method == http.MethodPost {
				s.StartTunnel(w, r)
			} else {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if strings.HasSuffix(path, "/stop") {
			if r.Method == http.MethodPost {
				s.StopTunnel(w, r)
			} else {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if strings.HasSuffix(path, "/logs") {
			if r.Method == http.MethodGet {
				s.GetConnectionLogs(w, r)
			} else {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if r.Method == http.MethodGet {
			s.GetTunnel(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	mux.HandleFunc("/records", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.ListRecords(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	mux.HandleFunc("/records/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/submit") {
			if r.Method == http.MethodPost {
				s.SubmitRecord(w, r)
			} else {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if strings.HasSuffix(path, "/approve") {
			if r.Method == http.MethodPost {
				s.ApproveRecord(w, r)
			} else {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if strings.HasSuffix(path, "/reject") {
			if r.Method == http.MethodPost {
				s.RejectRecord(w, r)
			} else {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if strings.HasSuffix(path, "/execute") {
			if r.Method == http.MethodPost {
				s.ExecuteRecord(w, r)
			} else {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if strings.HasSuffix(path, "/complete") {
			if r.Method == http.MethodPost {
				s.CompleteRecord(w, r)
			} else {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}

		parts := strings.Split(path, "/")
		if len(parts) >= 5 && parts[3] == "history" {
			if r.Method == http.MethodPost {
				s.AddHistoryNote(w, r)
			} else if r.Method == http.MethodGet {
				s.GetRecordHistory(w, r)
			} else {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if strings.Contains(path, "/history") {
			if r.Method == http.MethodGet {
				s.GetRecordHistory(w, r)
			} else {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}

		if r.Method == http.MethodGet {
			s.GetRecord(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})
}
