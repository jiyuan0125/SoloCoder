package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"userbehavior/internal/shared"
)

type Handler struct {
	store       *Store
	analytics   *AnalyticsEngine
	qualityMon  *QualityMonitor
}

func NewHandler(store *Store, analytics *AnalyticsEngine, qualityMon *QualityMonitor) *Handler {
	return &Handler{
		store:      store,
		analytics:  analytics,
		qualityMon: qualityMon,
	}
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, shared.APIError{
		Code:    status,
		Message: message,
	})
}

func (h *Handler) TrackBehavior(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req shared.TrackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UserID == "" {
		h.respondError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	session, _, err := h.store.GetOrCreateSession(req.UserID, req.SessionID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to manage session")
		return
	}

	behavior := &shared.Behavior{
		ID:           fmt.Sprintf("beh-%d", time.Now().UnixNano()),
		UserID:       req.UserID,
		SessionID:    session.ID,
		Type:         req.Type,
		Timestamp:    time.Now(),
		IsFiltered:   false,
		FilterReason: "",
	}

	switch req.Type {
	case shared.BehaviorTypePageView:
		if req.URL == "" || req.PageKey == "" {
			h.respondError(w, http.StatusBadRequest, "url and page_key are required for page_view")
			return
		}
		behavior.PageView = &shared.PageViewData{
			URL:      req.URL,
			Duration: time.Duration(req.Duration) * time.Millisecond,
			PageKey:  req.PageKey,
		}

		if h.store.CheckDeduplication(req.UserID, req.PageKey, behavior.Timestamp, req.Type) {
			h.respondJSON(w, http.StatusOK, shared.TrackResponse{
				Success:   true,
				SessionID: session.ID,
				Message:   "behavior deduplicated",
			})
			return
		}

	case shared.BehaviorTypeButtonClick:
		if req.ButtonID == "" {
			h.respondError(w, http.StatusBadRequest, "button_id is required for button_click")
			return
		}
		behavior.ButtonClick = &shared.ButtonClickData{
			ButtonID: req.ButtonID,
		}

	case shared.BehaviorTypeFeatureUse:
		if req.FeatureName == "" {
			h.respondError(w, http.StatusBadRequest, "feature_name is required for feature_use")
			return
		}
		behavior.FeatureUse = &shared.FeatureUseData{
			FeatureName: req.FeatureName,
			Duration:    time.Duration(req.Duration) * time.Millisecond,
		}

	default:
		h.respondError(w, http.StatusBadRequest, "unknown behavior type")
		return
	}

	isValid, filterReason := h.qualityMon.ValidateAndFilter(behavior)
	if !isValid {
		behavior.IsFiltered = true
		behavior.FilterReason = filterReason
	}

	if err := h.store.TrackBehavior(behavior); err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to track behavior")
		return
	}

	message := "behavior tracked successfully"
	if behavior.IsFiltered {
		message = fmt.Sprintf("behavior filtered: %s", behavior.FilterReason)
	}

	h.respondJSON(w, http.StatusOK, shared.TrackResponse{
		Success:   true,
		SessionID: session.ID,
		Message:   message,
	})
}

func (h *Handler) FunnelAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req shared.FunnelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Steps) == 0 {
		h.respondError(w, http.StatusBadRequest, "no funnel steps provided")
		return
	}

	if len(req.Steps) > shared.MaxFunnelSteps {
		h.respondError(w, http.StatusBadRequest, fmt.Sprintf("maximum %d funnel steps allowed", shared.MaxFunnelSteps))
		return
	}

	result, err := h.analytics.AnalyzeFunnel(req.Steps)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to analyze funnel")
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

func (h *Handler) RetentionAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req shared.RetentionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.StartDate.IsZero() || req.EndDate.IsZero() {
		h.respondError(w, http.StatusBadRequest, "start_date and end_date are required")
		return
	}

	result, err := h.analytics.AnalyzeRetention(req.StartDate, req.EndDate)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to analyze retention")
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

func (h *Handler) PathAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req shared.PathAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.FromPage == "" || req.ToPage == "" {
		h.respondError(w, http.StatusBadRequest, "from_page and to_page are required")
		return
	}

	result, err := h.analytics.AnalyzePaths(req.FromPage, req.ToPage)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to analyze paths")
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

func (h *Handler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.respondError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	profile, err := h.analytics.GenerateUserProfile(userID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to generate user profile")
		return
	}

	if profile == nil {
		h.respondError(w, http.StatusNotFound, "user not found or no behavior data")
		return
	}

	h.respondJSON(w, http.StatusOK, profile)
}

func (h *Handler) GetRealtimeStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	stats, err := h.analytics.GetRealtimeStats()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to get realtime stats")
		return
	}

	h.respondJSON(w, http.StatusOK, stats)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}
