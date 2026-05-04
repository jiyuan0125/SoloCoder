package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type ClaimRequest struct {
	BatchID string `json:"batch_id"`
	UserID  string `json:"user_id"`
}

type ReturnRequest struct {
	BatchID    string `json:"batch_id"`
	UserID     string `json:"user_id"`
	CouponCode string `json:"coupon_code"`
}

type ClaimHandler struct {
	storage *Storage
	claimMu sync.Mutex
}

func NewClaimHandler(storage *Storage) *ClaimHandler {
	return &ClaimHandler{storage: storage}
}

func (h *ClaimHandler) ClaimCoupon(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.BatchID == "" || req.UserID == "" {
		http.Error(w, "Batch ID and User ID are required", http.StatusBadRequest)
		return
	}

	h.claimMu.Lock()
	defer h.claimMu.Unlock()

	batch, exists := h.storage.GetBatch(req.BatchID)
	if !exists {
		http.Error(w, "Batch not found", http.StatusNotFound)
		return
	}

	now := time.Now()
	if now.Before(batch.ValidFrom) {
		http.Error(w, "Coupon batch is not yet valid", http.StatusBadRequest)
		return
	}

	if now.After(batch.ValidTo) {
		http.Error(w, "Coupon batch has expired", http.StatusBadRequest)
		return
	}

	if batch.ClaimedQuantity >= batch.TotalQuantity {
		http.Error(w, "No more coupons available in this batch", http.StatusBadRequest)
		return
	}

	userClaims := h.storage.GetUserClaims(req.UserID)
	userBatchClaims := 0
	for _, claim := range userClaims {
		if claim.BatchID == req.BatchID {
			userBatchClaims++
		}
	}

	if userBatchClaims >= batch.MaxPerUser {
		http.Error(w, "User has already claimed maximum allowed coupons for this batch", http.StatusBadRequest)
		return
	}

	claim := &CouponClaim{
		ID:         GenerateID(),
		BatchID:    req.BatchID,
		UserID:     req.UserID,
		CouponCode: GenerateCouponCode(),
		ClaimedAt:  time.Now(),
		IsRedeemed: false,
		RedeemedAt: nil,
	}

	if err := h.storage.CreateClaim(claim); err != nil {
		http.Error(w, "Failed to create claim", http.StatusInternalServerError)
		return
	}

	batch.ClaimedQuantity++
	if err := h.storage.UpdateBatch(batch); err != nil {
		http.Error(w, "Failed to update batch", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(claim)
}

func (h *ClaimHandler) GetUserCoupons(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	showUsed := r.URL.Query().Get("show_used") == "true"
	showExpired := r.URL.Query().Get("show_expired") == "true"

	userClaims := h.storage.GetUserClaims(userID)
	availableClaims := make([]*CouponClaim, 0)
	now := time.Now()

	for _, claim := range userClaims {
		batch, exists := h.storage.GetBatch(claim.BatchID)
		if !exists {
			continue
		}

		if claim.IsRedeemed && !showUsed {
			continue
		}

		isExpired := now.After(batch.ValidTo)
		if isExpired && !showExpired {
			continue
		}

		availableClaims = append(availableClaims, claim)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(availableClaims)
}

func (h *ClaimHandler) ReturnCoupon(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ReturnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.BatchID == "" || req.UserID == "" || req.CouponCode == "" {
		http.Error(w, "Batch ID, User ID, and Coupon Code are required", http.StatusBadRequest)
		return
	}

	h.claimMu.Lock()
	defer h.claimMu.Unlock()

	claim, exists := h.storage.GetClaimByCode(req.CouponCode)
	if !exists {
		http.Error(w, "Coupon not found", http.StatusNotFound)
		return
	}

	if claim.BatchID != req.BatchID {
		http.Error(w, "Coupon does not belong to the specified batch", http.StatusBadRequest)
		return
	}

	if claim.UserID != req.UserID {
		http.Error(w, "Coupon does not belong to the specified user", http.StatusBadRequest)
		return
	}

	if claim.IsRedeemed {
		http.Error(w, "Cannot return a redeemed coupon", http.StatusBadRequest)
		return
	}

	batch, exists := h.storage.GetBatch(req.BatchID)
	if !exists {
		http.Error(w, "Batch not found", http.StatusNotFound)
		return
	}

	if err := h.storage.DeleteClaim(claim); err != nil {
		http.Error(w, "Failed to delete claim", http.StatusInternalServerError)
		return
	}

	batch.ClaimedQuantity--
	if err := h.storage.UpdateBatch(batch); err != nil {
		http.Error(w, "Failed to update batch", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "Coupon returned successfully",
		"coupon_code": req.CouponCode,
	})
}
