package handlers

import (
	"encoding/json"
	"net/http"
	"recruit-flow/models"
	"recruit-flow/service"
	"strconv"
	"strings"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, err error) {
	code := service.GetServiceErrorCode(err)
	if code == 0 {
		code = 500
	}
	writeJSON(w, code, map[string]interface{}{
		"error": err.Error(),
	})
}

func CreateCandidateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name         string   `json:"name"`
		Email        string   `json:"email"`
		Phone        string   `json:"phone"`
		PositionID   int64    `json:"position_id"`
		Operator     string   `json:"operator"`
		Interviewers []string `json:"interviewers"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Name == "" {
		writeJSON(w, 400, map[string]string{"error": "name is required"})
		return
	}

	if req.PositionID == 0 {
		req.PositionID = 1
	}

	if req.Operator == "" {
		req.Operator = "anonymous"
	}

	candidate, err := service.CreateCandidate(req.Name, req.Email, req.Phone, req.PositionID, req.Operator, req.Interviewers)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, 201, candidate)
}

func GetCandidateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/candidates/")
	idStr = strings.Split(idStr, "/")[0]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid candidate id"})
		return
	}

	candidate, err := service.GetCandidate(id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, 200, candidate)
}

func AdvanceStageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/candidates/")
	idStr = strings.Replace(idStr, "/advance", "", 1)
	idStr = strings.Split(idStr, "/")[0]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid candidate id"})
		return
	}

	var req struct {
		Operator     string                 `json:"operator"`
		Score        *float64               `json:"score"`
		Owner        string                 `json:"owner"`
		ExtraData    map[string]interface{} `json:"extra_data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Operator == "" {
		req.Operator = "anonymous"
	}

	if req.ExtraData == nil {
		req.ExtraData = make(map[string]interface{})
	}
	if req.Score != nil {
		req.ExtraData["score"] = *req.Score
	}
	if req.Owner != "" {
		req.ExtraData["owner"] = req.Owner
	}

	nextStage, err := service.AdvanceStage(id, req.Operator, req.ExtraData)
	if err != nil {
		candidate, _ := service.GetCandidate(id)
		if candidate != nil {
			writeJSON(w, service.GetServiceErrorCode(err), map[string]interface{}{
				"error":        err.Error(),
				"currentStage": candidate.CurrentStage,
			})
			return
		}
		writeError(w, err)
		return
	}

	_ = service.ValidateData(id)

	writeJSON(w, 200, map[string]interface{}{
		"nextStage": nextStage,
		"success":   true,
	})
}

func RejectCandidateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/candidates/")
	idStr = strings.Replace(idStr, "/reject", "", 1)
	idStr = strings.Split(idStr, "/")[0]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid candidate id"})
		return
	}

	var req struct {
		Reason   string `json:"reason"`
		Operator string `json:"operator"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Reason == "" {
		writeJSON(w, 400, map[string]string{"error": "rejection reason is required"})
		return
	}

	if req.Operator == "" {
		req.Operator = "anonymous"
	}

	if err := service.RejectCandidate(id, req.Reason, req.Operator); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, 200, map[string]interface{}{
		"success": true,
		"message": "candidate rejected",
	})
}

func SubmitTechScoreHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/candidates/")
	idStr = strings.Replace(idStr, "/tech-score", "", 1)
	idStr = strings.Split(idStr, "/")[0]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid candidate id"})
		return
	}

	var req struct {
		Interviewer string  `json:"interviewer"`
		Score       float64 `json:"score"`
		Operator    string  `json:"operator"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Interviewer == "" {
		writeJSON(w, 400, map[string]string{"error": "interviewer is required"})
		return
	}

	if req.Operator == "" {
		req.Operator = "anonymous"
	}

	if err := service.SubmitTechInterviewScore(id, req.Interviewer, req.Score, req.Operator); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, 200, map[string]interface{}{
		"success": true,
	})
}

func GetHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/candidates/")
	idStr = strings.Replace(idStr, "/history", "", 1)
	idStr = strings.Split(idStr, "/")[0]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid candidate id"})
		return
	}

	history, err := service.GetCandidateHistory(id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, 200, history)
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func StageInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, 200, map[string]interface{}{
		"stages": models.StageOrder,
		"info": map[string]string{
			"resume_screening":  "简历筛选",
			"written_test":      "笔试",
			"tech_interview":    "技术面",
			"hr_interview":      "HR面",
			"offer":             "Offer",
			"onboarding":        "入职",
			"rejected":          "已淘汰",
			"offer_expired":     "Offer已过期",
		},
	})
}
