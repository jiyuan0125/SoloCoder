package main

import (
	"encoding/json"
	"io"
	"merchant-mgmt-system/internal/core"
	"merchant-mgmt-system/pkg/common"
	"net/http"
	"strings"
)

type Handler struct {
	service *core.Service
}

func NewHandler(service *core.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]interface{}{
		"success": false,
		"message": message,
	})
}

func (h *Handler) decodeBody(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if len(body) == 0 {
		return nil
	}
	return json.Unmarshal(body, v)
}

func (h *Handler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req common.CreateCustomerRequest
	if err := h.decodeBody(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "请求体解析失败: "+err.Error())
		return
	}

	customer, err := h.service.CreateCustomer(&req)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, common.CreateCustomerResponse{
		Success:  true,
		Message:  "创建成功",
		Customer: customer,
	})
}

func (h *Handler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	statusStr := strings.TrimSpace(r.URL.Query().Get("status"))
	status, err := core.ValidateStatus(statusStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	customers, err := h.service.ListCustomers(status)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, common.ListCustomersResponse{
		Success:   true,
		Message:   "查询成功",
		Customers: customers,
	})
}

func (h *Handler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/customers/")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "客户ID不能为空")
		return
	}

	customer, err := h.service.GetCustomer(id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, common.GetCustomerResponse{
		Success:  true,
		Message:  "查询成功",
		Customer: customer,
	})
}

func (h *Handler) CreateFollow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req common.CreateFollowRequest
	if err := h.decodeBody(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "请求体解析失败: "+err.Error())
		return
	}

	follow, err := h.service.CreateFollow(&req)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, common.CreateFollowResponse{
		Success: true,
		Message: "创建成功",
		Follow:  follow,
	})
}

func (h *Handler) ListFollows(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	customerID := strings.TrimSpace(r.URL.Query().Get("customer_id"))

	follows, err := h.service.ListFollows(customerID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, common.ListFollowsResponse{
		Success: true,
		Message: "查询成功",
		Follows: follows,
	})
}

func (h *Handler) CloseFollow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req common.CloseFollowRequest
	if err := h.decodeBody(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "请求体解析失败: "+err.Error())
		return
	}

	if err := h.service.CloseFollow(req.ID); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, common.CloseFollowResponse{
		Success: true,
		Message: "关闭成功",
	})
}

func (h *Handler) CreateContract(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req common.CreateContractRequest
	if err := h.decodeBody(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "请求体解析失败: "+err.Error())
		return
	}

	contract, err := h.service.CreateContract(&req)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, common.CreateContractResponse{
		Success:  true,
		Message:  "签约成功",
		Contract: contract,
	})
}

func (h *Handler) ListContracts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	contracts, err := h.service.ListContracts()
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, common.ListContractsResponse{
		Success:   true,
		Message:   "查询成功",
		Contracts: contracts,
	})
}

func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	stats, err := h.service.GetDashboardStats()
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, common.GetDashboardResponse{
		Success: true,
		Message: "查询成功",
		Stats:   stats,
	})
}
