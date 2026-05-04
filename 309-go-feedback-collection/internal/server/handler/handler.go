package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"feedback-system/internal/server/model"
	"feedback-system/internal/server/store"
	"feedback-system/pkg/common"
)

type Handler struct {
	store store.Store
}

func NewHandler(s store.Store) *Handler {
	return &Handler{store: s}
}

type jsonError struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func writeJSONError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(jsonError{
		Success: false,
		Message: message,
	})
}

func writeJSONSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	var req common.SubmitFeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	if req.UserID == "" {
		writeJSONError(w, http.StatusBadRequest, "用户标识不能为空")
		return
	}

	if !common.IsValidFeedbackType(req.Type) {
		writeJSONError(w, http.StatusBadRequest, "反馈类型无效，只能是：suggestion、bug、complaint")
		return
	}

	rating, err := common.ValidateRating(req.Rating)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := common.ValidateDescription(req.Description); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	today := store.GetTodayDateString()
	currentCount, err := h.store.GetUserDailyCount(req.UserID, today)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "系统错误")
		return
	}

	if currentCount >= common.MaxDailyFeedbacks {
		writeJSONError(w, http.StatusTooManyRequests,
			fmt.Sprintf("今日提交反馈已达上限%d条", common.MaxDailyFeedbacks))
		return
	}

	feedback := &model.Feedback{
		ID:          uuid.New().String(),
		UserID:      req.UserID,
		Type:        common.FeedbackType(strings.ToLower(req.Type)),
		Rating:      rating,
		Description: req.Description,
		Status:      common.StatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.store.CreateFeedback(feedback); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "保存反馈失败")
		return
	}

	if err := h.store.IncrementUserDailyCount(req.UserID, today); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "系统错误")
		return
	}

	writeJSONSuccess(w, common.SubmitFeedbackResponse{
		FeedbackID: feedback.ID,
		Success:    true,
		Message:    "反馈提交成功",
	})
}

func (h *Handler) ListFeedbacks(w http.ResponseWriter, r *http.Request) {
	page := 1
	pageSize := common.DefaultPageSize
	var filterType common.FeedbackType
	var filterStatus common.FeedbackStatus

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			page = parsed
		}
	}

	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil {
			pageSize = parsed
		}
	}

	if t := r.URL.Query().Get("type"); t != "" {
		if common.IsValidFeedbackType(t) {
			filterType = common.FeedbackType(strings.ToLower(t))
		}
	}

	if s := r.URL.Query().Get("status"); s != "" {
		if common.IsValidFeedbackStatus(s) {
			filterStatus = common.FeedbackStatus(strings.ToLower(s))
		}
	}

	page, pageSize = common.ValidatePageParams(page, pageSize)

	feedbacks, total, err := h.store.ListFeedbacks(page, pageSize, filterType, filterStatus)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "查询失败")
		return
	}

	items := make([]common.FeedbackListItem, 0, len(feedbacks))
	for _, f := range feedbacks {
		items = append(items, f.ToListItem())
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPages == 0 {
		totalPages = 1
	}

	writeJSONSuccess(w, common.ListFeedbackResponse{
		Success:    true,
		Feedbacks:  items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

func (h *Handler) GetFeedback(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "反馈ID不能为空")
		return
	}

	feedback, err := h.store.GetFeedback(id)
	if err != nil {
		if errors.Is(err, store.ErrFeedbackNotFound) {
			writeJSONError(w, http.StatusNotFound, "反馈不存在")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "查询失败")
		return
	}

	item := feedback.ToListItem()
	writeJSONSuccess(w, common.GetFeedbackResponse{
		Success:  true,
		Feedback: &item,
	})
}

func (h *Handler) AddNote(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "反馈ID不能为空")
		return
	}

	var req common.AddNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	if err := common.ValidateNote(req.Note); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	feedback, err := h.store.GetFeedback(id)
	if err != nil {
		if errors.Is(err, store.ErrFeedbackNotFound) {
			writeJSONError(w, http.StatusNotFound, "反馈不存在")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "查询失败")
		return
	}

	if feedback.InternalNote != "" {
		feedback.InternalNote = feedback.InternalNote + "\n" + req.Note
	} else {
		feedback.InternalNote = req.Note
	}

	if err := h.store.UpdateFeedback(feedback); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "保存备注失败")
		return
	}

	writeJSONSuccess(w, common.AddNoteResponse{
		Success: true,
		Message: "备注添加成功",
	})
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "反馈ID不能为空")
		return
	}

	var req common.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	if !common.IsValidFeedbackStatus(req.Status) {
		writeJSONError(w, http.StatusBadRequest, "状态无效，只能是：pending、in_process、closed")
		return
	}

	feedback, err := h.store.GetFeedback(id)
	if err != nil {
		if errors.Is(err, store.ErrFeedbackNotFound) {
			writeJSONError(w, http.StatusNotFound, "反馈不存在")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "查询失败")
		return
	}

	newStatus := common.FeedbackStatus(strings.ToLower(req.Status))

	if newStatus == common.StatusClosed && feedback.Type == common.FeedbackTypeBug {
		if strings.TrimSpace(req.ProcessNote) == "" {
			writeJSONError(w, http.StatusBadRequest, "Bug类型反馈关闭时必须填写处理说明")
			return
		}
	}

	feedback.Status = newStatus
	if req.ProcessNote != "" {
		feedback.ProcessNote = req.ProcessNote
	}

	if err := h.store.UpdateFeedback(feedback); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "更新状态失败")
		return
	}

	writeJSONSuccess(w, common.UpdateStatusResponse{
		Success: true,
		Message: "状态更新成功",
	})
}

func (h *Handler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	byType, totalCount, avgRating, err := h.store.GetStatistics()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "查询统计失败")
		return
	}

	var resp common.StatisticsResponse
	resp.Success = true
	resp.Data.ByType = byType
	resp.Data.Overall.TotalCount = totalCount
	resp.Data.Overall.AverageRating = avgRating

	writeJSONSuccess(w, resp)
}
