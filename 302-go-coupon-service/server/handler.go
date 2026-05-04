package main

import (
	"coupon-service/pkg/api"
	"encoding/json"
	"net/http"
	"strings"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/batches", h.handleBatches)
	mux.HandleFunc("/api/batches/", h.handleBatchByID)
	mux.HandleFunc("/api/claims", h.handleClaims)
	mux.HandleFunc("/api/redemptions", h.handleRedemptions)
	mux.HandleFunc("/api/users/", h.handleUsers)
	mux.HandleFunc("/api/returns", h.handleReturns)
}

func (h *Handler) handleBatches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodPost:
		var req api.CreateBatchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"success": false, "message": "Invalid request body"}`, http.StatusBadRequest)
			return
		}

		resp, err := h.service.CreateBatch(&req)
		if err != nil {
			http.Error(w, `{"success": false, "message": "Internal server error"}`, http.StatusInternalServerError)
			return
		}

		if !resp.Success {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusCreated)
		}
		json.NewEncoder(w).Encode(resp)

	default:
		http.Error(w, `{"success": false, "message": "Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleBatchByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/api/batches/")
	parts := strings.Split(path, "/")

	if len(parts) < 1 {
		http.Error(w, `{"success": false, "message": "Invalid path"}`, http.StatusBadRequest)
		return
	}

	batchID := parts[0]

	switch {
	case len(parts) == 2 && parts[1] == "stats" && r.Method == http.MethodGet:
		req := api.GetBatchStatsRequest{BatchID: batchID}
		resp, err := h.service.GetBatchStats(&req)
		if err != nil {
			http.Error(w, `{"success": false, "message": "Internal server error"}`, http.StatusInternalServerError)
			return
		}

		if !resp.Success {
			w.WriteHeader(http.StatusNotFound)
		}
		json.NewEncoder(w).Encode(resp)

	case r.Method == http.MethodPut:
		var req api.UpdateBatchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"success": false, "message": "Invalid request body"}`, http.StatusBadRequest)
			return
		}
		req.BatchID = batchID

		resp, err := h.service.UpdateBatch(&req)
		if err != nil {
			http.Error(w, `{"success": false, "message": "Internal server error"}`, http.StatusInternalServerError)
			return
		}

		if !resp.Success {
			w.WriteHeader(http.StatusBadRequest)
		}
		json.NewEncoder(w).Encode(resp)

	default:
		http.Error(w, `{"success": false, "message": "Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleClaims(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodPost:
		var req api.ClaimCouponRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"success": false, "message": "Invalid request body"}`, http.StatusBadRequest)
			return
		}

		resp, err := h.service.ClaimCoupon(&req)
		if err != nil {
			http.Error(w, `{"success": false, "message": "Internal server error"}`, http.StatusInternalServerError)
			return
		}

		if !resp.Success {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusCreated)
		}
		json.NewEncoder(w).Encode(resp)

	default:
		http.Error(w, `{"success": false, "message": "Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleRedemptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodPost:
		var req api.RedeemCouponRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"success": false, "message": "Invalid request body"}`, http.StatusBadRequest)
			return
		}

		resp, err := h.service.RedeemCoupon(&req)
		if err != nil {
			http.Error(w, `{"success": false, "message": "Internal server error"}`, http.StatusInternalServerError)
			return
		}

		if !resp.Success {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusCreated)
		}
		json.NewEncoder(w).Encode(resp)

	default:
		http.Error(w, `{"success": false, "message": "Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/api/users/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 || parts[1] != "coupons" || r.Method != http.MethodGet {
		http.Error(w, `{"success": false, "message": "Invalid path or method"}`, http.StatusBadRequest)
		return
	}

	userID := parts[0]
	req := api.GetUserCouponsRequest{UserID: userID}

	resp, err := h.service.GetUserCoupons(&req)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) handleReturns(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodPost:
		var req api.ReturnCouponRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"success": false, "message": "Invalid request body"}`, http.StatusBadRequest)
			return
		}

		resp, err := h.service.ReturnCoupon(&req)
		if err != nil {
			http.Error(w, `{"success": false, "message": "Internal server error"}`, http.StatusInternalServerError)
			return
		}

		if !resp.Success {
			w.WriteHeader(http.StatusBadRequest)
		}
		json.NewEncoder(w).Encode(resp)

	default:
		http.Error(w, `{"success": false, "message": "Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}
