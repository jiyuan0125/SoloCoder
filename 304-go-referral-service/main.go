package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type App struct {
	store *Store
}

func NewApp(store *Store) *App {
	return &App{store: store}
}

func (app *App) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/referral-code", app.handleReferralCode)
	mux.HandleFunc("/api/referral-bind", app.handleReferralBind)
	mux.HandleFunc("/api/referral-stats", app.handleReferralStats)
	mux.HandleFunc("/api/first-order-paid", app.handleFirstOrderPaid)
	mux.HandleFunc("/api/order-refund", app.handleOrderRefund)
	mux.HandleFunc("/api/admin/config", app.handleAdminConfig)
	mux.HandleFunc("/api/admin/stats", app.handleAdminStats)
}

func main() {
	store := NewStore("referral_data.json")
	app := NewApp(store)
	
	mux := http.NewServeMux()
	app.RegisterRoutes(mux)
	
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

type ReferralCodeRequest struct {
	UserID string `json:"user_id"`
}

type ReferralCodeResponse struct {
	UserID       string `json:"user_id"`
	ReferralCode string `json:"referral_code"`
}

type ReferralBindRequest struct {
	RefereeID    string `json:"referee_id"`
	ReferralCode string `json:"referral_code"`
}

type ReferralBindResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ReferralStatsRequest struct {
	UserID string `json:"user_id"`
}

type ReferralStatsResponse struct {
	UserID               string `json:"user_id"`
	TotalInvited         int    `json:"total_invited"`
	CompletedFirstOrder  int    `json:"completed_first_order"`
	TotalPointsEarned    int64  `json:"total_points_earned"`
	CurrentPoints        int64  `json:"current_points"`
}

type FirstOrderPaidRequest struct {
	UserID   string `json:"user_id"`
	OrderID  string `json:"order_id"`
}

type FirstOrderPaidResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	RewardPoints int64  `json:"reward_points"`
}

type OrderRefundRequest struct {
	UserID  string `json:"user_id"`
	OrderID string `json:"order_id"`
}

type OrderRefundResponse struct {
	Success         bool   `json:"success"`
	Message         string `json:"message"`
	PointsDeducted  int64  `json:"points_deducted"`
}

type AdminConfigRequest struct {
	RewardPointsPerFirstOrder int64 `json:"reward_points_per_first_order"`
}

type AdminConfigResponse struct {
	RewardPointsPerFirstOrder int64 `json:"reward_points_per_first_order"`
}

type AdminStatsResponse struct {
	TotalReferrals            int     `json:"total_referrals"`
	FirstOrderCompletionRate  float64 `json:"first_order_completion_rate"`
	TotalPointsDistributed    int64   `json:"total_points_distributed"`
}
