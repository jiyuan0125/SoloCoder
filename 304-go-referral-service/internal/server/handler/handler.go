package handler

import (
	"encoding/json"
	"net/http"

	"referral-service/internal/server/service"
	"referral-service/pkg/api"
)

type Handler struct {
	service *service.Service
}

func NewHandler(s *service.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) GenerateReferralCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.GenerateReferralCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.UserID == "" {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Error: "user_id is required"})
		return
	}

	code, err := h.service.GenerateReferralCode(req.UserID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, api.GenerateReferralCodeResponse{ReferralCode: code})
}

func (h *Handler) BindReferral(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.BindReferralRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.NewUserID == "" || req.ReferralCode == "" {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Error: "new_user_id and referral_code are required"})
		return
	}

	err := h.service.BindReferral(req.NewUserID, req.ReferralCode)
	if err != nil {
		switch err {
		case service.ErrInvalidReferralCode:
			writeJSON(w, http.StatusNotFound, api.BindReferralResponse{
				Success: false,
				Message: "invalid referral code",
			})
		case service.ErrSelfReferral:
			writeJSON(w, http.StatusBadRequest, api.BindReferralResponse{
				Success: false,
				Message: "cannot use your own referral code",
			})
		case service.ErrAlreadyBound:
			writeJSON(w, http.StatusConflict, api.BindReferralResponse{
				Success: false,
				Message: "user already bound to a referrer",
			})
		default:
			writeJSON(w, http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
		}
		return
	}

	writeJSON(w, http.StatusOK, api.BindReferralResponse{
		Success: true,
		Message: "referral bound successfully",
	})
}

func (h *Handler) CompleteFirstOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.CompleteFirstOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.UserID == "" || req.OrderID == "" {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Error: "user_id and order_id are required"})
		return
	}

	points, err := h.service.CompleteFirstOrder(req.UserID, req.OrderID)
	if err != nil {
		switch err {
		case service.ErrNoReferral:
			writeJSON(w, http.StatusOK, api.CompleteFirstOrderResponse{
				Success: true,
				Message: "no referral relationship, no points awarded",
			})
		case service.ErrOrderAlreadyCompleted:
			writeJSON(w, http.StatusConflict, api.CompleteFirstOrderResponse{
				Success: false,
				Message: "first order already completed",
			})
		default:
			writeJSON(w, http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
		}
		return
	}

	writeJSON(w, http.StatusOK, api.CompleteFirstOrderResponse{
		Success:       true,
		Message:       "first order completed, points awarded",
		PointsAwarded: points,
	})
}

func (h *Handler) RefundFirstOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.RefundFirstOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.UserID == "" || req.OrderID == "" {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Error: "user_id and order_id are required"})
		return
	}

	points, err := h.service.RefundFirstOrder(req.UserID, req.OrderID)
	if err != nil {
		switch err {
		case service.ErrNoReferral:
			writeJSON(w, http.StatusNotFound, api.RefundFirstOrderResponse{
				Success: false,
				Message: "no referral relationship found",
			})
		case service.ErrInvalidOrder:
			writeJSON(w, http.StatusBadRequest, api.RefundFirstOrderResponse{
				Success: false,
				Message: "invalid order or order not completed",
			})
		default:
			writeJSON(w, http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
		}
		return
	}

	writeJSON(w, http.StatusOK, api.RefundFirstOrderResponse{
		Success:        true,
		Message:        "first order refunded, points deducted",
		PointsDeducted: points,
	})
}

func (h *Handler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.GetUserStatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.UserID == "" {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Error: "user_id is required"})
		return
	}

	code, totalReferrals, completedFirstOrders, totalPoints := h.service.GetUserStats(req.UserID)

	writeJSON(w, http.StatusOK, api.GetUserStatsResponse{
		UserID:               req.UserID,
		ReferralCode:         code,
		TotalReferrals:       totalReferrals,
		CompletedFirstOrders: completedFirstOrders,
		TotalPoints:          totalPoints,
	})
}

func (h *Handler) HandleRewardPoints(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetRewardPoints(w, r)
	case http.MethodPost:
		h.SetRewardPoints(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) GetRewardPoints(w http.ResponseWriter, r *http.Request) {
	points := h.service.GetRewardPoints()
	writeJSON(w, http.StatusOK, api.GetRewardPointsResponse{Points: points})
}

func (h *Handler) SetRewardPoints(w http.ResponseWriter, r *http.Request) {
	var req api.SetRewardPointsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Points < 0 {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Error: "points cannot be negative"})
		return
	}

	if err := h.service.SetRewardPoints(req.Points); err != nil {
		writeJSON(w, http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, api.SetRewardPointsResponse{
		Success: true,
		Message: "reward points updated",
	})
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	totalReferrals, completedFirstOrders, rate, totalPointsIssued := h.service.GetStats()

	writeJSON(w, http.StatusOK, api.GetStatsResponse{
		TotalReferrals:      totalReferrals,
		CompletedFirstOrders: completedFirstOrders,
		FirstOrderRate:      rate,
		TotalPointsIssued:   totalPointsIssued,
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
