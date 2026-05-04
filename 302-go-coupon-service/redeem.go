package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type RedeemRequest struct {
	CouponCode  string  `json:"coupon_code"`
	UserID      string  `json:"user_id"`
	OrderID     string  `json:"order_id"`
	OrderAmount float64 `json:"order_amount"`
}

type RedeemResponse struct {
	ID             string    `json:"id"`
	BatchID        string    `json:"batch_id"`
	UserID         string    `json:"user_id"`
	CouponCode     string    `json:"coupon_code"`
	OrderID        string    `json:"order_id"`
	OrderAmount    float64   `json:"order_amount"`
	DiscountAmount float64   `json:"discount_amount"`
	RedeemedAt     time.Time `json:"redeemed_at"`
}

type RedeemHandler struct {
	storage  *Storage
	redeemMu sync.Mutex
}

func NewRedeemHandler(storage *Storage) *RedeemHandler {
	return &RedeemHandler{storage: storage}
}

func (h *RedeemHandler) RedeemCoupon(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RedeemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.CouponCode == "" || req.UserID == "" || req.OrderID == "" {
		http.Error(w, "Coupon code, User ID, and Order ID are required", http.StatusBadRequest)
		return
	}

	if req.OrderAmount <= 0 {
		http.Error(w, "Order amount must be greater than zero", http.StatusBadRequest)
		return
	}

	h.redeemMu.Lock()
	defer h.redeemMu.Unlock()

	claim, exists := h.storage.GetClaimByCode(req.CouponCode)
	if !exists {
		http.Error(w, "Coupon not found", http.StatusNotFound)
		return
	}

	if claim.UserID != req.UserID {
		http.Error(w, "Coupon does not belong to this user", http.StatusBadRequest)
		return
	}

	if claim.IsRedeemed {
		http.Error(w, "Coupon has already been redeemed", http.StatusBadRequest)
		return
	}

	batch, exists := h.storage.GetBatch(claim.BatchID)
	if !exists {
		http.Error(w, "Coupon batch not found", http.StatusNotFound)
		return
	}

	if !IsValidNow(batch.ValidFrom, batch.ValidTo) {
		http.Error(w, "Coupon has expired or is not yet valid", http.StatusBadRequest)
		return
	}

	if !IsThresholdMet(req.OrderAmount, batch.ThresholdAmount) {
		http.Error(w, "Order amount does not meet the threshold", http.StatusBadRequest)
		return
	}

	now := time.Now()
	claim.IsRedeemed = true
	claim.RedeemedAt = &now

	if err := h.storage.UpdateClaim(claim); err != nil {
		http.Error(w, "Failed to update claim", http.StatusInternalServerError)
		return
	}

	batch.RedeemedQuantity++
	if err := h.storage.UpdateBatch(batch); err != nil {
		http.Error(w, "Failed to update batch", http.StatusInternalServerError)
		return
	}

	redeem := &RedeemRecord{
		ID:             GenerateID(),
		BatchID:        batch.ID,
		UserID:         req.UserID,
		CouponCode:     req.CouponCode,
		OrderID:        req.OrderID,
		OrderAmount:    req.OrderAmount,
		DiscountAmount: batch.DiscountAmount,
		RedeemedAt:     now,
	}

	if err := h.storage.CreateRedeem(redeem); err != nil {
		http.Error(w, "Failed to create redeem record", http.StatusInternalServerError)
		return
	}

	response := RedeemResponse{
		ID:             redeem.ID,
		BatchID:        redeem.BatchID,
		UserID:         redeem.UserID,
		CouponCode:     redeem.CouponCode,
		OrderID:        redeem.OrderID,
		OrderAmount:    redeem.OrderAmount,
		DiscountAmount: redeem.DiscountAmount,
		RedeemedAt:     redeem.RedeemedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *RedeemHandler) GetRedeemRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	redeemID := r.URL.Query().Get("id")
	if redeemID == "" {
		http.Error(w, "Redeem ID is required", http.StatusBadRequest)
		return
	}

	redeem, exists := h.storage.GetRedeem(redeemID)
	if !exists {
		http.Error(w, "Redeem record not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(redeem)
}

func (h *RedeemHandler) ValidateCoupon(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	couponCode := r.URL.Query().Get("coupon_code")
	userID := r.URL.Query().Get("user_id")
	orderAmountStr := r.URL.Query().Get("order_amount")

	if couponCode == "" || userID == "" {
		http.Error(w, "Coupon code and User ID are required", http.StatusBadRequest)
		return
	}

	claim, exists := h.storage.GetClaimByCode(couponCode)
	if !exists {
		http.Error(w, "Coupon not found", http.StatusNotFound)
		return
	}

	if claim.UserID != userID {
		http.Error(w, "Coupon does not belong to this user", http.StatusBadRequest)
		return
	}

	if claim.IsRedeemed {
		http.Error(w, "Coupon has already been redeemed", http.StatusBadRequest)
		return
	}

	batch, exists := h.storage.GetBatch(claim.BatchID)
	if !exists {
		http.Error(w, "Coupon batch not found", http.StatusNotFound)
		return
	}

	if !IsValidNow(batch.ValidFrom, batch.ValidTo) {
		http.Error(w, "Coupon has expired or is not yet valid", http.StatusBadRequest)
		return
	}

	var orderAmount float64
	if orderAmountStr != "" {
		var err error
		orderAmount, err = strconv.ParseFloat(orderAmountStr, 64)
		if err != nil {
			http.Error(w, "Invalid order amount", http.StatusBadRequest)
			return
		}
		if orderAmount > 0 && !IsThresholdMet(orderAmount, batch.ThresholdAmount) {
			http.Error(w, "Order amount does not meet the threshold", http.StatusBadRequest)
			return
		}
	}

	response := map[string]interface{}{
		"valid":           true,
		"coupon_code":     couponCode,
		"batch_id":        batch.ID,
		"batch_name":      batch.Name,
		"discount_amount": batch.DiscountAmount,
		"threshold_amount": batch.ThresholdAmount,
		"valid_from":      batch.ValidFrom,
		"valid_to":        batch.ValidTo,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
