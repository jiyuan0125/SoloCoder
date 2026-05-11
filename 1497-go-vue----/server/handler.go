package main

import (
	"encoding/json"
	"marketplace/common"
	"marketplace/core"
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
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *Handler) errorResponse(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, common.BaseResponse{
		Success: false,
		Message: message,
	})
}

func (h *Handler) successResponse(w http.ResponseWriter, data interface{}) {
	switch v := data.(type) {
	case *common.User:
		h.writeJSON(w, http.StatusOK, common.UserResponse{
			BaseResponse: common.BaseResponse{Success: true},
			Data:         v,
		})
	case *common.Item:
		h.writeJSON(w, http.StatusOK, common.ItemResponse{
			BaseResponse: common.BaseResponse{Success: true},
			Data:         v,
		})
	case []*common.Item:
		items := make([]common.Item, 0, len(v))
		for _, item := range v {
			items = append(items, *item)
		}
		h.writeJSON(w, http.StatusOK, common.ItemsResponse{
			BaseResponse: common.BaseResponse{Success: true},
			Data:         items,
		})
	case *common.Negotiation:
		h.writeJSON(w, http.StatusOK, common.NegotiationResponse{
			BaseResponse: common.BaseResponse{Success: true},
			Data:         v,
		})
	case *common.Order:
		h.writeJSON(w, http.StatusOK, common.OrderResponse{
			BaseResponse: common.BaseResponse{Success: true},
			Data:         v,
		})
	default:
		h.writeJSON(w, http.StatusOK, common.BaseResponse{Success: true})
	}
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.errorResponse(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req common.RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	user, err := h.service.RegisterUser(req.Username, req.Role)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.successResponse(w, user)
}

func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.errorResponse(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	sellerID := r.Header.Get("X-User-ID")
	if sellerID == "" {
		h.errorResponse(w, http.StatusBadRequest, "缺少用户ID")
		return
	}

	var req common.CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	item, err := h.service.CreateItem(sellerID, req)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.successResponse(w, item)
}

func (h *Handler) ApproveItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.errorResponse(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	adminID := r.Header.Get("X-User-ID")
	if adminID == "" {
		h.errorResponse(w, http.StatusBadRequest, "缺少用户ID")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/items/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "approve" {
		h.errorResponse(w, http.StatusBadRequest, "路径错误")
		return
	}
	itemID := parts[0]

	var req common.ApproveItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	item, err := h.service.ApproveItem(adminID, itemID, req.Approved, req.Reason)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.successResponse(w, item)
}

func (h *Handler) SearchItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.errorResponse(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	req := common.SearchItemsRequest{
		Keyword:   r.URL.Query().Get("keyword"),
		Category:  common.Category(r.URL.Query().Get("category")),
		Condition: common.Condition(r.URL.Query().Get("condition")),
	}

	if minPriceStr := r.URL.Query().Get("min_price"); minPriceStr != "" {
		var minPrice int64
		if json.Unmarshal([]byte(minPriceStr), &minPrice) == nil {
			req.MinPrice = minPrice
		}
	}
	if maxPriceStr := r.URL.Query().Get("max_price"); maxPriceStr != "" {
		var maxPrice int64
		if json.Unmarshal([]byte(maxPriceStr), &maxPrice) == nil {
			req.MaxPrice = maxPrice
		}
	}

	items := h.service.SearchItems(req)
	h.successResponse(w, items)
}

func (h *Handler) GetItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.errorResponse(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	itemID := strings.TrimPrefix(r.URL.Path, "/items/")
	item, err := h.service.GetItem(itemID)
	if err != nil {
		h.errorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	h.successResponse(w, item)
}

func (h *Handler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.errorResponse(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	sellerID := r.Header.Get("X-User-ID")
	if sellerID == "" {
		h.errorResponse(w, http.StatusBadRequest, "缺少用户ID")
		return
	}

	itemID := strings.TrimPrefix(r.URL.Path, "/items/")
	err := h.service.RemoveItem(sellerID, itemID)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.successResponse(w, nil)
}

func (h *Handler) MakeOffer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.errorResponse(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	buyerID := r.Header.Get("X-User-ID")
	if buyerID == "" {
		h.errorResponse(w, http.StatusBadRequest, "缺少用户ID")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/items/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "offer" {
		h.errorResponse(w, http.StatusBadRequest, "路径错误")
		return
	}
	itemID := parts[0]

	var req common.MakeOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	negotiation, err := h.service.MakeOffer(buyerID, itemID, req.Price)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.successResponse(w, negotiation)
}

func (h *Handler) SellerRespond(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.errorResponse(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	sellerID := r.Header.Get("X-User-ID")
	if sellerID == "" {
		h.errorResponse(w, http.StatusBadRequest, "缺少用户ID")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/negotiations/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "seller-respond" {
		h.errorResponse(w, http.StatusBadRequest, "路径错误")
		return
	}
	negotiationID := parts[0]

	var req common.RespondOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	negotiation, err := h.service.SellerRespond(sellerID, negotiationID, req.Action, req.CounterPrice)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.successResponse(w, negotiation)
}

func (h *Handler) BuyerRespond(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.errorResponse(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	buyerID := r.Header.Get("X-User-ID")
	if buyerID == "" {
		h.errorResponse(w, http.StatusBadRequest, "缺少用户ID")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/negotiations/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "buyer-respond" {
		h.errorResponse(w, http.StatusBadRequest, "路径错误")
		return
	}
	negotiationID := parts[0]

	var req common.RespondOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	negotiation, err := h.service.BuyerRespond(buyerID, negotiationID, req.Action, req.CounterPrice)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.successResponse(w, negotiation)
}

func (h *Handler) PayOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.errorResponse(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	buyerID := r.Header.Get("X-User-ID")
	if buyerID == "" {
		h.errorResponse(w, http.StatusBadRequest, "缺少用户ID")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/orders/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "pay" {
		h.errorResponse(w, http.StatusBadRequest, "路径错误")
		return
	}
	orderID := parts[0]

	order, err := h.service.PayOrder(buyerID, orderID)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.successResponse(w, order)
}

func (h *Handler) ShipOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.errorResponse(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	sellerID := r.Header.Get("X-User-ID")
	if sellerID == "" {
		h.errorResponse(w, http.StatusBadRequest, "缺少用户ID")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/orders/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "ship" {
		h.errorResponse(w, http.StatusBadRequest, "路径错误")
		return
	}
	orderID := parts[0]

	var req common.ShipOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	order, err := h.service.ShipOrder(sellerID, orderID, req.LogisticsNo)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.successResponse(w, order)
}

func (h *Handler) ConfirmReceiveOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.errorResponse(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	buyerID := r.Header.Get("X-User-ID")
	if buyerID == "" {
		h.errorResponse(w, http.StatusBadRequest, "缺少用户ID")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/orders/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "confirm" {
		h.errorResponse(w, http.StatusBadRequest, "路径错误")
		return
	}
	orderID := parts[0]

	order, err := h.service.ConfirmReceiveOrder(buyerID, orderID)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.successResponse(w, order)
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.errorResponse(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	orderID := strings.TrimPrefix(r.URL.Path, "/orders/")
	order, err := h.service.GetOrder(orderID)
	if err != nil {
		h.errorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	h.successResponse(w, order)
}
