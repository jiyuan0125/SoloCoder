package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"insurance-claim/pkg/common"
	"insurance-claim/pkg/core"
)

type Handler struct {
	service core.ClaimService
}

func NewHandler(service core.ClaimService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/policies", h.handlePolicies)
	mux.HandleFunc("/api/claims", h.handleClaims)
	mux.HandleFunc("/api/claims/", h.handleClaimByID)

	return mux
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, common.ErrorResponse{
		Success: false,
		Error:   err.Error(),
	})
}

func (h *Handler) handlePolicies(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createPolicy(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) createPolicy(w http.ResponseWriter, r *http.Request) {
	var policy common.Policy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.service.CreatePolicy(&policy); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success":   true,
		"policy_no": policy.PolicyNo,
	})
}

func (h *Handler) handleClaims(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.submitClaim(w, r)
	case http.MethodGet:
		h.listClaims(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleClaimByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/claims/")
	parts := strings.Split(path, "/")
	caseID := parts[0]

	if caseID == "" {
		http.Error(w, "Case ID required", http.StatusBadRequest)
		return
	}

	if len(parts) > 1 {
		action := parts[1]
		h.handleClaimAction(w, r, caseID, action)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getClaim(w, r, caseID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleClaimAction(w http.ResponseWriter, r *http.Request, caseID, action string) {
	switch r.Method {
	case http.MethodPost:
	case http.MethodPut:
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	switch action {
	case "assign-investigator":
		h.assignInvestigator(w, r, caseID)
	case "submit-investigation":
		h.submitInvestigation(w, r, caseID)
	case "assign-assessor":
		h.assignAssessor(w, r, caseID)
	case "submit-assessment":
		h.submitAssessment(w, r, caseID)
	case "calculate-payout":
		h.calculatePayout(w, r, caseID)
	case "approve-payout":
		h.approvePayout(w, r, caseID)
	case "mark-paid":
		h.markAsPaid(w, r, caseID)
	case "flag-review":
		h.flagForReview(w, r, caseID)
	case "resolve-review":
		h.resolveReview(w, r, caseID)
	default:
		http.Error(w, "Unknown action", http.StatusNotFound)
	}
}

func (h *Handler) submitClaim(w http.ResponseWriter, r *http.Request) {
	var req common.SubmitClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	log.Printf("Received claim submission for policy: %s", req.PolicyNo)

	caseItem, err := h.service.SubmitClaim(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusCreated, common.SubmitClaimResponse{
		Success: true,
		CaseID:  caseItem.ID,
		CaseNo:  caseItem.CaseNo,
	})
}

func (h *Handler) listClaims(w http.ResponseWriter, r *http.Request) {
	cases := h.service.ListCases()
	writeJSON(w, http.StatusOK, common.ListCasesResponse{
		Success: true,
		Cases:   cases,
	})
}

func (h *Handler) getClaim(w http.ResponseWriter, r *http.Request, caseID string) {
	c, ok := h.service.GetCase(caseID)
	if !ok {
		c, ok = h.service.GetCaseByNo(caseID)
	}

	if !ok {
		writeError(w, http.StatusNotFound, nil)
		return
	}

	writeJSON(w, http.StatusOK, common.GetCaseResponse{
		Success: true,
		Case:    c,
	})
}

func (h *Handler) assignInvestigator(w http.ResponseWriter, r *http.Request, caseID string) {
	var req common.AssignInvestigatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	req.CaseID = caseID

	if err := h.service.AssignInvestigator(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, common.UpdateCaseResponse{Success: true})
}

func (h *Handler) submitInvestigation(w http.ResponseWriter, r *http.Request, caseID string) {
	var req common.SubmitInvestigationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	req.CaseID = caseID

	if err := h.service.SubmitInvestigation(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, common.UpdateCaseResponse{Success: true})
}

func (h *Handler) assignAssessor(w http.ResponseWriter, r *http.Request, caseID string) {
	var req common.AssignAssessorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	req.CaseID = caseID

	if err := h.service.AssignAssessor(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, common.UpdateCaseResponse{Success: true})
}

func (h *Handler) submitAssessment(w http.ResponseWriter, r *http.Request, caseID string) {
	var req common.SubmitAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	req.CaseID = caseID

	if err := h.service.SubmitAssessment(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, common.UpdateCaseResponse{Success: true})
}

func (h *Handler) calculatePayout(w http.ResponseWriter, r *http.Request, caseID string) {
	payout, err := h.service.CalculatePayout(caseID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, common.CalculatePayoutResponse{
		Success: true,
		Payout:  payout,
	})
}

func (h *Handler) approvePayout(w http.ResponseWriter, r *http.Request, caseID string) {
	var body struct {
		Operator string `json:"operator"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.service.ApprovePayout(caseID, body.Operator); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, common.UpdateCaseResponse{Success: true})
}

func (h *Handler) markAsPaid(w http.ResponseWriter, r *http.Request, caseID string) {
	var body struct {
		Operator string `json:"operator"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.service.MarkAsPaid(caseID, body.Operator); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, common.UpdateCaseResponse{Success: true})
}

func (h *Handler) flagForReview(w http.ResponseWriter, r *http.Request, caseID string) {
	var req common.FlagForReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	req.CaseID = caseID

	if err := h.service.FlagForReview(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, common.UpdateCaseResponse{Success: true})
}

func (h *Handler) resolveReview(w http.ResponseWriter, r *http.Request, caseID string) {
	var req common.ResolveReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	req.CaseID = caseID

	if err := h.service.ResolveReview(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, common.UpdateCaseResponse{Success: true})
}
