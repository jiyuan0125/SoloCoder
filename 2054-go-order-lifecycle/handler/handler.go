package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"order-lifecycle/models"
	"order-lifecycle/service"
)

type CreateOrderReq struct {
	OrderID   string  `json:"order_id"`
	UserID    string  `json:"user_id"`
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type PayReq struct {
	OrderID       string `json:"order_id"`
	PaymentMethod string `json:"payment_method"`
}

type ErrorResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Handler struct {
	orderService *service.OrderService
}

func NewHandler(orderService *service.OrderService) *Handler {
	return &Handler{orderService: orderService}
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}
	var req CreateOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if req.OrderID == "" || req.UserID == "" || req.ProductID == "" || req.Quantity <= 0 || req.Price <= 0 {
		writeError(w, http.StatusBadRequest, "参数错误")
		return
	}

	order, err := h.orderService.CreateOrder(&service.CreateOrderRequest{
		OrderID:   req.OrderID,
		UserID:    req.UserID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Price:     req.Price,
	})
	if err != nil {
		if strings.Contains(err.Error(), "库存不足") {
			writeJSON(w, http.StatusConflict, map[string]interface{}{
				"code":    http.StatusConflict,
				"message": err.Error(),
			})
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, order)
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}
	orderID := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	if orderID == "" {
		writeError(w, http.StatusBadRequest, "订单号不能为空")
		return
	}

	order, err := h.orderService.GetOrder(orderID)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			writeError(w, http.StatusNotFound, "订单不存在")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	history, _ := h.orderService.GetOrderHistory(orderID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"order":   order,
		"history": history,
	})
}

func (h *Handler) Pay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}
	var req PayReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if req.OrderID == "" || req.PaymentMethod == "" {
		writeError(w, http.StatusBadRequest, "参数错误")
		return
	}

	var method models.PaymentMethod
	switch strings.ToUpper(req.PaymentMethod) {
	case "BALANCE", "余额":
		method = models.PaymentMethodBalance
	case "BANK_CARD", "银行卡":
		method = models.PaymentMethodBankCard
	case "THIRD_PARTY", "第三方":
		method = models.PaymentMethodThirdParty
	default:
		writeError(w, http.StatusBadRequest, "不支持的支付方式")
		return
	}

	err := h.orderService.Pay(&service.PayRequest{
		OrderID:       req.OrderID,
		PaymentMethod: method,
	})
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			writeError(w, http.StatusNotFound, "订单不存在")
			return
		}
		if errors.Is(err, service.ErrDuplicatePayment) {
			writeJSON(w, http.StatusConflict, map[string]interface{}{
				"code":    http.StatusConflict,
				"message": "重复支付",
			})
			return
		}
		if errors.Is(err, service.ErrConflict) {
			writeJSON(w, http.StatusConflict, map[string]interface{}{
				"code":    http.StatusConflict,
				"message": "并发冲突，请稍后重试",
			})
			return
		}
		if errors.Is(err, service.ErrInvalidTransition) {
			order, _ := h.orderService.GetOrder(req.OrderID)
			msg := "状态流转无效"
			if order != nil {
				msg = fmt.Sprintf("无法从状态 %s 进行支付操作", order.Status)
			}
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"code":    http.StatusBadRequest,
				"message": msg,
			})
			return
		}
		if errors.Is(err, service.ErrTerminalState) {
			order, _ := h.orderService.GetOrder(req.OrderID)
			msg := "订单已终止，无法操作"
			if order != nil {
				msg = fmt.Sprintf("订单已处于终态 %s，无法操作", order.Status)
			}
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"code":    http.StatusBadRequest,
				"message": msg,
			})
			return
		}
		if errors.Is(err, service.ErrPaymentFailed) {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"success": false,
				"message": err.Error(),
			})
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "支付成功",
	})
}

