package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type CreateBatchRequest struct {
	Name            string  `json:"name"`
	DiscountAmount  float64 `json:"discount_amount"`
	ThresholdAmount float64 `json:"threshold_amount"`
	TotalQuantity   int     `json:"total_quantity"`
	ValidFrom       string  `json:"valid_from"`
	ValidTo         string  `json:"valid_to"`
	MaxPerUser      int     `json:"max_per_user"`
}

type UpdateBatchRequest struct {
	Name            *string  `json:"name,omitempty"`
	DiscountAmount  *float64 `json:"discount_amount,omitempty"`
	ThresholdAmount *float64 `json:"threshold_amount,omitempty"`
	ValidFrom       *string  `json:"valid_from,omitempty"`
	ValidTo         *string  `json:"valid_to,omitempty"`
	MaxPerUser      *int     `json:"max_per_user,omitempty"`
}

type BatchProgressResponse struct {
	BatchID          string `json:"batch_id"`
	BatchName        string `json:"batch_name"`
	TotalQuantity    int    `json:"total_quantity"`
	ClaimedQuantity  int    `json:"claimed_quantity"`
	RedeemedQuantity int    `json:"redeemed_quantity"`
	RemainingQuantity int    `json:"remaining_quantity"`
}

type BatchHandler struct {
	storage *Storage
}

func NewBatchHandler(storage *Storage) *BatchHandler {
	return &BatchHandler{storage: storage}
}

func (h *BatchHandler) CreateBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Coupon name cannot be empty", http.StatusBadRequest)
		return
	}

	if req.DiscountAmount <= 0 {
		http.Error(w, "Discount amount must be greater than zero", http.StatusBadRequest)
		return
	}

	if _, exists := h.storage.GetBatchByName(req.Name); exists {
		http.Error(w, "Coupon name already exists", http.StatusBadRequest)
		return
	}

	validFrom, err := time.Parse(time.RFC3339, req.ValidFrom)
	if err != nil {
		http.Error(w, "Invalid valid_from format, use RFC3339", http.StatusBadRequest)
		return
	}

	validTo, err := time.Parse(time.RFC3339, req.ValidTo)
	if err != nil {
		http.Error(w, "Invalid valid_to format, use RFC3339", http.StatusBadRequest)
		return
	}

	validTo = time.Date(validTo.Year(), validTo.Month(), validTo.Day(), 23, 59, 59, 0, validTo.Location())

	if validTo.Before(validFrom) {
		http.Error(w, "valid_to must be after valid_from", http.StatusBadRequest)
		return
	}

	batch := &CouponBatch{
		ID:               GenerateID(),
		Name:             req.Name,
		DiscountAmount:   req.DiscountAmount,
		ThresholdAmount:  req.ThresholdAmount,
		TotalQuantity:    req.TotalQuantity,
		ClaimedQuantity:  0,
		RedeemedQuantity: 0,
		ValidFrom:        validFrom,
		ValidTo:          validTo,
		MaxPerUser:       req.MaxPerUser,
		CreatedAt:        time.Now(),
	}

	if err := h.storage.CreateBatch(batch); err != nil {
		http.Error(w, "Failed to create batch", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(batch)
}

func (h *BatchHandler) UpdateBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	batchID := r.URL.Query().Get("id")
	if batchID == "" {
		http.Error(w, "Batch ID is required", http.StatusBadRequest)
		return
	}

	batch, exists := h.storage.GetBatch(batchID)
	if !exists {
		http.Error(w, "Batch not found", http.StatusNotFound)
		return
	}

	if batch.ClaimedQuantity >= batch.TotalQuantity {
		http.Error(w, "Cannot update batch: all coupons have been claimed", http.StatusBadRequest)
		return
	}

	var req UpdateBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name != nil {
		if *req.Name == "" {
			http.Error(w, "Coupon name cannot be empty", http.StatusBadRequest)
			return
		}
		if existing, exists := h.storage.GetBatchByName(*req.Name); exists && existing.ID != batchID {
			http.Error(w, "Coupon name already exists", http.StatusBadRequest)
			return
		}
		batch.Name = *req.Name
	}

	if req.DiscountAmount != nil {
		if *req.DiscountAmount <= 0 {
			http.Error(w, "Discount amount must be greater than zero", http.StatusBadRequest)
			return
		}
		batch.DiscountAmount = *req.DiscountAmount
	}

	if req.ThresholdAmount != nil {
		batch.ThresholdAmount = *req.ThresholdAmount
	}

	if req.ValidFrom != nil {
		validFrom, err := time.Parse(time.RFC3339, *req.ValidFrom)
		if err != nil {
			http.Error(w, "Invalid valid_from format, use RFC3339", http.StatusBadRequest)
			return
		}
		batch.ValidFrom = validFrom
	}

	if req.ValidTo != nil {
		validTo, err := time.Parse(time.RFC3339, *req.ValidTo)
		if err != nil {
			http.Error(w, "Invalid valid_to format, use RFC3339", http.StatusBadRequest)
			return
		}
		validTo = time.Date(validTo.Year(), validTo.Month(), validTo.Day(), 23, 59, 59, 0, validTo.Location())
		if validTo.Before(batch.ValidFrom) {
			http.Error(w, "valid_to must be after valid_from", http.StatusBadRequest)
			return
		}
		batch.ValidTo = validTo
	}

	if req.MaxPerUser != nil {
		batch.MaxPerUser = *req.MaxPerUser
	}

	if err := h.storage.UpdateBatch(batch); err != nil {
		http.Error(w, "Failed to update batch", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(batch)
}

func (h *BatchHandler) GetBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	batchID := r.URL.Query().Get("id")
	if batchID == "" {
		http.Error(w, "Batch ID is required", http.StatusBadRequest)
		return
	}

	batch, exists := h.storage.GetBatch(batchID)
	if !exists {
		http.Error(w, "Batch not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(batch)
}

func (h *BatchHandler) ListBatches(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	batches := h.storage.GetAllBatches()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(batches)
}

func (h *BatchHandler) GetBatchProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	batchID := r.URL.Query().Get("id")
	if batchID == "" {
		http.Error(w, "Batch ID is required", http.StatusBadRequest)
		return
	}

	batch, exists := h.storage.GetBatch(batchID)
	if !exists {
		http.Error(w, "Batch not found", http.StatusNotFound)
		return
	}

	response := BatchProgressResponse{
		BatchID:          batch.ID,
		BatchName:        batch.Name,
		TotalQuantity:    batch.TotalQuantity,
		ClaimedQuantity:  batch.ClaimedQuantity,
		RedeemedQuantity: batch.RedeemedQuantity,
		RemainingQuantity: batch.TotalQuantity - batch.ClaimedQuantity,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
