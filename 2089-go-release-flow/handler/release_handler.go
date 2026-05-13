package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"release-flow/dao"
	"release-flow/model"
	"release-flow/service"
)

type ReleaseHandler struct {
	service *service.ReleaseService
}

func NewReleaseHandler(dao *dao.ReleaseDAO) *ReleaseHandler {
	return &ReleaseHandler{
		service: service.NewReleaseService(dao),
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
	CurrentStatus model.Status `json:"current_status,omitempty"`
}

type CreateReleaseRequest struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type StatusTransitionRequest struct {
	Operator string `json:"operator"`
	Reason   string `json:"reason"`
}

type CanaryRequest struct {
	Operator   string `json:"operator"`
	Reason     string `json:"reason"`
	CanaryRatio int   `json:"canary_ratio"`
}

func (h *ReleaseHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *ReleaseHandler) respondError(w http.ResponseWriter, status int, message string, currentStatus ...model.Status) {
	resp := ErrorResponse{Error: message}
	if len(currentStatus) > 0 {
		resp.CurrentStatus = currentStatus[0]
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func (h *ReleaseHandler) parseReleaseID(path string) (int64, error) {
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		return 0, errors.New("invalid path")
	}
	return strconv.ParseInt(parts[3], 10, 64)
}

func (h *ReleaseHandler) getSubResource(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 5 {
		return parts[4]
	}
	return ""
}

func (h *ReleaseHandler) ListReleases(w http.ResponseWriter, r *http.Request) {
	releases, err := h.service.GetAllReleases()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, releases)
}

func (h *ReleaseHandler) CreateRelease(w http.ResponseWriter, r *http.Request) {
	var req CreateReleaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	if req.Name == "" || req.Version == "" {
		h.respondError(w, http.StatusBadRequest, "name 和 version 不能为空")
		return
	}

	release, err := h.service.CreateRelease(req.Name, req.Version)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, release)
}

func (h *ReleaseHandler) GetRelease(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseReleaseID(r.URL.Path)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "无效的发布单 ID")
		return
	}

	release, err := h.service.GetReleaseByID(id)
	if err != nil {
		if err == dao.ErrReleaseNotFound {
			h.respondError(w, http.StatusNotFound, "发布单不存在")
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, release)
}

func (h *ReleaseHandler) HandleReleaseSubResource(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseReleaseID(r.URL.Path)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "无效的发布单 ID")
		return
	}

	subResource := h.getSubResource(r.URL.Path)
	switch subResource {
	case "status-logs":
		h.ListStatusLogs(w, r, id)
	case "rollback-records":
		h.ListRollbackRecords(w, r, id)
	case "commit":
		h.TryCommit(w, r, id)
	default:
		h.respondError(w, http.StatusNotFound, "子资源不存在")
	}
}

func (h *ReleaseHandler) ListStatusLogs(w http.ResponseWriter, r *http.Request, id int64) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	_, err := h.service.GetReleaseByID(id)
	if err != nil {
		if err == dao.ErrReleaseNotFound {
			h.respondError(w, http.StatusNotFound, "发布单不存在")
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	logs, err := h.service.GetStatusLogs(id)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, logs)
}

func (h *ReleaseHandler) ListRollbackRecords(w http.ResponseWriter, r *http.Request, id int64) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	_, err := h.service.GetReleaseByID(id)
	if err != nil {
		if err == dao.ErrReleaseNotFound {
			h.respondError(w, http.StatusNotFound, "发布单不存在")
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	records, err := h.service.GetRollbackRecords(id)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, records)
}

func (h *ReleaseHandler) TryCommit(w http.ResponseWriter, r *http.Request, id int64) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	release, err := h.service.GetReleaseByID(id)
	if err != nil {
		if err == dao.ErrReleaseNotFound {
			h.respondError(w, http.StatusNotFound, "发布单不存在")
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.service.TryCommit(id); err != nil {
		h.respondError(w, http.StatusConflict, err.Error(), release.Status)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "允许提交"})
}

