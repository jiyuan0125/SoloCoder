package main

import (
	"encoding/json"
	"io"
	"net/http"
	"school-enrollment/common"
	"strings"
)

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, common.APIResponse{
		Success: false,
		Message: message,
		Data:    nil,
	})
}

func (h *Handler) respondSuccess(w http.ResponseWriter, message string, data interface{}) {
	h.respondJSON(w, http.StatusOK, common.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func (h *Handler) parseBody(r *http.Request, dest interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.Unmarshal(body, dest)
}

func (h *Handler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	var req common.CreatePlanRequest
	if err := h.parseBody(r, &req); err != nil {
		h.respondError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if !common.ValidatePlanName(req.Name) {
		h.respondError(w, http.StatusBadRequest, "招生计划名称不能为空且不能超过50个字")
		return
	}

	if strings.TrimSpace(req.Grade) == "" {
		h.respondError(w, http.StatusBadRequest, "年级不能为空")
		return
	}

	if req.Classes <= 0 {
		h.respondError(w, http.StatusBadRequest, "班级数量必须大于0")
		return
	}

	if req.MaxPerClass <= 0 {
		h.respondError(w, http.StatusBadRequest, "每班人数必须大于0")
		return
	}

	plan := &common.EnrollmentPlan{
		ID:          common.GenerateID(),
		Name:        strings.TrimSpace(req.Name),
		Grade:       strings.TrimSpace(req.Grade),
		Classes:     req.Classes,
		MaxPerClass: req.MaxPerClass,
	}

	h.store.CreatePlan(plan)
	h.respondSuccess(w, "创建成功", plan)
}

func (h *Handler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans := h.store.ListPlans()
	h.respondSuccess(w, "", plans)
}

func (h *Handler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	var req common.UpdatePlanRequest
	if err := h.parseBody(r, &req); err != nil {
		h.respondError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	planID := r.URL.Query().Get("id")
	if planID == "" {
		h.respondError(w, http.StatusBadRequest, "缺少计划ID")
		return
	}

	plan := h.store.GetPlan(planID)
	if plan == nil {
		h.respondError(w, http.StatusNotFound, "招生计划不存在")
		return
	}

	if req.Classes <= 0 {
		req.Classes = plan.Classes
	}
	if req.MaxPerClass <= 0 {
		req.MaxPerClass = plan.MaxPerClass
	}

	newCapacity := req.Classes * req.MaxPerClass
	if newCapacity < plan.EnrolledCount {
		h.respondError(w, http.StatusBadRequest, "总容量不能小于已录取人数")
		return
	}

	h.store.UpdatePlanCapacity(planID, req.Classes, req.MaxPerClass)
	updatedPlan := h.store.GetPlan(planID)
	h.respondSuccess(w, "更新成功", updatedPlan)
}

func (h *Handler) ClosePlan(w http.ResponseWriter, r *http.Request) {
	var req common.ClosePlanRequest
	if err := h.parseBody(r, &req); err != nil {
		planID := r.URL.Query().Get("id")
		if planID == "" {
			h.respondError(w, http.StatusBadRequest, "缺少计划ID")
			return
		}
		req.PlanID = planID
	}

	plan := h.store.GetPlan(req.PlanID)
	if plan == nil {
		h.respondError(w, http.StatusNotFound, "招生计划不存在")
		return
	}

	h.store.ClosePlan(req.PlanID)
	updatedPlan := h.store.GetPlan(req.PlanID)
	h.respondSuccess(w, "已关闭", updatedPlan)
}

func (h *Handler) SubmitRegistration(w http.ResponseWriter, r *http.Request) {
	var req common.SubmitRegistrationRequest
	if err := h.parseBody(r, &req); err != nil {
		h.respondError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if strings.TrimSpace(req.PlanID) == "" {
		h.respondError(w, http.StatusBadRequest, "请选择招生计划")
		return
	}

	if !common.ValidateStudentName(req.StudentName) {
		h.respondError(w, http.StatusBadRequest, "学生姓名不能为空且不能超过20个字")
		return
	}

	if !common.ValidateIDCard(req.IDCard) {
		h.respondError(w, http.StatusBadRequest, "身份证号格式错误，必须是18位，末位可以是数字或字母X")
		return
	}

	if !common.ValidatePhone(req.ParentPhone) {
		h.respondError(w, http.StatusBadRequest, "联系电话格式错误，必须是11位手机号且以1开头")
		return
	}

	if !common.ValidateAddress(req.Address) {
		h.respondError(w, http.StatusBadRequest, "户籍地址不能为空")
		return
	}

	plan := h.store.GetPlan(req.PlanID)
	if plan == nil {
		h.respondError(w, http.StatusNotFound, "招生计划不存在")
		return
	}

	if !plan.IsOpen {
		h.respondError(w, http.StatusBadRequest, "该招生计划已关闭，不接受新的报名")
		return
	}

	if plan.EnrolledCount >= plan.TotalCapacity {
		h.respondError(w, http.StatusBadRequest, "报名人数已达上限，报名失败")
		return
	}

	existing := h.store.GetRegistrationByIDCard(req.PlanID, req.IDCard)
	if existing != nil {
		h.respondError(w, http.StatusBadRequest, "该身份证号已在此计划中报名，不能重复报名")
		return
	}

	reg := &common.Registration{
		ID:          common.GenerateID(),
		PlanID:      strings.TrimSpace(req.PlanID),
		StudentName: strings.TrimSpace(req.StudentName),
		IDCard:      strings.ToUpper(strings.TrimSpace(req.IDCard)),
		ParentPhone: strings.TrimSpace(req.ParentPhone),
		Address:     strings.TrimSpace(req.Address),
	}

	h.store.CreateRegistration(reg)
	h.respondSuccess(w, "报名成功", reg)
}

func (h *Handler) QueryRegistration(w http.ResponseWriter, r *http.Request) {
	var req common.QueryRegistrationByIDCardRequest
	if err := h.parseBody(r, &req); err != nil {
		req.PlanID = r.URL.Query().Get("plan_id")
		req.IDCard = r.URL.Query().Get("id_card")
	}

	if req.PlanID == "" || req.IDCard == "" {
		h.respondError(w, http.StatusBadRequest, "请提供计划ID和身份证号")
		return
	}

	reg := h.store.GetRegistrationByIDCard(req.PlanID, req.IDCard)
	if reg == nil {
		h.respondError(w, http.StatusNotFound, "未找到该报名记录")
		return
	}

	h.respondSuccess(w, "", reg)
}

func (h *Handler) ListRegistrations(w http.ResponseWriter, r *http.Request) {
	var req common.QueryRegistrationsRequest
	if err := h.parseBody(r, &req); err != nil {
		req.PlanID = r.URL.Query().Get("plan_id")
		req.Status = r.URL.Query().Get("status")
	}

	if req.PlanID == "" {
		h.respondError(w, http.StatusBadRequest, "请提供计划ID")
		return
	}

	regs := h.store.ListRegistrations(req.PlanID, req.Status)
	h.respondSuccess(w, "", regs)
}

func (h *Handler) ReviewRegistration(w http.ResponseWriter, r *http.Request) {
	var req common.ReviewRegistrationRequest
	if err := h.parseBody(r, &req); err != nil {
		h.respondError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.RegistrationID == "" {
		h.respondError(w, http.StatusBadRequest, "请提供报名记录ID")
		return
	}

	reg := h.store.GetRegistration(req.RegistrationID)
	if reg == nil {
		h.respondError(w, http.StatusNotFound, "报名记录不存在")
		return
	}

	if reg.Status != common.StatusPending {
		h.respondError(w, http.StatusBadRequest, "只有待审核状态的报名可以被操作")
		return
	}

	switch req.Action {
	case "accept":
		h.store.AcceptRegistration(req.RegistrationID)
		h.respondSuccess(w, "已录取", h.store.GetRegistration(req.RegistrationID))
	case "reject":
		if strings.TrimSpace(req.RejectReason) == "" {
			h.respondError(w, http.StatusBadRequest, "拒绝时需要填写拒绝原因")
			return
		}
		h.store.RejectRegistration(req.RegistrationID, strings.TrimSpace(req.RejectReason))
		h.respondSuccess(w, "已拒绝", h.store.GetRegistration(req.RegistrationID))
	default:
		h.respondError(w, http.StatusBadRequest, "无效的操作类型，仅支持 accept 或 reject")
	}
}
