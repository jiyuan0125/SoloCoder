package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"jobposting/common"
	"jobposting/server/store"
)

type Handler struct {
	store *store.DataStore
}

func NewHandler(s *store.DataStore) *Handler {
	return &Handler{store: s}
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, common.ErrorResponse{Error: message})
}

func (h *Handler) CreateJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req common.CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	if err := common.ValidateCreateJobRequest(req); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	job, err := h.store.CreateJob(req)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "创建职位失败")
		return
	}

	h.respondJSON(w, http.StatusOK, common.CreateJobResponse{JobID: job.ID})
}

func (h *Handler) ListJobsForJobseeker(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var filter common.ListJobsRequest

	minSalaryStr := r.URL.Query().Get("min_salary")
	if minSalaryStr != "" {
		if minSalary, err := strconv.Atoi(minSalaryStr); err == nil {
			filter.MinSalary = &minSalary
		}
	}

	city := r.URL.Query().Get("city")
	if city != "" {
		filter.City = &city
	}

	education := r.URL.Query().Get("education")
	if education != "" {
		edu := common.EducationLevel(education)
		filter.Education = &edu
	}

	jobs := h.store.ListJobsForJobseeker(filter)
	h.respondJSON(w, http.StatusOK, common.ListJobsResponse{Jobs: jobs})
}

func (h *Handler) ListJobsForCompany(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	jobs := h.store.ListJobsForCompany()
	h.respondJSON(w, http.StatusOK, common.ListJobsResponse{Jobs: jobs})
}

func (h *Handler) OfflineJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		h.respondError(w, http.StatusBadRequest, "缺少职位ID")
		return
	}
	jobID := parts[len(parts)-1]

	if jobID == "" {
		h.respondError(w, http.StatusBadRequest, "职位ID不能为空")
		return
	}

	_, ok := h.store.GetJob(jobID)
	if !ok {
		h.respondError(w, http.StatusNotFound, "职位不存在")
		return
	}

	if err := h.store.OfflineJob(jobID); err != nil {
		h.respondError(w, http.StatusInternalServerError, "下架职位失败")
		return
	}

	h.respondJSON(w, http.StatusOK, common.OfflineJobResponse{Success: true})
}

func (h *Handler) ApplyJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		h.respondError(w, http.StatusBadRequest, "缺少职位ID")
		return
	}
	jobID := parts[len(parts)-1]

	if jobID == "" {
		h.respondError(w, http.StatusBadRequest, "职位ID不能为空")
		return
	}

	job, ok := h.store.GetJob(jobID)
	if !ok {
		h.respondError(w, http.StatusNotFound, "职位不存在")
		return
	}
	if job.IsOffline {
		h.respondError(w, http.StatusBadRequest, "职位已下架，无法投递")
		return
	}

	var req common.ApplyJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	if err := common.ValidateApplyJobRequest(req); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	app, err := h.store.ApplyJob(jobID, req)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "投递失败")
		return
	}
	if app == nil {
		h.respondError(w, http.StatusBadRequest, "您已投递过该职位")
		return
	}

	h.respondJSON(w, http.StatusOK, common.ApplyJobResponse{ApplicationID: app.ID})
}

func (h *Handler) ListApplications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var jobID *string
	jobIDStr := r.URL.Query().Get("job_id")
	if jobIDStr != "" {
		jobID = &jobIDStr
	}

	apps := h.store.ListApplications(jobID)
	h.respondJSON(w, http.StatusOK, common.ListApplicationsResponse{Applications: apps})
}

func (h *Handler) ListMyApplications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	phone := r.URL.Query().Get("phone")
	if phone == "" {
		h.respondError(w, http.StatusBadRequest, "缺少手机号参数")
		return
	}

	apps := h.store.ListApplicationsByPhone(phone)
	h.respondJSON(w, http.StatusOK, common.ListMyApplicationsResponse{Applications: apps})
}

func (h *Handler) UpdateApplicationStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		h.respondError(w, http.StatusBadRequest, "缺少投递记录ID")
		return
	}
	appID := parts[len(parts)-1]

	if appID == "" {
		h.respondError(w, http.StatusBadRequest, "投递记录ID不能为空")
		return
	}

	_, ok := h.store.GetApplication(appID)
	if !ok {
		h.respondError(w, http.StatusNotFound, "投递记录不存在")
		return
	}

	var req common.UpdateApplicationStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	if err := common.ValidateUpdateApplicationStatusRequest(req); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.store.UpdateApplicationStatus(appID, req); err != nil {
		h.respondError(w, http.StatusInternalServerError, "更新状态失败")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]bool{"success": true})
}
