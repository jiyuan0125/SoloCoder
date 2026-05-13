package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"flashsale/internal/model"
	"flashsale/internal/service"
)

type JSONResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(JSONResponse{Success: true, Data: data})
}

func writeError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(JSONResponse{Success: false, Error: err.Error()})
}

type ActivityHandler struct {
	service *service.ActivityService
}

func NewActivityHandler(service *service.ActivityService) *ActivityHandler {
	return &ActivityHandler{service: service}
}

func (h *ActivityHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/activities", h.handleActivities)
	mux.HandleFunc("/api/activities/", h.handleActivityByID)
}

func (h *ActivityHandler) handleActivities(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listActivities(w, r)
	case http.MethodPost:
		h.createActivity(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ActivityHandler) handleActivityByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/activities/")

	if strings.Contains(path, "/purchase") {
		parts := strings.Split(path, "/purchase")
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if r.Method == http.MethodPost {
			h.purchase(w, r, id)
			return
		}
	}

	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getActivity(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ActivityHandler) listActivities(w http.ResponseWriter, r *http.Request) {
	activities, err := h.service.ListActivities()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, activities)
}

type CreateActivityRequest struct {
	ProductName  string  `json:"product_name"`
	OriginalPrice float64 `json:"original_price"`
	FlashPrice   float64 `json:"flash_price"`
	TotalStock   int     `json:"total_stock"`
	StartTime    string  `json:"start_time"`
}

func (h *ActivityHandler) createActivity(w http.ResponseWriter, r *http.Request) {
	var req CreateActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.FlashPrice >= req.OriginalPrice {
		writeError(w, http.StatusBadRequest, service.ErrFlashPriceTooHigh)
		return
	}

	activity, err := h.service.CreateActivity(
		req.ProductName,
		req.OriginalPrice,
		req.FlashPrice,
		req.TotalStock,
		startTime,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusCreated, activity)
}

func (h *ActivityHandler) getActivity(w http.ResponseWriter, r *http.Request, id int64) {
	activity, stock, err := h.service.GetActivityWithStock(id)
	if err != nil {
		if err == service.ErrActivityNotFound {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	response := map[string]interface{}{
		"activity":        activity,
		"available_stock": stock.AvailableStock,
	}

	writeJSON(w, http.StatusOK, response)
}

type PurchaseRequest struct {
	UserID string `json:"user_id"`
}

func (h *ActivityHandler) purchase(w http.ResponseWriter, r *http.Request, activityID int64) {
	var req PurchaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	order, err := h.service.Purchase(activityID, req.UserID)
	if err != nil {
		switch err {
		case service.ErrActivityNotFound:
			writeError(w, http.StatusNotFound, err)
		case service.ErrActivityNotStarted:
			writeError(w, http.StatusBadRequest, err)
		case service.ErrStockEmpty:
			writeError(w, http.StatusGone, err)
		case service.ErrAlreadyPurchased:
			writeError(w, http.StatusConflict, err)
		case service.ErrPurchaseFailed:
			writeError(w, http.StatusServiceUnavailable, err)
		default:
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, http.StatusCreated, order)
}

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(service *service.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/orders/", h.handleOrder)
}

func (h *OrderHandler) handleOrder(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	orderNo := parts[0]

	if len(parts) >= 2 {
		action := parts[1]
		switch r.Method {
		case http.MethodPost:
			switch action {
			case "pay":
				h.payOrder(w, r, orderNo)
				return
			case "cancel":
				h.cancelOrder(w, r, orderNo)
				return
			}
		}
	}

	switch r.Method {
	case http.MethodGet:
		h.getOrder(w, r, orderNo)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *OrderHandler) getOrder(w http.ResponseWriter, r *http.Request, orderNo string) {
	order, err := h.service.GetOrder(orderNo)
	if err != nil {
		if err == service.ErrOrderNotFound {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (h *OrderHandler) payOrder(w http.ResponseWriter, r *http.Request, orderNo string) {
	err := h.service.PayOrder(orderNo)
	if err != nil {
		switch err {
		case service.ErrOrderNotFound:
			writeError(w, http.StatusNotFound, err)
		case service.ErrOrderAlreadyPaid:
			writeError(w, http.StatusConflict, err)
		case service.ErrOrderAlreadyCancelled:
			writeError(w, http.StatusGone, err)
		case service.ErrOrderExpired:
			writeError(w, http.StatusGone, err)
		default:
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": string(model.OrderStatusPaid)})
}

func (h *OrderHandler) cancelOrder(w http.ResponseWriter, r *http.Request, orderNo string) {
	err := h.service.CancelOrder(orderNo)
	if err != nil {
		switch err {
		case service.ErrOrderNotFound:
			writeError(w, http.StatusNotFound, err)
		case service.ErrOrderAlreadyPaid:
			writeError(w, http.StatusConflict, err)
		case service.ErrOrderAlreadyCancelled:
			writeError(w, http.StatusGone, err)
		default:
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": string(model.OrderStatusCancelled)})
}

type ReportHandler struct {
	service *service.ReportService
}

func NewReportHandler(service *service.ReportService) *ReportHandler {
	return &ReportHandler{service: service}
}

func (h *ReportHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/reports/", h.handleReport)
}

func (h *ReportHandler) handleReport(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/reports/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	activityIDStr := parts[0]
	activityID, err := strconv.ParseInt(activityIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if len(parts) >= 2 && parts[1] == "generate" && r.Method == http.MethodPost {
		h.generateReport(w, r, activityID)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getReport(w, r, activityID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ReportHandler) getReport(w http.ResponseWriter, r *http.Request, activityID int64) {
	report, err := h.service.GetReport(activityID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (h *ReportHandler) generateReport(w http.ResponseWriter, r *http.Request, activityID int64) {
	report, err := h.service.GenerateReport(activityID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

type CacheHandler struct {
	service *service.CacheService
}

func NewCacheHandler(service *service.CacheService) *CacheHandler {
	return &CacheHandler{service: service}
}

func (h *CacheHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/quotas/", h.handleQuotas)
}

type AllocateQuotaRequest struct {
	Segments map[string]float64 `json:"segments"`
}

func (h *CacheHandler) handleQuotas(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/quotas/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	activityIDStr := parts[0]
	activityID, err := strconv.ParseInt(activityIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if len(parts) >= 2 && parts[1] == "allocate" && r.Method == http.MethodPost {
		h.allocateQuotas(w, r, activityID)
		return
	}

	if len(parts) >= 2 && parts[1] == "redistribute" && r.Method == http.MethodPost {
		h.redistributeQuotas(w, r, activityID)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getQuotas(w, r, activityID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *CacheHandler) getQuotas(w http.ResponseWriter, r *http.Request, activityID int64) {
	quotas, err := h.service.GetQuotaAllocations(activityID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, quotas)
}

func (h *CacheHandler) allocateQuotas(w http.ResponseWriter, r *http.Request, activityID int64) {
	var req AllocateQuotaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	quotas, err := h.service.AllocateQuotas(activityID, req.Segments)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, quotas)
}

func (h *CacheHandler) redistributeQuotas(w http.ResponseWriter, r *http.Request, activityID int64) {
	if err := h.service.RedistributeQuotas(activityID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
