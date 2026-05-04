package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"seating-arrangement/pkg/protocol"
)

type Handler struct {
	venueManager   *VenueManager
	sessionManager *SessionManager
	orderManager   *OrderManager
	storage        *Storage
}

func NewHandler(vm *VenueManager, sm *SessionManager, om *OrderManager, storage *Storage) *Handler {
	return &Handler{
		venueManager:   vm,
		sessionManager: sm,
		orderManager:   om,
		storage:        storage,
	}
}

func (h *Handler) StartLockCleaner(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			h.sessionManager.CheckExpiredLocks()
		}
	}()
}

func (h *Handler) respondJSON(w http.ResponseWriter, statusCode int, success bool, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	resp := protocol.APIResponse{
		Success: success,
		Message: message,
		Data:    data,
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) getStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}
	errMsg := err.Error()

	if strings.Contains(errMsg, "不存在") {
		return http.StatusNotFound
	}

	if strings.Contains(errMsg, "已被选中") ||
		strings.Contains(errMsg, "已售出") ||
		strings.Contains(errMsg, "已取消") ||
		strings.Contains(errMsg, "状态已改变") ||
		strings.Contains(errMsg, "状态不正确") {
		return http.StatusConflict
	}

	if strings.Contains(errMsg, "参数错误") ||
		strings.Contains(errMsg, "缺少必要参数") ||
		strings.Contains(errMsg, "不能为空") ||
		strings.Contains(errMsg, "不能重复") ||
		strings.Contains(errMsg, "不能超过") ||
		strings.Contains(errMsg, "必须大于") ||
		strings.Contains(errMsg, "超时") ||
		strings.Contains(errMsg, "未被锁定") {
		return http.StatusBadRequest
	}

	return http.StatusInternalServerError
}

func (h *Handler) CreateVenue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.CreateVenueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondJSON(w, http.StatusBadRequest, false, "请求参数错误: "+err.Error(), nil)
		return
	}

	venue, err := h.venueManager.CreateVenue(&req)
	if err != nil {
		h.respondJSON(w, h.getStatusCode(err), false, err.Error(), nil)
		return
	}

	if err := h.storage.Save(); err != nil {
		fmt.Printf("保存数据失败: %v\n", err)
	}

	h.respondJSON(w, http.StatusOK, true, "场馆创建成功", venue)
}

func (h *Handler) ListVenues(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	venues := h.venueManager.ListVenues()
	h.respondJSON(w, http.StatusOK, true, "获取成功", venues)
}

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondJSON(w, http.StatusBadRequest, false, "请求参数错误: "+err.Error(), nil)
		return
	}

	session, err := h.sessionManager.CreateSession(&req)
	if err != nil {
		h.respondJSON(w, h.getStatusCode(err), false, err.Error(), nil)
		return
	}

	if err := h.storage.Save(); err != nil {
		fmt.Printf("保存数据失败: %v\n", err)
	}

	h.respondJSON(w, http.StatusOK, true, "场次创建成功", session)
}

func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessions := h.sessionManager.ListSessions()
	h.respondJSON(w, http.StatusOK, true, "获取成功", sessions)
}

func (h *Handler) LockSeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.LockSeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondJSON(w, http.StatusBadRequest, false, "请求参数错误: "+err.Error(), nil)
		return
	}

	order, err := h.sessionManager.LockSeat(req.SessionID, req.SectionName, req.SeatID)
	if err != nil {
		h.respondJSON(w, h.getStatusCode(err), false, err.Error(), nil)
		return
	}

	h.orderManager.AddOrder(order)

	if err := h.storage.Save(); err != nil {
		fmt.Printf("保存数据失败: %v\n", err)
	}

	h.respondJSON(w, http.StatusOK, true, "座位锁定成功", order)
}

func (h *Handler) ConfirmOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.ConfirmOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondJSON(w, http.StatusBadRequest, false, "请求参数错误: "+err.Error(), nil)
		return
	}

	if err := h.orderManager.ConfirmOrder(req.OrderID); err != nil {
		h.respondJSON(w, h.getStatusCode(err), false, err.Error(), nil)
		return
	}

	if err := h.storage.Save(); err != nil {
		fmt.Printf("保存数据失败: %v\n", err)
	}

	h.respondJSON(w, http.StatusOK, true, "订单确认成功", nil)
}

func (h *Handler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.CancelOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondJSON(w, http.StatusBadRequest, false, "请求参数错误: "+err.Error(), nil)
		return
	}

	if err := h.orderManager.CancelOrder(req.OrderID); err != nil {
		h.respondJSON(w, h.getStatusCode(err), false, err.Error(), nil)
		return
	}

	if err := h.storage.Save(); err != nil {
		fmt.Printf("保存数据失败: %v\n", err)
	}

	h.respondJSON(w, http.StatusOK, true, "订单取消成功", nil)
}

func (h *Handler) ReleaseSeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.ReleaseSeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondJSON(w, http.StatusBadRequest, false, "请求参数错误: "+err.Error(), nil)
		return
	}

	if err := h.sessionManager.ReleaseSeat(req.SessionID, req.SeatID, true); err != nil {
		h.respondJSON(w, h.getStatusCode(err), false, err.Error(), nil)
		return
	}

	if err := h.storage.Save(); err != nil {
		fmt.Printf("保存数据失败: %v\n", err)
	}

	h.respondJSON(w, http.StatusOK, true, "座位释放成功", nil)
}

func (h *Handler) GetSessionStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	sessionID = strings.TrimSuffix(sessionID, "/stats")

	stats, err := h.sessionManager.GetSessionStats(sessionID)
	if err != nil {
		h.respondJSON(w, h.getStatusCode(err), false, err.Error(), nil)
		return
	}

	h.respondJSON(w, http.StatusOK, true, "获取成功", stats)
}

func (h *Handler) GetAvailableSeats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	sessionID := query.Get("session_id")
	sectionName := query.Get("section_name")

	if sessionID == "" || sectionName == "" {
		h.respondJSON(w, http.StatusBadRequest, false, "缺少必要参数", nil)
		return
	}

	seats, err := h.sessionManager.GetAvailableSeats(sessionID, sectionName)
	if err != nil {
		h.respondJSON(w, h.getStatusCode(err), false, err.Error(), nil)
		return
	}

	h.respondJSON(w, http.StatusOK, true, "获取成功", seats)
}

func (h *Handler) ExportSoldSeats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	sessionID = strings.TrimSuffix(sessionID, "/sold")

	orderDetails := h.orderManager.GetOrderDetails(sessionID)
	h.respondJSON(w, http.StatusOK, true, "获取成功", orderDetails)
}
