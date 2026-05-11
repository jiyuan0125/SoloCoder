package main

import (
	"canteen/internal/common"
	"canteen/internal/core"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	service *core.CanteenService
}

func NewHandler(service *core.CanteenService) *Handler {
	return &Handler{service: service}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(common.Response{Success: true, Data: data})
}

func writeError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(common.Response{Success: false, Error: err.Error()})
}

func (h *Handler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var req common.CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	employee := h.service.CreateEmployee(req.Name, req.Gender, req.Weight)
	writeJSON(w, http.StatusCreated, employee)
}

func (h *Handler) GetEmployee(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/employees/")
	if id == "" {
		writeError(w, http.StatusBadRequest, errMissingID)
		return
	}

	employee, err := h.service.GetEmployee(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, employee)
}

func (h *Handler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	employees := h.service.ListEmployees()
	writeJSON(w, http.StatusOK, employees)
}

func (h *Handler) Recharge(w http.ResponseWriter, r *http.Request) {
	var req common.RechargeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	record, err := h.service.Recharge(req.EmployeeID, req.Amount)
	if err != nil {
		status := http.StatusBadRequest
		if err == core.ErrEmployeeNotFound {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}

	writeJSON(w, http.StatusOK, record)
}

func (h *Handler) ListRechargeRecords(w http.ResponseWriter, r *http.Request) {
	employeeID := r.URL.Query().Get("employee_id")
	if employeeID == "" {
		writeError(w, http.StatusBadRequest, errMissingEmployeeID)
		return
	}

	records := h.service.GetRechargeRecords(employeeID)
	writeJSON(w, http.StatusOK, records)
}

func (h *Handler) CreateDish(w http.ResponseWriter, r *http.Request) {
	var req common.CreateDishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	dish := h.service.CreateDish(req.Name, req.Price, req.Nutrition)
	writeJSON(w, http.StatusCreated, dish)
}

func (h *Handler) GetDish(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/dishes/")
	if id == "" {
		writeError(w, http.StatusBadRequest, errMissingID)
		return
	}

	dish, err := h.service.GetDish(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, dish)
}

func (h *Handler) ListDishes(w http.ResponseWriter, r *http.Request) {
	onlyOnSale := r.URL.Query().Get("only_on_sale") == "true"
	dishes := h.service.ListDishes(onlyOnSale)
	writeJSON(w, http.StatusOK, dishes)
}

func (h *Handler) UpdateDish(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/dishes/")
	if id == "" {
		writeError(w, http.StatusBadRequest, errMissingID)
		return
	}

	var req common.UpdateDishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	dish, err := h.service.UpdateDish(id, req.Name, req.Price, req.Nutrition)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, dish)
}

func (h *Handler) SetDishOnSale(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/dishes/")
	id = strings.TrimSuffix(id, "/onsale")
	if id == "" {
		writeError(w, http.StatusBadRequest, errMissingID)
		return
	}

	onSale := r.URL.Query().Get("on_sale") == "true"
	if err := h.service.SetDishOnSale(id, onSale); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	var req common.CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	record, err := h.service.Checkout(req.EmployeeID, req.Items, req.MealType)
	if err != nil {
		status := http.StatusBadRequest
		if err == core.ErrEmployeeNotFound || err == core.ErrDishNotFound {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}

	writeJSON(w, http.StatusOK, record)
}

func (h *Handler) ListConsumptionRecords(w http.ResponseWriter, r *http.Request) {
	employeeID := r.URL.Query().Get("employee_id")
	if employeeID == "" {
		writeError(w, http.StatusBadRequest, errMissingEmployeeID)
		return
	}

	records := h.service.GetConsumptionRecords(employeeID)
	writeJSON(w, http.StatusOK, records)
}

func (h *Handler) GetNutritionSummary(w http.ResponseWriter, r *http.Request) {
	employeeID := r.URL.Query().Get("employee_id")
	if employeeID == "" {
		writeError(w, http.StatusBadRequest, errMissingEmployeeID)
		return
	}

	daysStr := r.URL.Query().Get("days")
	days := 7
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil {
			days = d
		}
	}

	summary, err := h.service.GetNutritionSummary(employeeID, days)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}
