package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SubmitRequest struct {
	UserID      string `json:"user_id"`
	Type        string `json:"type"`
	Rating      interface{} `json:"rating"`
	Description string `json:"description"`
}

type SubmissionHandler struct {
	storage *Storage
}

func NewSubmissionHandler(storage *Storage) *SubmissionHandler {
	return &SubmissionHandler{storage: storage}
}

func (h *SubmissionHandler) validateRating(rating interface{}) (int, error) {
	var ratingInt int

	switch v := rating.(type) {
	case float64:
		if v != float64(int(v)) {
			return 0, errors.New("评分必须是整数")
		}
		ratingInt = int(v)
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return 0, errors.New("评分格式无效")
		}
		ratingInt = parsed
	default:
		return 0, errors.New("评分格式无效")
	}

	if ratingInt < MinRating || ratingInt > MaxRating {
		return 0, fmt.Errorf("评分必须在%d到%d之间", MinRating, MaxRating)
	}

	return ratingInt, nil
}

func (h *SubmissionHandler) validateType(feedbackType string) error {
	if !ValidFeedbackTypes[feedbackType] {
		return errors.New("无效的反馈类型，只能选择：功能建议、Bug报告、体验投诉")
	}
	return nil
}

func (h *SubmissionHandler) validateDescription(description string) error {
	if len(description) == 0 {
		return errors.New("反馈描述不能为空")
	}
	if len([]rune(description)) > MaxDescriptionLength {
		return fmt.Errorf("反馈描述不能超过%d个字符", MaxDescriptionLength)
	}
	return nil
}

func (h *SubmissionHandler) validateUserID(userID string) error {
	if len(userID) == 0 {
		return errors.New("用户标识不能为空")
	}
	return nil
}

func (h *SubmissionHandler) checkDailyLimit(userID string) error {
	count := h.storage.GetUserDailyCount(userID, time.Now())
	if count >= MaxDailySubmissions {
		return fmt.Errorf("您今天已经提交了%d条反馈，达到每日上限", MaxDailySubmissions)
	}
	return nil
}

func (h *SubmissionHandler) HandleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "请求体解析失败"}`, http.StatusBadRequest)
		return
	}

	if err := h.validateUserID(req.UserID); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	if err := h.validateType(req.Type); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	rating, err := h.validateRating(req.Rating)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	if err := h.validateDescription(req.Description); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	if err := h.checkDailyLimit(req.UserID); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusTooManyRequests)
		return
	}

	now := time.Now()
	feedback := Feedback{
		ID:          uuid.New().String(),
		UserID:      req.UserID,
		Type:        req.Type,
		Rating:      rating,
		Description: req.Description,
		Status:      StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.storage.AddFeedback(feedback); err != nil {
		http.Error(w, `{"error": "保存反馈失败"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"id":      feedback.ID,
		"message": "反馈提交成功",
	}

	json.NewEncoder(w).Encode(response)
}
