package handler

import (
	"billing-service/service"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type BillHandler struct {
	billService *service.BillService
}

func NewBillHandler(billService *service.BillService) *BillHandler {
	return &BillHandler{billService: billService}
}

type GenerateBillRequest struct {
	CustomerID uint `json:"customer_id"`
	Year       int  `json:"year"`
	Month      int  `json:"month"`
}

type MarkPaidRequest struct {
	Operator string `json:"operator"`
	Remark   string `json:"remark"`
}

func (h *BillHandler) HandleBills(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listBills(w, r)
	case http.MethodPost:
		h.generateBill(w, r)
	default:
		RespondError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *BillHandler) HandleBillByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/bills/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		RespondError(w, http.StatusBadRequest, "invalid bill ID")
		return
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "invalid bill ID")
		return
	}

	billID := uint(id)

	if len(parts) > 1 && parts[1] == "detail" {
		if r.Method == http.MethodGet {
			h.getBillDetail(w, r, billID)
			return
		}
		RespondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if len(parts) > 1 && parts[1] == "payments" {
		if r.Method == http.MethodGet {
			h.getPaymentHistory(w, r, billID)
			return
		}
		RespondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if len(parts) > 1 && parts[1] == "mark-paid" {
		if r.Method == http.MethodPost {
			h.markAsPaid(w, r, billID)
			return
		}
		RespondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getBill(w, r, billID)
	default:
		RespondError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *BillHandler) listBills(w http.ResponseWriter, r *http.Request) {
	customerIDStr := r.URL.Query().Get("customer_id")
	if customerIDStr != "" {
		customerID, err := strconv.ParseUint(customerIDStr, 10, 64)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "invalid customer_id")
			return
		}

		bills, err := h.billService.GetCustomerBills(uint(customerID))
		if err != nil {
			RespondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		RespondSuccess(w, bills)
		return
	}

	bills, err := h.billService.GetAllBills()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondSuccess(w, bills)
}

func (h *BillHandler) getBill(w http.ResponseWriter, r *http.Request, id uint) {
	bill, err := h.billService.GetBillByID(id)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if bill == nil {
		RespondError(w, http.StatusNotFound, "bill not found")
		return
	}

	RespondSuccess(w, bill)
}

func (h *BillHandler) getBillDetail(w http.ResponseWriter, r *http.Request, id uint) {
	detail, err := h.billService.GetBillDetail(id)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if detail == nil {
		RespondError(w, http.StatusNotFound, "bill not found")
		return
	}

	RespondSuccess(w, detail)
}

func (h *BillHandler) getPaymentHistory(w http.ResponseWriter, r *http.Request, id uint) {
	payments, err := h.billService.GetPaymentHistory(id)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondSuccess(w, payments)
}

func (h *BillHandler) generateBill(w http.ResponseWriter, r *http.Request) {
	var req GenerateBillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.CustomerID == 0 {
		RespondError(w, http.StatusBadRequest, "customer_id is required")
		return
	}

	if req.Year == 0 {
		RespondError(w, http.StatusBadRequest, "year is required")
		return
	}

	if req.Month == 0 || req.Month > 12 {
		RespondError(w, http.StatusBadRequest, "month is required and must be 1-12")
		return
	}

	bill, err := h.billService.GenerateMonthlyBill(req.CustomerID, req.Year, req.Month)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondSuccess(w, bill)
}

func (h *BillHandler) markAsPaid(w http.ResponseWriter, r *http.Request, billID uint) {
	var req MarkPaidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Operator == "" {
		RespondError(w, http.StatusBadRequest, "operator is required")
		return
	}

	if err := h.billService.MarkBillAsPaid(billID, req.Operator, req.Remark); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondSuccess(w, map[string]string{"message": "bill marked as paid successfully"})
}
