package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const (
	referralCodeLength = 6
	referralCodeChars  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (app *App) handleReferralCode(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		app.handleGenerateReferralCode(w, r)
	case http.MethodGet:
		app.handleGetReferralCode(w, r)
	default:
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (app *App) handleGenerateReferralCode(w http.ResponseWriter, r *http.Request) {
	var req ReferralCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserID == "" {
		respondWithError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	user := app.store.GetUserByID(req.UserID)
	if user == nil {
		user = app.store.CreateUser(req.UserID)
	}

	if user.ReferralCode != "" {
		respondWithJSON(w, http.StatusOK, ReferralCodeResponse{
			UserID:       user.ID,
			ReferralCode: user.ReferralCode,
		})
		return
	}

	code, err := app.generateUniqueReferralCode()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate referral code")
		return
	}

	app.store.SetReferralCode(req.UserID, code)

	respondWithJSON(w, http.StatusOK, ReferralCodeResponse{
		UserID:       req.UserID,
		ReferralCode: code,
	})
}

func (app *App) handleGetReferralCode(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		respondWithError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	user := app.store.GetUserByID(userID)
	if user == nil {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	if user.ReferralCode == "" {
		respondWithError(w, http.StatusNotFound, "Referral code not found for this user")
		return
	}

	respondWithJSON(w, http.StatusOK, ReferralCodeResponse{
		UserID:       user.ID,
		ReferralCode: user.ReferralCode,
	})
}

func (app *App) generateUniqueReferralCode() (string, error) {
	maxAttempts := 100
	for i := 0; i < maxAttempts; i++ {
		code := generateRandomCode()
		if !app.store.ReferralCodeExists(code) {
			return code, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique referral code after %d attempts", maxAttempts)
}

func generateRandomCode() string {
	code := make([]byte, referralCodeLength)
	for i := range code {
		code[i] = referralCodeChars[rand.Intn(len(referralCodeChars))]
	}
	return string(code)
}

func normalizeReferralCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}
