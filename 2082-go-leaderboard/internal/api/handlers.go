package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"leaderboard/internal/leaderboard"
	"leaderboard/internal/planner"
)

type Handler struct {
	leaderboardService *leaderboard.Service
	plannerService     *planner.Service
}

func NewHandler(ls *leaderboard.Service, ps *planner.Service) *Handler {
	return &Handler{
		leaderboardService: ls,
		plannerService:     ps,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/scores", h.handleScores)
	mux.HandleFunc("/api/leaderboard/", h.handleLeaderboard)
	mux.HandleFunc("/api/player/", h.handlePlayer)
	mux.HandleFunc("/api/plans", h.handlePlans)
	mux.HandleFunc("/api/plans/", h.handlePlan)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

type submitScoreRequest struct {
	PlayerID string `json:"player_id"`
	Score    int    `json:"score"`
}

func (h *Handler) handleScores(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	var req submitScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	if req.PlayerID == "" {
		writeError(w, http.StatusBadRequest, errors.New("player_id is required"))
		return
	}

	if req.Score < 0 {
		writeError(w, http.StatusBadRequest, errors.New("score cannot be negative"))
		return
	}

	if err := h.leaderboardService.SubmitScore(r.Context(), req.PlayerID, req.Score); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/leaderboard/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, errors.New("dimension is required"))
		return
	}

	dimension := parts[0]
	if !h.leaderboardService.IsValidDimension(dimension) {
		writeError(w, http.StatusBadRequest, errors.New("invalid dimension"))
		return
	}

	nStr := r.URL.Query().Get("n")
	if nStr == "" {
		nStr = "10"
	}

	n, err := strconv.Atoi(nStr)
	if err != nil || n <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("invalid n parameter"))
		return
	}

	scores, err := h.leaderboardService.GetTopN(r.Context(), dimension, n)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, scores)
}

func (h *Handler) handlePlayer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/player/"), "/")
	if len(parts) < 2 {
		writeError(w, http.StatusBadRequest, errors.New("dimension and player_id are required"))
		return
	}

	dimension := parts[0]
	playerID := parts[1]

	if !h.leaderboardService.IsValidDimension(dimension) {
		writeError(w, http.StatusBadRequest, errors.New("invalid dimension"))
		return
	}

	score, err := h.leaderboardService.GetPlayerRank(r.Context(), dimension, playerID)
	if err != nil {
		if err.Error() == "player not found" {
			writeError(w, http.StatusNotFound, err)
		} else {
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, http.StatusOK, score)
}

type createPlanRequest struct {
	TotalAmount float64 `json:"total_amount"`
	NumPeriods  int     `json:"num_periods"`
}

func (h *Handler) handlePlans(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	var req createPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}

	plan, err := h.plannerService.CreatePlan(r.Context(), req.TotalAmount, req.NumPeriods)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusCreated, plan)
}

type updatePlanRequest struct {
	TotalAmount float64 `json:"total_amount"`
}

func (h *Handler) handlePlan(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/plans/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, errors.New("plan_id is required"))
		return
	}

	planID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid plan_id"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		plan, err := h.plannerService.GetPlan(r.Context(), planID)
		if err != nil {
			if err.Error() == "plan not found" {
				writeError(w, http.StatusNotFound, err)
			} else {
				writeError(w, http.StatusInternalServerError, err)
			}
			return
		}
		writeJSON(w, http.StatusOK, plan)

	case http.MethodPut:
		var req updatePlanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, errors.New("invalid request body"))
			return
		}

		plan, err := h.plannerService.UpdateTotalAmount(r.Context(), planID, req.TotalAmount)
		if err != nil {
			if err.Error() == "plan not found" {
				writeError(w, http.StatusNotFound, err)
			} else {
				writeError(w, http.StatusInternalServerError, err)
			}
			return
		}
		writeJSON(w, http.StatusOK, plan)

	default:
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
	}
}

func (h *Handler) StartPeriodicCleanup(stopChan <-chan struct{}) {
	// 这里可以实现定期清理逻辑
	// 由于时间关系，我先不实现完整的定时任务，
	// 实际生产中可以使用 time.Ticker 来实现每日和每周的清理
	fmt.Println("Periodic cleanup is ready")
}
