package handler

import (
	"billing/internal/server/service"
	"billing/pkg/api"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type BillingHandler struct {
	service *service.BillingService
}

func NewBillingHandler(svc *service.BillingService) *BillingHandler {
	return &BillingHandler{service: svc}
}

func (h *BillingHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *BillingHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, api.ErrorResponse{Error: message})
}

func (h *BillingHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	var req api.CreatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		h.respondError(w, http.StatusBadRequest, "plan name is required")
		return
	}

	plan, err := h.service.CreatePlan(&req)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, api.CreatePlanResponse{Plan: plan})
}

func (h *BillingHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.respondError(w, http.StatusBadRequest, "plan id is required")
		return
	}

	plan, err := h.service.GetPlan(id)
	if err != nil {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, api.GetPlanResponse{Plan: plan})
}

func (h *BillingHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans := h.service.ListPlans()
	h.respondJSON(w, http.StatusOK, api.ListPlansResponse{Plans: plans})
}

func (h *BillingHandler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var req api.CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Email == "" || req.PlanID == "" {
		h.respondError(w, http.StatusBadRequest, "name, email and plan_id are required")
		return
	}

	customer, err := h.service.CreateCustomer(&req)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, api.CreateCustomerResponse{Customer: customer})
}

func (h *BillingHandler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.respondError(w, http.StatusBadRequest, "customer id is required")
		return
	}

	customer, err := h.service.GetCustomer(id)
	if err != nil {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, api.GetCustomerResponse{Customer: customer})
}

func (h *BillingHandler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	customers := h.service.ListCustomers()
	h.respondJSON(w, http.StatusOK, api.ListCustomersResponse{Customers: customers})
}

func (h *BillingHandler) ChangePlan(w http.ResponseWriter, r *http.Request) {
	var req api.ChangePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.CustomerID == "" || req.NewPlanID == "" {
		h.respondError(w, http.StatusBadRequest, "customer_id and new_plan_id are required")
		return
	}

	if req.EffectiveDate.IsZero() {
		req.EffectiveDate = time.Now()
	}

	change, err := h.service.ChangePlan(&req)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, api.ChangePlanResponse{PlanChange: change})
}

func (h *BillingHandler) RecordUsage(w http.ResponseWriter, r *http.Request) {
	var req api.RecordUsageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.CustomerID == "" {
		h.respondError(w, http.StatusBadRequest, "customer_id is required")
		return
	}

	usage, err := h.service.RecordUsage(&req)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, api.RecordUsageResponse{Usage: usage})
}

func (h *BillingHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	customerID := r.PathValue("customer_id")
	if customerID == "" {
		h.respondError(w, http.StatusBadRequest, "customer_id is required")
		return
	}

	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")

	var year, month int
	var err error

	if yearStr != "" {
		year, err = strconv.Atoi(yearStr)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "invalid year")
			return
		}
	} else {
		year = time.Now().Year()
	}

	if monthStr != "" {
		month, err = strconv.Atoi(monthStr)
		if err != nil || month < 1 || month > 12 {
			h.respondError(w, http.StatusBadRequest, "invalid month")
			return
		}
	} else {
		month = int(time.Now().Month())
	}

	usage, err := h.service.GetUsage(customerID, year, month)
	if err != nil {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, api.GetUsageResponse{Usage: usage})
}

func (h *BillingHandler) GenerateBill(w http.ResponseWriter, r *http.Request) {
	var req api.GenerateBillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Year == 0 || req.Month == 0 {
		h.respondError(w, http.StatusBadRequest, "year and month are required")
		return
	}

	bill, err := h.service.GenerateBill(&req)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, api.GenerateBillResponse{Bill: bill})
}

func (h *BillingHandler) GetBill(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.respondError(w, http.StatusBadRequest, "bill id is required")
		return
	}

	bill, err := h.service.GetBill(id)
	if err != nil {
		h.respondError(w, http.StatusNotFound, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, api.GetBillResponse{Bill: bill})
}

func (h *BillingHandler) ListCustomerBills(w http.ResponseWriter, r *http.Request) {
	customerID := r.PathValue("customer_id")
	if customerID == "" {
		h.respondError(w, http.StatusBadRequest, "customer_id is required")
		return
	}

	bills := h.service.ListCustomerBills(customerID)
	h.respondJSON(w, http.StatusOK, api.ListCustomerBillsResponse{Bills: bills})
}

func (h *BillingHandler) ListAllBills(w http.ResponseWriter, r *http.Request) {
	bills := h.service.ListAllBills()
	h.respondJSON(w, http.StatusOK, api.ListAllBillsResponse{Bills: bills})
}

func (h *BillingHandler) MarkPaid(w http.ResponseWriter, r *http.Request) {
	var req api.MarkPaidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.BillID == "" || req.Operator == "" {
		h.respondError(w, http.StatusBadRequest, "bill_id and operator are required")
		return
	}

	bill, err := h.service.MarkPaid(&req)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, api.MarkPaidResponse{Bill: bill})
}

func (h *BillingHandler) GetPricing(w http.ResponseWriter, r *http.Request) {
	pricing := h.service.GetPricing()
	h.respondJSON(w, http.StatusOK, api.GetPricingResponse{Pricing: pricing})
}

func (h *BillingHandler) UpdatePricing(w http.ResponseWriter, r *http.Request) {
	var req api.UpdatePricingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.UpdatePricing(&req); err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	pricing := h.service.GetPricing()
	h.respondJSON(w, http.StatusOK, api.GetPricingResponse{Pricing: pricing})
}