func (h *Handler) Ship(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}
	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	orderID := req["order_id"]
	if orderID == "" {
		writeError(w, http.StatusBadRequest, "订单号不能为空")
		return
	}

	trackingNumber, err := h.orderService.Ship(orderID)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			writeError(w, http.StatusNotFound, "订单不存在")
			return
		}
		if errors.Is(err, service.ErrInvalidTransition) {
			order, _ := h.orderService.GetOrder(orderID)
			msg := "状态流转无效"
			if order != nil {
				msg = fmt.Sprintf("当前状态为 %s，无法进行发货操作", order.Status)
			}
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"code":    http.StatusBadRequest,
				"message": msg,
			})
			return
		}
		if errors.Is(err, service.ErrTerminalState) {
			order, _ := h.orderService.GetOrder(orderID)
			msg := "订单已终止，无法操作"
			if order != nil {
				msg = fmt.Sprintf("订单已处于终态 %s，无法操作", order.Status)
			}
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"code":    http.StatusBadRequest,
				"message": msg,
			})
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":         true,
		"tracking_number": trackingNumber,
	})
}

func (h *Handler) Deliver(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}
	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	orderID := req["order_id"]
	if orderID == "" {
		writeError(w, http.StatusBadRequest, "订单号不能为空")
		return
	}

	err := h.orderService.Deliver(orderID)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			writeError(w, http.StatusNotFound, "订单不存在")
			return
		}
		if errors.Is(err, service.ErrInvalidTransition) {
			order, _ := h.orderService.GetOrder(orderID)
			msg := "状态流转无效"
			if order != nil {
				msg = fmt.Sprintf("当前状态为 %s，无法进行签收操作", order.Status)
			}
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"code":    http.StatusBadRequest,
				"message": msg,
			})
			return
		}
		if errors.Is(err, service.ErrTerminalState) {
			order, _ := h.orderService.GetOrder(orderID)
			msg := "订单已终止，无法操作"
			if order != nil {
				msg = fmt.Sprintf("订单已处于终态 %s，无法操作", order.Status)
			}
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"code":    http.StatusBadRequest,
				"message": msg,
			})
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "签收成功",
	})
}

func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}
	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	orderID := req["order_id"]
	if orderID == "" {
		writeError(w, http.StatusBadRequest, "订单号不能为空")
		return
	}

	err := h.orderService.Complete(orderID)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			writeError(w, http.StatusNotFound, "订单不存在")
			return
		}
		if errors.Is(err, service.ErrInvalidTransition) {
			order, _ := h.orderService.GetOrder(orderID)
			msg := "状态流转无效"
			if order != nil {
				msg = fmt.Sprintf("当前状态为 %s，无法进行完成操作", order.Status)
			}
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"code":    http.StatusBadRequest,
				"message": msg,
			})
			return
		}
		if errors.Is(err, service.ErrTerminalState) {
			order, _ := h.orderService.GetOrder(orderID)
			msg := "订单已终止，无法操作"
			if order != nil {
				msg = fmt.Sprintf("订单已处于终态 %s，无法操作", order.Status)
			}
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"code":    http.StatusBadRequest,
				"message": msg,
			})
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "订单完成",
	})
}

func (h *Handler) Refund(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}
	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	orderID := req["order_id"]
	if orderID == "" {
		writeError(w, http.StatusBadRequest, "订单号不能为空")
		return
	}

	err := h.orderService.RequestRefund(orderID)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			writeError(w, http.StatusNotFound, "订单不存在")
			return
		}
		if errors.Is(err, service.ErrRefundWindowExpired) {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"code":    http.StatusBadRequest,
				"message": "签收超过7天，无法申请退货",
			})
			return
		}
		if errors.Is(err, service.ErrInvalidTransition) {
			order, _ := h.orderService.GetOrder(orderID)
			msg := "状态流转无效"
			if order != nil {
				msg = fmt.Sprintf("当前状态为 %s，无法进行退货操作", order.Status)
			}
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"code":    http.StatusBadRequest,
				"message": msg,
			})
			return
		}
		if errors.Is(err, service.ErrTerminalState) {
			order, _ := h.orderService.GetOrder(orderID)
			msg := "订单已终止，无法操作"
			if order != nil {
				msg = fmt.Sprintf("订单已处于终态 %s，无法操作", order.Status)
			}
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"code":    http.StatusBadRequest,
				"message": msg,
			})
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "退货退款成功",
	})
}

func (h *Handler) AddInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}
	var req struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if req.ProductID == "" || req.Quantity <= 0 {
		writeError(w, http.StatusBadRequest, "参数错误")
		return
	}

	if err := h.orderService.AddInventory(req.ProductID, req.Quantity); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "库存添加成功",
	})
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, ErrorResp{Code: code, Message: message})
}

func writeJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
