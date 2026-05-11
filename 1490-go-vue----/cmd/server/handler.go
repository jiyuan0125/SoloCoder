package main

import (
	"encoding/json"
	"io"
	"net/http"

	"piperepair/api"
	"piperepair/core"
)

type Handler struct {
	service *core.PipeRepairService
}

func NewHandler(service *core.PipeRepairService) *Handler {
	return &Handler{service: service}
}

func writeJSONResponse(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := api.Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
	json.NewEncoder(w).Encode(resp)
}

func readJSONBody(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.Unmarshal(body, v)
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONResponse(w, 405, "方法不允许", nil)
		return
	}

	var req api.CreateRepairOrderRequest
	if err := readJSONBody(r, &req); err != nil {
		writeJSONResponse(w, 400, "请求参数错误: "+err.Error(), nil)
		return
	}

	order, master, err := h.service.CreateRepairOrder(&req)
	if err != nil {
		writeJSONResponse(w, 500, "创建工单失败: "+err.Error(), nil)
		return
	}

	resp := api.CreateRepairOrderResponse{
		OrderID: order.ID,
		Status:  string(order.Status),
	}

	if master != nil {
		resp.MasterName = master.Name
		_, _, site, _, _ := h.service.GetOrder(order.ID)
		if site != nil {
			resp.SiteName = site.Name
		}
	}

	writeJSONResponse(w, 200, "创建工单成功", resp)
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONResponse(w, 405, "方法不允许", nil)
		return
	}

	orderID := r.URL.Query().Get("order_id")
	if orderID == "" {
		writeJSONResponse(w, 400, "缺少工单ID参数", nil)
		return
	}

	order, master, site, record, err := h.service.GetOrder(orderID)
	if err != nil {
		writeJSONResponse(w, 404, err.Error(), nil)
		return
	}

	resp := api.GetOrderResponse{
		Order:        order,
		RepairRecord: record,
	}

	if master != nil {
		resp.MasterName = master.Name
	}

	if site != nil {
		resp.SiteName = site.Name
	}

	writeJSONResponse(w, 200, "查询成功", resp)
}

func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONResponse(w, 405, "方法不允许", nil)
		return
	}

	orders := h.service.ListAllOrders()
	resp := api.ListOrdersResponse{
		Orders: orders,
	}

	writeJSONResponse(w, 200, "查询成功", resp)
}

func (h *Handler) StartProcessing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONResponse(w, 405, "方法不允许", nil)
		return
	}

	orderID := r.URL.Query().Get("order_id")
	masterID := r.URL.Query().Get("master_id")

	if orderID == "" || masterID == "" {
		writeJSONResponse(w, 400, "缺少必要参数", nil)
		return
	}

	err := h.service.StartProcessing(orderID, masterID)
	if err != nil {
		writeJSONResponse(w, 500, "开始处理失败: "+err.Error(), nil)
		return
	}

	writeJSONResponse(w, 200, "开始处理成功", nil)
}

func (h *Handler) SubmitRepairRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONResponse(w, 405, "方法不允许", nil)
		return
	}

	var req api.SubmitRepairRecordRequest
	if err := readJSONBody(r, &req); err != nil {
		writeJSONResponse(w, 400, "请求参数错误: "+err.Error(), nil)
		return
	}

	record, err := h.service.SubmitRepairRecord(&req)
	if err != nil {
		writeJSONResponse(w, 500, "提交维修记录失败: "+err.Error(), nil)
		return
	}

	writeJSONResponse(w, 200, "提交维修记录成功", record)
}

func (h *Handler) AcceptOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONResponse(w, 405, "方法不允许", nil)
		return
	}

	var req api.AcceptOrderRequest
	if err := readJSONBody(r, &req); err != nil {
		writeJSONResponse(w, 400, "请求参数错误: "+err.Error(), nil)
		return
	}

	err := h.service.AcceptOrder(&req)
	if err != nil {
		writeJSONResponse(w, 500, "验收失败: "+err.Error(), nil)
		return
	}

	if req.Accepted {
		writeJSONResponse(w, 200, "验收成功，工单已完成", nil)
	} else {
		writeJSONResponse(w, 200, "验收驳回，系统将重新派单", nil)
	}
}

func (h *Handler) GetMasterInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONResponse(w, 405, "方法不允许", nil)
		return
	}

	masterID := r.URL.Query().Get("master_id")
	if masterID == "" {
		writeJSONResponse(w, 400, "缺少师傅ID参数", nil)
		return
	}

	master, currentOrder, err := h.service.GetMasterInfo(masterID)
	if err != nil {
		writeJSONResponse(w, 404, err.Error(), nil)
		return
	}

	resp := api.MasterInfoResponse{
		Master:       master,
		CurrentOrder: currentOrder,
		TodayOrders:  master.DailyOrders,
	}

	writeJSONResponse(w, 200, "查询成功", resp)
}
