package main

import (
	"errors"
	"net/http"
	"strings"
)

var (
	errMissingID         = errors.New("missing id parameter")
	errMissingEmployeeID = errors.New("missing employee_id parameter")
)

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case path == "/api/health":
		h.health(w, r)

	case path == "/api/employees" && r.Method == http.MethodGet:
		h.ListEmployees(w, r)
	case path == "/api/employees" && r.Method == http.MethodPost:
		h.CreateEmployee(w, r)
	case strings.HasPrefix(path, "/api/employees/") && r.Method == http.MethodGet:
		h.GetEmployee(w, r)

	case path == "/api/recharge" && r.Method == http.MethodPost:
		h.Recharge(w, r)
	case path == "/api/recharges" && r.Method == http.MethodGet:
		h.ListRechargeRecords(w, r)

	case path == "/api/dishes" && r.Method == http.MethodGet:
		h.ListDishes(w, r)
	case path == "/api/dishes" && r.Method == http.MethodPost:
		h.CreateDish(w, r)
	case strings.HasPrefix(path, "/api/dishes/") && strings.HasSuffix(path, "/onsale") && r.Method == http.MethodPost:
		h.SetDishOnSale(w, r)
	case strings.HasPrefix(path, "/api/dishes/") && r.Method == http.MethodGet:
		h.GetDish(w, r)
	case strings.HasPrefix(path, "/api/dishes/") && r.Method == http.MethodPut:
		h.UpdateDish(w, r)

	case path == "/api/checkout" && r.Method == http.MethodPost:
		h.Checkout(w, r)
	case path == "/api/consumptions" && r.Method == http.MethodGet:
		h.ListConsumptionRecords(w, r)

	case path == "/api/nutrition" && r.Method == http.MethodGet:
		h.GetNutritionSummary(w, r)

	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