func (h *ReleaseHandler) HandleStatusAction(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseReleaseID(r.URL.Path)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "无效的发布单 ID")
		return
	}

	release, err := h.service.GetReleaseByID(id)
	if err != nil {
		if err == dao.ErrReleaseNotFound {
			h.respondError(w, http.StatusNotFound, "发布单不存在")
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 6 {
		h.respondError(w, http.StatusBadRequest, "无效的操作")
		return
	}

	action := parts[5]
	switch action {
	case "submit-testing":
		h.submitForTesting(w, r, release)
	case "reject-testing":
		h.rejectTesting(w, r, release)
	case "pass-testing":
		h.passTesting(w, r, release)
	case "start-canary":
		h.startCanary(w, r, release)
	case "full-release":
		h.fullRelease(w, r, release)
	case "complete":
		h.completeRelease(w, r, release)
	case "rollback":
		h.rollback(w, r, release)
	default:
		h.respondError(w, http.StatusBadRequest, fmt.Sprintf("无效的操作: %s", action), release.Status)
	}
}

func (h *ReleaseHandler) submitForTesting(w http.ResponseWriter, r *http.Request, release *model.Release) {
	var req StatusTransitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "无效的请求体", release.Status)
		return
	}

	if req.Operator == "" {
		h.respondError(w, http.StatusBadRequest, "operator 不能为空", release.Status)
		return
	}

	updated, err := h.service.SubmitForTesting(release.ID, req.Operator, req.Reason)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error(), release.Status)
		return
	}

	h.respondJSON(w, http.StatusOK, updated)
}

func (h *ReleaseHandler) rejectTesting(w http.ResponseWriter, r *http.Request, release *model.Release) {
	var req StatusTransitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "无效的请求体", release.Status)
		return
	}

	if req.Operator == "" {
		h.respondError(w, http.StatusBadRequest, "operator 不能为空", release.Status)
		return
	}

	updated, err := h.service.RejectTesting(release.ID, req.Operator, req.Reason)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error(), release.Status)
		return
	}

	h.respondJSON(w, http.StatusOK, updated)
}

func (h *ReleaseHandler) passTesting(w http.ResponseWriter, r *http.Request, release *model.Release) {
	var req StatusTransitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "无效的请求体", release.Status)
		return
	}

	if req.Operator == "" {
		h.respondError(w, http.StatusBadRequest, "operator 不能为空", release.Status)
		return
	}

	updated, err := h.service.PassTesting(release.ID, req.Operator, req.Reason)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error(), release.Status)
		return
	}

	h.respondJSON(w, http.StatusOK, updated)
}

func (h *ReleaseHandler) startCanary(w http.ResponseWriter, r *http.Request, release *model.Release) {
	var req CanaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "无效的请求体", release.Status)
		return
	}

	if req.Operator == "" {
		h.respondError(w, http.StatusBadRequest, "operator 不能为空", release.Status)
		return
	}

	updated, err := h.service.StartCanary(release.ID, req.CanaryRatio, req.Operator, req.Reason)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error(), release.Status)
		return
	}

	h.respondJSON(w, http.StatusOK, updated)
}

func (h *ReleaseHandler) fullRelease(w http.ResponseWriter, r *http.Request, release *model.Release) {
	var req StatusTransitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "无效的请求体", release.Status)
		return
	}

	if req.Operator == "" {
		h.respondError(w, http.StatusBadRequest, "operator 不能为空", release.Status)
		return
	}

	updated, err := h.service.FullRelease(release.ID, req.Operator, req.Reason)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error(), release.Status)
		return
	}

	h.respondJSON(w, http.StatusOK, updated)
}

func (h *ReleaseHandler) completeRelease(w http.ResponseWriter, r *http.Request, release *model.Release) {
	var req StatusTransitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "无效的请求体", release.Status)
		return
	}

	if req.Operator == "" {
		h.respondError(w, http.StatusBadRequest, "operator 不能为空", release.Status)
		return
	}

	updated, err := h.service.CompleteRelease(release.ID, req.Operator, req.Reason)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error(), release.Status)
		return
	}

	h.respondJSON(w, http.StatusOK, updated)
}

func (h *ReleaseHandler) rollback(w http.ResponseWriter, r *http.Request, release *model.Release) {
	var req StatusTransitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "无效的请求体", release.Status)
		return
	}

	if req.Operator == "" {
		h.respondError(w, http.StatusBadRequest, "operator 不能为空", release.Status)
		return
	}

	record, err := h.service.CanaryRollback(release.ID, req.Operator, req.Reason)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error(), release.Status)
		return
	}

	h.respondJSON(w, http.StatusOK, record)
}
