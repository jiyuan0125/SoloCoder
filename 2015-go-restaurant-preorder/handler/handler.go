package handler

import (
	"encoding/json"
	"net/http"
	"restaurant-preorder/models"
	"restaurant-preorder/service"
	"strconv"
	"strings"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, err string) {
	WriteJSON(w, status, Response{
		Success: false,
		Error:   err,
	})
}

func WriteSuccess(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req service.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	order, err := service.CreateOrder(req)
	if err != nil {
		switch err {
		case service.ErrInvalidTimeSlot, service.ErrInvalidGuestCount:
			WriteError(w, http.StatusBadRequest, err.Error())
		case service.ErrSlotFull:
			WriteError(w, http.StatusConflict, "该时段已满")
		case service.ErrDuplicateBooking:
			WriteError(w, http.StatusConflict, "该手机号当天已预订")
		case service.ErrBlacklisted:
			WriteError(w, http.StatusForbidden, "该手机号已被拉黑")
		default:
			WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	WriteSuccess(w, order)
}

func CancelOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	idStr = strings.TrimSuffix(idStr, "/cancel")
	orderID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	err = service.CancelOrder(orderID)
	if err != nil {
		switch err {
		case service.ErrOrderNotFound:
			WriteError(w, http.StatusNotFound, err.Error())
		case service.ErrTooLateToCancel:
			WriteError(w, http.StatusBadRequest, "距就餐不足 2 小时")
		case service.ErrAlreadyCancelled:
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	WriteSuccess(w, map[string]string{"message": "订单已取消"})
}

func UpdateOrderStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
		WriteError(w, http.StatusBadRequest, "Invalid URL path")
		return
	}

	orderID, err := strconv.ParseInt(pathParts[3], 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	newStatus := models.OrderStatus(pathParts[5])

	err = service.UpdateOrderStatus(orderID, newStatus)
	if err != nil {
		switch err {
		case service.ErrOrderNotFound:
			WriteError(w, http.StatusNotFound, err.Error())
		case service.ErrInvalidStatus, service.ErrInvalidStatusTransition:
			WriteError(w, http.StatusBadRequest, err.Error())
		default:
			WriteError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	WriteSuccess(w, map[string]string{"message": "状态已更新"})
}

func GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	orderID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	order, err := service.GetOrder(orderID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "Order not found")
		return
	}

	WriteSuccess(w, order)
}

func GetAllOrdersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	orders, err := service.GetAllOrders()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, orders)
}

func GetKitchenSummaryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		WriteError(w, http.StatusBadRequest, "Date parameter is required")
		return
	}

	summary, err := service.GetKitchenSummary(date)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, summary)
}

func GetResourceSummaryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	resourceType := models.ResourceType(r.URL.Query().Get("type"))
	resourceID := r.URL.Query().Get("id")

	if resourceType == "" || resourceID == "" {
		WriteError(w, http.StatusBadRequest, "type and id parameters are required")
		return
	}

	summary, err := service.GetResourceSummary(resourceType, resourceID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, summary)
}
