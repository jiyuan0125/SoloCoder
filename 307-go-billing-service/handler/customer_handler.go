package handler

import (
	"billing-service/model"
	"billing-service/service"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type CustomerHandler struct {
	customerService *service.CustomerService
}

func NewCustomerHandler(customerService *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{customerService: customerService}
}

type CreateCustomerRequest struct {
	Name    string `json:"name"`
	Package string `json:"package"`
}

type ChangePackageRequest struct {
	NewPackage string `json:"new_package"`
	ChangeDate string `json:"change_date"`
}

func (h *CustomerHandler) HandleCustomers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listCustomers(w, r)
	case http.MethodPost:
		h.createCustomer(w, r)
	default:
		RespondError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *CustomerHandler) HandleCustomerByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/customers/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		RespondError(w, http.StatusBadRequest, "invalid customer ID")
		return
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "invalid customer ID")
		return
	}

	customerID := uint(id)

	if len(parts) > 1 && parts[1] == "change-package" {
		if r.Method == http.MethodPost {
			h.changePackage(w, r, customerID)
			return
		}
		RespondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if len(parts) > 1 && parts[1] == "bills" {
		if r.Method == http.MethodGet {
			h.getCustomerBills(w, r, customerID)
			return
		}
		RespondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getCustomer(w, r, customerID)
	default:
		RespondError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *CustomerHandler) listCustomers(w http.ResponseWriter, r *http.Request) {
	customers, err := h.customerService.GetAllCustomers()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondSuccess(w, customers)
}

func (h *CustomerHandler) getCustomer(w http.ResponseWriter, r *http.Request, id uint) {
	customer, err := h.customerService.GetCustomerByID(id)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if customer == nil {
		RespondError(w, http.StatusNotFound, "customer not found")
		return
	}

	RespondSuccess(w, customer)
}

func (h *CustomerHandler) createCustomer(w http.ResponseWriter, r *http.Request) {
	var req CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		RespondError(w, http.StatusBadRequest, "name is required")
		return
	}

	if req.Package == "" {
		RespondError(w, http.StatusBadRequest, "package is required")
		return
	}

	pkgType := model.PackageType(req.Package)
	customer, err := h.customerService.CreateCustomer(req.Name, pkgType)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if customer == nil {
		RespondError(w, http.StatusBadRequest, "invalid package type")
		return
	}

	RespondSuccess(w, customer)
}

func (h *CustomerHandler) changePackage(w http.ResponseWriter, r *http.Request, customerID uint) {
	var req ChangePackageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.NewPackage == "" {
		RespondError(w, http.StatusBadRequest, "new_package is required")
		return
	}

	var changeDate time.Time
	if req.ChangeDate != "" {
		var err error
		changeDate, err = time.Parse("2006-01-02", req.ChangeDate)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "invalid change_date format, use YYYY-MM-DD")
			return
		}
	} else {
		changeDate = time.Now()
	}

	pkgType := model.PackageType(req.NewPackage)
	if err := h.customerService.ChangePackage(customerID, pkgType, changeDate); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondSuccess(w, map[string]string{"message": "package changed successfully"})
}

func (h *CustomerHandler) getCustomerBills(w http.ResponseWriter, r *http.Request, customerID uint) {
	RespondSuccess(w, map[string]string{"message": "use /api/bills?customer_id=xxx to get customer bills"})
}
