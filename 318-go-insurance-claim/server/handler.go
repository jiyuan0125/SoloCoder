package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"insurance-claim/common"
)

type Handler struct {
	service *Service
	mux     *http.ServeMux
}

func NewHandler(service *Service) *Handler {
	h := &Handler{
		service: service,
		mux:     http.NewServeMux(),
	}
	h.SetupRoutes(h.mux)
	return h
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
	})
}

func (h *Handler) respondSuccess(w http.ResponseWriter, data interface{}) {
	h.respondJSON(w, http.StatusOK, common.APIResponse{
		Success: true,
		Data:    data,
	})
}

func (h *Handler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.CreatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	policy, err := h.service.CreatePolicy(req)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondSuccess(w, policy)
}

func (h *Handler) SubmitClaim(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.SubmitClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	claim, err := h.service.SubmitClaim(req)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondSuccess(w, claim)
}

func (h *Handler) ReviewClaim(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.ReviewClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	claim, err := h.service.ReviewClaim(req)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondSuccess(w, claim)
}

func (h *Handler) ConfirmPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.ConfirmPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	claim, err := h.service.ConfirmPayment(req)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondSuccess(w, claim)
}

func (h *Handler) GetClaim(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claimID := strings.TrimPrefix(r.URL.Path, "/claims/")
	if claimID == "" {
		h.respondError(w, http.StatusBadRequest, "claim ID is required")
		return
	}

	claim, err := h.service.GetClaim(claimID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}

	h.respondSuccess(w, claim)
}

func (h *Handler) GetPolicyClaims(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	policyNumber := strings.TrimPrefix(r.URL.Path, "/policies/")
	policyNumber = strings.TrimSuffix(policyNumber, "/claims")
	if policyNumber == "" {
		h.respondError(w, http.StatusBadRequest, "policy number is required")
		return
	}

	claims, err := h.service.GetPolicyClaims(policyNumber)
	if err != nil {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}

	h.respondSuccess(w, common.ClaimsListResponse{Claims: claims})
}

func (h *Handler) GetPendingClaims(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims := h.service.GetPendingClaims()
	h.respondSuccess(w, common.ClaimsListResponse{Claims: claims})
}

func (h *Handler) GetApprovedClaims(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims := h.service.GetApprovedClaims()
	h.respondSuccess(w, common.ClaimsListResponse{Claims: claims})
}

func (h *Handler) ListAllPolicies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	policies := h.service.ListAllPolicies()
	h.respondSuccess(w, common.PoliciesListResponse{Policies: policies})
}

func (h *Handler) ListAllClaims(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims := h.service.ListAllClaims()
	h.respondSuccess(w, common.ClaimsListResponse{Claims: claims})
}

func (h *Handler) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/policies", h.ListAllPolicies)
	mux.HandleFunc("/policies/create", h.CreatePolicy)
	mux.HandleFunc("/policies/", h.handlePoliciesPath)

	mux.HandleFunc("/claims", h.ListAllClaims)
	mux.HandleFunc("/claims/submit", h.SubmitClaim)
	mux.HandleFunc("/claims/review", h.ReviewClaim)
	mux.HandleFunc("/claims/pay", h.ConfirmPayment)
	mux.HandleFunc("/claims/pending", h.GetPendingClaims)
	mux.HandleFunc("/claims/approved", h.GetApprovedClaims)
	mux.HandleFunc("/claims/", h.handleClaimsPath)
}

func (h *Handler) handlePoliciesPath(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/policies/")
	parts := strings.Split(path, "/")

	if len(parts) == 2 && parts[1] == "claims" {
		h.GetPolicyClaims(w, r)
		return
	}

	h.respondError(w, http.StatusNotFound, "endpoint not found")
}

func (h *Handler) handleClaimsPath(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/claims/")
	if path != "" && !strings.Contains(path, "/") {
		h.GetClaim(w, r)
		return
	}

	h.respondError(w, http.StatusNotFound, "endpoint not found")
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s\n", r.Method, r.URL.Path)
	h.mux.ServeHTTP(w, r)
}
