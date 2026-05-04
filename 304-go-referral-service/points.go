package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (app *App) handleFirstOrderPaid(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req FirstOrderPaidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserID == "" || req.OrderID == "" {
		respondWithError(w, http.StatusBadRequest, "User ID and order ID are required")
		return
	}

	relation := app.store.GetReferralRelation(req.UserID)
	if relation == nil {
		respondWithError(w, http.StatusNotFound, "No referral relation found for this user")
		return
	}

	if relation.FirstOrderPaid {
		respondWithError(w, http.StatusConflict, "First order has already been processed")
		return
	}

	rewardPoints := app.store.GetRewardPointsPerFirstOrder()
	
	relation.FirstOrderID = req.OrderID
	relation.FirstOrderPaid = true
	relation.RewardPoints = rewardPoints
	
	app.store.UpdateReferralRelation(relation)

	description := fmt.Sprintf("Referral reward for first order %s by user %s", req.OrderID, req.UserID)
	app.store.AddPoints(relation.ReferrerID, rewardPoints, TransactionTypeReward, description, req.OrderID)

	respondWithJSON(w, http.StatusOK, FirstOrderPaidResponse{
		Success:      true,
		Message:      "Reward points issued successfully",
		RewardPoints: rewardPoints,
	})
}

func (app *App) handleOrderRefund(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req OrderRefundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserID == "" || req.OrderID == "" {
		respondWithError(w, http.StatusBadRequest, "User ID and order ID are required")
		return
	}

	relation := app.store.GetReferralRelation(req.UserID)
	if relation == nil {
		respondWithError(w, http.StatusNotFound, "No referral relation found for this user")
		return
	}

	if !relation.FirstOrderPaid || relation.FirstOrderID != req.OrderID {
		respondWithError(w, http.StatusNotFound, "No rewarded first order found for this refund")
		return
	}

	if relation.RewardPoints == 0 {
		respondWithError(w, http.StatusConflict, "Points have already been refunded")
		return
	}

	pointsToRefund := relation.RewardPoints
	relation.RewardPoints = 0
	relation.FirstOrderPaid = false
	app.store.UpdateReferralRelation(relation)

	currentPoints := app.store.GetUserPoints(relation.ReferrerID)
	actualDeduction := pointsToRefund
	if currentPoints < pointsToRefund {
		actualDeduction = currentPoints
	}

	description := fmt.Sprintf("Points refund due to order %s refund by user %s", req.OrderID, req.UserID)
	app.store.SubtractPoints(relation.ReferrerID, actualDeduction, TransactionTypeRefund, description, req.OrderID)

	respondWithJSON(w, http.StatusOK, OrderRefundResponse{
		Success:        true,
		Message:        "Points refunded successfully",
		PointsDeducted: actualDeduction,
	})
}

func (app *App) handleAdminConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		app.handleGetAdminConfig(w, r)
	case http.MethodPut:
		app.handleUpdateAdminConfig(w, r)
	default:
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (app *App) handleGetAdminConfig(w http.ResponseWriter, r *http.Request) {
	config := AdminConfigResponse{
		RewardPointsPerFirstOrder: app.store.GetRewardPointsPerFirstOrder(),
	}
	respondWithJSON(w, http.StatusOK, config)
}

func (app *App) handleUpdateAdminConfig(w http.ResponseWriter, r *http.Request) {
	var req AdminConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.RewardPointsPerFirstOrder < 0 {
		respondWithError(w, http.StatusBadRequest, "Reward points cannot be negative")
		return
	}

	app.store.SetRewardPointsPerFirstOrder(req.RewardPointsPerFirstOrder)

	config := AdminConfigResponse{
		RewardPointsPerFirstOrder: app.store.GetRewardPointsPerFirstOrder(),
	}
	respondWithJSON(w, http.StatusOK, config)
}

func (app *App) handleAdminStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	allRelations := app.store.GetAllReferralRelations()
	
	totalReferrals := len(allRelations)
	completedFirstOrder := 0
	totalPointsDistributed := int64(0)
	
	for _, relation := range allRelations {
		if relation.FirstOrderPaid {
			completedFirstOrder++
		}
		totalPointsDistributed += relation.RewardPoints
	}
	
	var completionRate float64
	if totalReferrals > 0 {
		completionRate = float64(completedFirstOrder) / float64(totalReferrals) * 100
	}

	respondWithJSON(w, http.StatusOK, AdminStatsResponse{
		TotalReferrals:           totalReferrals,
		FirstOrderCompletionRate: completionRate,
		TotalPointsDistributed:   totalPointsDistributed,
	})
}
