package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"contract-management/common"
	"contract-management/core"
)

type Handler struct {
	contractSvc  *core.ContractService
	milestoneSvc *core.MilestoneService
	paymentSvc   *core.PaymentService
	auditSvc     *core.AuditService
}

func NewHandler(
	contractSvc *core.ContractService,
	milestoneSvc *core.MilestoneService,
	paymentSvc *core.PaymentService,
	auditSvc *core.AuditService,
) *Handler {
	return &Handler{
		contractSvc:  contractSvc,
		milestoneSvc: milestoneSvc,
		paymentSvc:   paymentSvc,
		auditSvc:     auditSvc,
	}
}

func (h *Handler) CreateContract(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.CreateContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := core.CreateContractInput{
		ContractNo:   req.ContractNo,
		PartyA:       req.PartyA,
		PartyB:       req.PartyB,
		TotalAmount:  req.TotalAmount,
		SignDate:     req.SignDate,
		ContractType: req.ContractType,
		Operator:     req.Operator,
	}

	if req.ServiceWindow != nil {
		input.ServiceWindow = &core.ServiceWindow{
			StartTime: req.ServiceWindow.StartTime,
			EndTime:   req.ServiceWindow.EndTime,
			Weekdays:  req.ServiceWindow.Weekdays,
		}
	}

	for _, m := range req.Milestones {
		input.Milestones = append(input.Milestones, core.CreateMilestoneInput{
			Name:              m.Name,
			PlannedDate:       m.PlannedDate,
			PlannedPercentage: m.PlannedPercentage,
			Owner:             m.Owner,
		})
	}

	contract, err := h.contractSvc.CreateContract(input)
	if err != nil {
		if errors.Is(err, core.ErrMilestonePercentageSum) || errors.Is(err, core.ErrInvalidPercentage) || errors.Is(err, core.ErrInvalidAmount) {
			writeError(w, http.StatusBadRequest, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	resp := contractToResponse(contract)
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) GetContract(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id := extractID(r.URL.Path, "/contracts/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing contract id")
		return
	}

	contract, err := h.contractSvc.GetContract(id)
	if err != nil {
		if errors.Is(err, core.ErrContractNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	resp := contractToResponse(contract)
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) ListContracts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	contracts := h.contractSvc.ListContracts()
	resp := common.ListContractsResponse{
		Contracts: make([]common.ContractResponse, 0, len(contracts)),
	}
	for _, c := range contracts {
		resp.Contracts = append(resp.Contracts, contractToResponse(c))
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) UpdateContractAmount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id := extractID(r.URL.Path, "/contracts/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing contract id")
		return
	}

	var req common.UpdateContractAmountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	contract, err := h.contractSvc.UpdateContractAmount(core.UpdateContractAmountInput{
		ContractID:     id,
		NewTotalAmount: req.NewTotalAmount,
		Operator:       req.Operator,
	})
	if err != nil {
		if errors.Is(err, core.ErrContractNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
		} else if errors.Is(err, core.ErrContractAlreadyCompleted) || errors.Is(err, core.ErrInvalidAmount) {
			writeError(w, http.StatusBadRequest, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	resp := contractToResponse(contract)
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) CompleteMilestone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 || parts[0] != "contracts" || parts[2] != "milestones" || parts[3] == "" {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	contractID := parts[1]
	milestoneID := parts[3]

	var req common.CompleteMilestoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	milestone, err := h.milestoneSvc.CompleteMilestone(core.CompleteMilestoneInput{
		ContractID:  contractID,
		MilestoneID: milestoneID,
		ActualDate:  req.ActualDate,
		Approved:    req.Approved,
		Operator:    req.Operator,
	})
	if err != nil {
		if errors.Is(err, core.ErrContractNotFound) || errors.Is(err, core.ErrMilestoneNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
		} else if errors.Is(err, core.ErrMilestoneAlreadyCompleted) || errors.Is(err, core.ErrContractAlreadyCompleted) {
			writeError(w, http.StatusBadRequest, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	resp := milestoneToResponse(milestone)
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) PayMilestone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 6 || parts[0] != "contracts" || parts[2] != "milestones" || parts[4] != "pay" {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	contractID := parts[1]
	milestoneID := parts[3]

	var req common.PayMilestoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	payment, err := h.paymentSvc.PayMilestone(core.PayMilestoneInput{
		ContractID:  contractID,
		MilestoneID: milestoneID,
		PaidAmount:  req.PaidAmount,
		Operator:    req.Operator,
	})
	if err != nil {
		if errors.Is(err, core.ErrContractNotFound) || errors.Is(err, core.ErrMilestoneNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
		} else if errors.Is(err, core.ErrMilestoneNotApproved) || errors.Is(err, core.ErrPaymentAlreadyPaid) || errors.Is(err, core.ErrInvalidAmount) {
			writeError(w, http.StatusBadRequest, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	resp := paymentToResponse(payment)
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 || parts[0] != "contracts" || parts[2] != "progress" {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	contractID := parts[1]

	details, err := h.milestoneSvc.GetProgressDetails(contractID)
	if err != nil {
		if errors.Is(err, core.ErrContractNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, details)
}

func (h *Handler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 || parts[0] != "contracts" || parts[2] != "audit-logs" {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	contractID := parts[1]

	logs := h.auditSvc.GetAuditLogs(contractID)
	resp := common.ListAuditLogsResponse{
		Logs: make([]common.AuditLogEntry, 0, len(logs)),
	}
	for _, l := range logs {
		resp.Logs = append(resp.Logs, common.AuditLogEntry{
			ID:         l.ID,
			ContractID: l.ContractID,
			Operation:  l.Operation,
			Operator:   l.Operator,
			Timestamp:  l.Timestamp,
			Field:      l.Field,
			OldValue:   l.OldValue,
			NewValue:   l.NewValue,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func extractID(path, prefix string) string {
	if len(path) <= len(prefix) {
		return ""
	}
	remainder := path[len(prefix):]
	idx := strings.Index(remainder, "/")
	if idx == -1 {
		return remainder
	}
	if idx == 0 {
		return ""
	}
	return remainder[:idx]
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(common.ErrorResponse{Error: message})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func contractToResponse(c *core.Contract) common.ContractResponse {
	resp := common.ContractResponse{
		ID:           c.ID,
		ContractNo:   c.ContractNo,
		PartyA:       c.PartyA,
		PartyB:       c.PartyB,
		TotalAmount:  c.TotalAmount,
		SignDate:     c.SignDate,
		ContractType: c.ContractType,
		Status:       string(c.Status),
		Milestones:   make([]common.MilestoneResp, 0, len(c.Milestones)),
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}

	if c.ServiceWindow != nil {
		resp.ServiceWindow = &common.ServiceWindow{
			StartTime: c.ServiceWindow.StartTime,
			EndTime:   c.ServiceWindow.EndTime,
			Weekdays:  c.ServiceWindow.Weekdays,
		}
	}

	for _, m := range c.Milestones {
		resp.Milestones = append(resp.Milestones, milestoneToResponse(m))
	}

	return resp
}

func milestoneToResponse(m *core.Milestone) common.MilestoneResp {
	resp := common.MilestoneResp{
		ID:                m.ID,
		Name:              m.Name,
		PlannedDate:       m.PlannedDate,
		PlannedPercentage: m.PlannedPercentage,
		PlannedAmount:     m.PlannedAmount,
		Owner:             m.Owner,
		Status:            string(m.Status),
		IsOffHours:        m.IsOffHours,
	}

	if m.ActualDate != nil {
		actual := *m.ActualDate
		resp.ActualDate = &actual
	}

	if m.Payment != nil {
		p := paymentToResponse(m.Payment)
		resp.Payment = &p
	}

	return resp
}

func paymentToResponse(p *core.Payment) common.PaymentResp {
	resp := common.PaymentResp{
		ID:        p.ID,
		DueAmount: p.DueAmount,
		Status:    string(p.Status),
	}

	if p.PaidAmount != nil {
		paid := *p.PaidAmount
		resp.PaidAmount = &paid
	}

	if p.PaidDate != nil {
		paidDate := *p.PaidDate
		resp.PaidDate = &paidDate
	}

	return resp
}
