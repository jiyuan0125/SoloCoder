package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type QueryHandler struct {
	storage *Storage
}

func NewQueryHandler(storage *Storage) *QueryHandler {
	return &QueryHandler{storage: storage}
}

func (h *QueryHandler) parsePageParams(r *http.Request) (page, pageSize int, err error) {
	pageStr := strings.TrimSpace(r.URL.Query().Get("page"))
	if pageStr == "" {
		page = 1
	} else {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			return 0, 0, fmt.Errorf("页码格式无效")
		}
		if page <= 0 {
			return 0, 0, fmt.Errorf("页码不能为负数或零")
		}
	}

	pageSizeStr := strings.TrimSpace(r.URL.Query().Get("page_size"))
	if pageSizeStr == "" {
		pageSize = DefaultPageSize
	} else {
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil {
			return 0, 0, fmt.Errorf("每页数量格式无效")
		}
		if pageSize <= 0 {
			return 0, 0, fmt.Errorf("每页数量不能为负数或零")
		}
		if pageSize > MaxPageSize {
			pageSize = MaxPageSize
		}
	}

	return page, pageSize, nil
}

func (h *QueryHandler) filterFeedbacks(feedbacks []Feedback, feedbackType, status string) []Feedback {
	var result []Feedback

	for _, f := range feedbacks {
		matchType := feedbackType == "" || f.Type == feedbackType
		matchStatus := status == "" || f.Status == status

		if matchType && matchStatus {
			result = append(result, f)
		}
	}

	return result
}

func (h *QueryHandler) sortFeedbacksByTime(feedbacks []Feedback, ascending bool) {
	sort.Slice(feedbacks, func(i, j int) bool {
		if ascending {
			return feedbacks[i].CreatedAt.Before(feedbacks[j].CreatedAt)
		}
		return feedbacks[i].CreatedAt.After(feedbacks[j].CreatedAt)
	})
}

func (h *QueryHandler) paginateFeedbacks(feedbacks []Feedback, page, pageSize int) ([]Feedback, int, int) {
	totalCount := len(feedbacks)
	if totalCount == 0 {
		return []Feedback{}, 0, 0
	}

	totalPages := (totalCount + pageSize - 1) / pageSize

	if page > totalPages {
		return []Feedback{}, totalCount, totalPages
	}

	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize

	if endIndex > totalCount {
		endIndex = totalCount
	}

	return feedbacks[startIndex:endIndex], totalCount, totalPages
}

func (h *QueryHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	page, pageSize, err := h.parsePageParams(r)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	feedbackType := strings.TrimSpace(r.URL.Query().Get("type"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))

	if feedbackType != "" && !ValidFeedbackTypes[feedbackType] {
		http.Error(w, `{"error": "无效的反馈类型"}`, http.StatusBadRequest)
		return
	}

	if status != "" && !ValidStatuses[status] {
		http.Error(w, `{"error": "无效的状态值"}`, http.StatusBadRequest)
		return
	}

	feedbacks := h.storage.GetAllFeedbacks()
	filtered := h.filterFeedbacks(feedbacks, feedbackType, status)
	h.sortFeedbacksByTime(filtered, false)
	data, totalCount, totalPages := h.paginateFeedbacks(filtered, page, pageSize)

	response := PaginatedResponse{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *QueryHandler) HandleStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	stats := h.storage.GetStatistics()
	json.NewEncoder(w).Encode(stats)
}

func (h *QueryHandler) HandleGetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/api/feedbacks/")
	id := strings.TrimSpace(path)

	if id == "" {
		http.Error(w, `{"error": "反馈ID不能为空"}`, http.StatusBadRequest)
		return
	}

	feedback, exists := h.storage.GetFeedbackByID(id)
	if !exists {
		http.Error(w, `{"error": "反馈不存在"}`, http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(feedback)
}
