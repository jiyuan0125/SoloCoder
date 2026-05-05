package handler

import (
	"audit-log/common"
	"audit-log/server/model"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type AuditHandler struct {
	storage          *model.InMemoryStorage
	securityMgr      *model.SecurityManager
	archiveMgr       *model.ArchiveManager
	statisticsMgr    *model.StatisticsManager
	exportApprovalMgr *model.ExportApprovalManager
}

func NewAuditHandler(storage *model.InMemoryStorage) *AuditHandler {
	return &AuditHandler{
		storage:          storage,
		securityMgr:      model.NewSecurityManager(storage),
		archiveMgr:       model.NewArchiveManager(storage),
		statisticsMgr:    model.NewStatisticsManager(storage),
		exportApprovalMgr: model.NewExportApprovalManager(storage),
	}
}

func (h *AuditHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *AuditHandler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{
		"error": message,
	})
}

func (h *AuditHandler) CreateLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.CreateLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if common.RequireApprovalOperations[req.OperationType] && req.Approver == "" {
		if !h.exportApprovalMgr.IsExportApproved(req.UserID) {
			h.writeError(w, http.StatusForbidden, "export operation requires approval")
			return
		}
	}

	if req.OperationType == common.OperationTypeDelete {
		if req.Snapshot == nil || len(req.Snapshot) == 0 {
			h.writeError(w, http.StatusBadRequest, "delete operation requires snapshot data")
			return
		}
	}

	if req.OperationType == common.OperationTypeExport {
		if req.ExportRange == nil {
			h.writeError(w, http.StatusBadRequest, "export operation requires export_range data")
			return
		}
	}

	log, err := h.storage.CreateLog(&req)
	if err != nil {
		switch err {
		case common.ErrMissingUserID, common.ErrMissingOperationType:
			h.writeError(w, http.StatusBadRequest, err.Error())
		case common.ErrStorageCapacity:
			h.writeError(w, http.StatusServiceUnavailable, err.Error())
		default:
			h.writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	h.storage.RecordUserOperation(req.UserID)

	if isAbnormal, reason := h.securityMgr.CheckAbnormalBehavior(req.UserID); isAbnormal {
		h.storage.MarkLogAsAbnormal(log.ID, reason)
		fmt.Printf("[SECURITY ALERT] Abnormal behavior detected: user=%s, reason=%s\n", req.UserID, reason)
	}

	h.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":      log.ID,
		"success": true,
	})
}

func (h *AuditHandler) GetLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/logs/")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "log id is required")
		return
	}

	log, err := h.storage.GetLogByID(id)
	if err != nil {
		if err == common.ErrLogNotFound {
			h.writeError(w, http.StatusNotFound, "log not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, log)
}

func (h *AuditHandler) QueryLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.QueryLogsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.storage.QueryLogs(&req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

func (h *AuditHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.securityMgr.RecordLoginAttempt(req.UserID, req.UserName, req.IPAddress, req.Success)

	logReq := &common.CreateLogRequest{
		UserID:        req.UserID,
		UserName:      req.UserName,
		IPAddress:     req.IPAddress,
		OperationType: common.OperationTypeLogin,
		Description:   "User login attempt",
		Result:        common.ResultSuccess,
	}

	if !req.Success {
		logReq.Result = common.ResultFailed
	}

	if err == common.ErrUserLocked {
		logReq.Description = "User login attempt - account locked"
		logReq.Result = common.ResultFailed
	}

	h.storage.CreateLog(logReq)

	if err != nil {
		if err == common.ErrUserLocked {
			h.writeJSON(w, http.StatusLocked, result)
			return
		}
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

func (h *AuditHandler) GetUserLockStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	result := h.securityMgr.GetUserLockInfo(userID)
	h.writeJSON(w, http.StatusOK, result)
}

func (h *AuditHandler) RequestExportApproval(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.ExportApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.exportApprovalMgr.RequestApproval(&req)
	if err != nil {
		if err == common.ErrNeedApproval {
			h.writeJSON(w, http.StatusOK, result)
			return
		}
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

func (h *AuditHandler) TriggerArchive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	count := h.archiveMgr.ArchiveOldLogs()

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"archived_count": count,
		"success":        true,
	})
}

func (h *AuditHandler) GetArchivedList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	archives := h.archiveMgr.GetArchivedMonths()
	h.writeJSON(w, http.StatusOK, archives)
}

func (h *AuditHandler) GetStorageStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	status := h.storage.GetStorageStatus()

	if status.NeedAlert {
		fmt.Printf("[STORAGE ALERT] Storage usage exceeded 80%%: %.2f%%\n", status.UsagePercent*100)
	}

	h.writeJSON(w, http.StatusOK, status)
}

func (h *AuditHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.StatisticsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.statisticsMgr.GetStatistics(&req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

func (h *AuditHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
