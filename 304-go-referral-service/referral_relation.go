package main

import (
	"encoding/json"
	"net/http"
)

func (app *App) handleReferralBind(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req ReferralBindRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.RefereeID == "" || req.ReferralCode == "" {
		respondWithError(w, http.StatusBadRequest, "Referee ID and referral code are required")
		return
	}

	normalizedCode := normalizeReferralCode(req.ReferralCode)

	existingRelation := app.store.GetReferralRelation(req.RefereeID)
	if existingRelation != nil {
		respondWithError(w, http.StatusConflict, "This user has already been bound to a referrer")
		return
	}

	referrerID, exists := app.store.GetUserIDByReferralCode(normalizedCode)
	if !exists {
		respondWithError(w, http.StatusNotFound, "Invalid referral code")
		return
	}

	if referrerID == req.RefereeID {
		respondWithError(w, http.StatusBadRequest, "Cannot bind to your own referral code")
		return
	}

	app.store.CreateReferralRelation(referrerID, req.RefereeID)

	respondWithJSON(w, http.StatusOK, ReferralBindResponse{
		Success: true,
		Message: "Referral relation created successfully",
	})
}

func (app *App) handleReferralStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		respondWithError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	referrals := app.store.GetReferralsByReferrer(userID)
	
	totalInvited := len(referrals)
	completedFirstOrder := 0
	totalPointsEarned := int64(0)
	
	for _, relation := range referrals {
		if relation.FirstOrderPaid {
			completedFirstOrder++
			totalPointsEarned += relation.RewardPoints
		}
	}
	
	currentPoints := app.store.GetUserPoints(userID)

	respondWithJSON(w, http.StatusOK, ReferralStatsResponse{
		UserID:              userID,
		TotalInvited:        totalInvited,
		CompletedFirstOrder: completedFirstOrder,
		TotalPointsEarned:   totalPointsEarned,
		CurrentPoints:       currentPoints,
	})
}
