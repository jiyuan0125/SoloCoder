package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"return-exchange/service"
)

// Handler HTTP请求处理器
type Handler struct {
	service *service.ReturnExchangeService
}

func NewHandler() *Handler {
	return &Handler{
		service: service.NewReturnExchangeService(),
	}
}

// APIResponse 统一API响应格式
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// CreateReturnRequestRequest 创建退货申请请求
type CreateReturnRequestRequest struct {
	UserID         int64    `json:"user_id"`
	OrderNo        string   `json:"order_no"`
	Reason         string   `json:"reason"`
	ReasonDetail   string   `json:"reason_detail"`
	EvidenceImages []string `json:"evidence_images"`
}

// CreateExchangeRequestRequest 创建换货申请请求
type CreateExchangeRequestRequest struct {
	UserID           int64    `json:"user_id"`
	OrderNo          string   `json:"order_no"`
	Reason           string   `json:"reason"`
	ReasonDetail     string   `json:"reason_detail"`
	EvidenceImages   []string `json:"evidence_images"`
	NewSKU           string   `json:"new_sku"`
	NewSpecification string   `json:"new_specification"`
	NewPrice         float64  `json:"new_price"`
	ShippingAddress  string   `json:"shipping_address"`
}

// ApproveRequestRequest 审核通过请求
type ApproveRequestRequest struct {
	AdminID         int64  `json:"admin_id"`
	Comment         string `json:"comment"`
	ShippingAddress string `json:"shipping_address"`
}

// RejectRequestRequest 审核拒绝请求
type RejectRequestRequest struct {
	AdminID int64  `json:"admin_id"`
	Comment string `json:"comment"`
}

// respondJSON 返回JSON响应
func respondJSON(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(APIResponse{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// respondError 返回错误响应
func respondError(w http.ResponseWriter, code int, message string) {
	respondJSON(w, code, message, nil)
}

// CreateReturnRequestHandler 创建退货申请
func (h *Handler) CreateReturnRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req CreateReturnRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求参数解析失败")
		return
	}

	// 验证必填参数
	if req.UserID <= 0 {
		respondError(w, http.StatusBadRequest, "用户ID不能为空")
		return
	}
	if req.OrderNo == "" {
		respondError(w, http.StatusBadRequest, "订单号不能为空")
		return
	}
	if req.Reason == "" {
		respondError(w, http.StatusBadRequest, "退货原因不能为空")
		return
	}

	re, err := h.service.CreateReturnRequest(
		req.UserID,
		req.OrderNo,
		req.Reason,
		req.ReasonDetail,
		req.EvidenceImages,
	)

	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, "退货申请创建成功", re)
}

// CreateExchangeRequestHandler 创建换货申请
func (h *Handler) CreateExchangeRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req CreateExchangeRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求参数解析失败")
		return
	}

	// 验证必填参数
	if req.UserID <= 0 {
		respondError(w, http.StatusBadRequest, "用户ID不能为空")
		return
	}
	if req.OrderNo == "" {
		respondError(w, http.StatusBadRequest, "订单号不能为空")
		return
	}
	if req.Reason == "" {
		respondError(w, http.StatusBadRequest, "换货原因不能为空")
		return
	}
	if req.NewSKU == "" {
		respondError(w, http.StatusBadRequest, "新SKU不能为空")
		return
	}
	if req.ShippingAddress == "" {
		respondError(w, http.StatusBadRequest, "收货地址不能为空")
		return
	}

	re, err := h.service.CreateExchangeRequest(
		req.UserID,
		req.OrderNo,
		req.Reason,
		req.ReasonDetail,
		req.EvidenceImages,
		req.NewSKU,
		req.NewSpecification,
		req.NewPrice,
		req.ShippingAddress,
	)

	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, "换货申请创建成功", re)
}

// GetRequestByIDHandler 获取申请详情
func (h *Handler) GetRequestByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从URL路径获取ID
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, "申请ID不能为空")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "申请ID格式错误")
		return
	}

	re, err := h.service.GetRequestByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if re == nil {
		respondError(w, http.StatusNotFound, "申请不存在")
		return
	}

	respondJSON(w, http.StatusOK, "获取成功", re)
}

// GetRequestsByUserIDHandler 获取用户的所有申请
func (h *Handler) GetRequestsByUserIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从URL路径获取用户ID
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		respondError(w, http.StatusBadRequest, "用户ID不能为空")
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "用户ID格式错误")
		return
	}

	list, err := h.service.GetRequestsByUserID(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, "获取成功", list)
}

// ApproveRequestHandler 审核通过申请
func (h *Handler) ApproveRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从URL路径获取申请ID
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, "申请ID不能为空")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "申请ID格式错误")
		return
	}

	var req ApproveRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求参数解析失败")
		return
	}

	if req.AdminID <= 0 {
		respondError(w, http.StatusBadRequest, "管理员ID不能为空")
		return
	}

	err = h.service.ApproveRequest(id, req.AdminID, req.Comment, req.ShippingAddress)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, "审核通过成功", nil)
}

// RejectRequestHandler 审核拒绝申请
func (h *Handler) RejectRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从URL路径获取申请ID
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, "申请ID不能为空")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "申请ID格式错误")
		return
	}

	var req RejectRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求参数解析失败")
		return
	}

	if req.AdminID <= 0 {
		respondError(w, http.StatusBadRequest, "管理员ID不能为空")
		return
	}

	err = h.service.RejectRequest(id, req.AdminID, req.Comment)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, "审核拒绝成功", nil)
}

// GetRefundByRequestIDHandler 获取退款记录
func (h *Handler) GetRefundByRequestIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从URL路径获取申请ID
	idStr := r.URL.Query().Get("request_id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, "申请ID不能为空")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "申请ID格式错误")
		return
	}

	refund, err := h.service.GetRefundByRequestID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if refund == nil {
		respondError(w, http.StatusNotFound, "退款记录不存在")
		return
	}

	respondJSON(w, http.StatusOK, "获取成功", refund)
}

// GetShipmentByRequestIDHandler 获取发货单
func (h *Handler) GetShipmentByRequestIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从URL路径获取申请ID
	idStr := r.URL.Query().Get("request_id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, "申请ID不能为空")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "申请ID格式错误")
		return
	}

	shipment, err := h.service.GetShipmentByRequestID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if shipment == nil {
		respondError(w, http.StatusNotFound, "发货单不存在")
		return
	}

	respondJSON(w, http.StatusOK, "获取成功", shipment)
}

// GetOrderByNoHandler 获取订单信息（测试用）
func (h *Handler) GetOrderByNoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	orderNo := r.URL.Query().Get("order_no")
	if orderNo == "" {
		respondError(w, http.StatusBadRequest, "订单号不能为空")
		return
	}

	order, err := h.service.GetOrderByNo(orderNo)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if order == nil {
		respondError(w, http.StatusNotFound, "订单不存在")
		return
	}

	respondJSON(w, http.StatusOK, "获取成功", order)
}
