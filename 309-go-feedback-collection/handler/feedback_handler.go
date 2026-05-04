package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"feedback-collection/models"
	"feedback-collection/service"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func WriteJSONResponse(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func WriteErrorResponse(w http.ResponseWriter, code int, err error) {
	if err != nil {
		WriteJSONResponse(w, code, err.Error(), nil)
	} else {
		WriteJSONResponse(w, code, "method not allowed", nil)
	}
}

func DecodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func writeJSONResponse(w http.ResponseWriter, code int, message string, data interface{}) {
	WriteJSONResponse(w, code, message, data)
}

func writeErrorResponse(w http.ResponseWriter, code int, err error) {
	WriteErrorResponse(w, code, err)
}

func CreateFeedbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, nil)
		return
	}

	var req models.CreateFeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	feedback, err := service.CreateFeedback(&req)
	if err != nil {
		switch err {
		case service.ErrInvalidRating,
			service.ErrInvalidFeedbackType,
			service.ErrDescriptionTooLong,
			service.ErrUserIDRequired:
			writeErrorResponse(w, http.StatusBadRequest, err)
		case service.ErrDailyLimitExceeded:
			writeErrorResponse(w, http.StatusTooManyRequests, err)
		default:
			writeErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSONResponse(w, http.StatusCreated, "feedback created successfully", feedback)
}

func GetFeedbackListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorResponse(w, http.StatusMethodNotAllowed, nil)
		return
	}

	feedbackType := r.URL.Query().Get("type")
	status := r.URL.Query().Get("status")
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := 1
	if pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, service.ErrInvalidPage)
			return
		}
	}

	pageSize := 0
	if pageSizeStr != "" {
		var err error
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, service.ErrInvalidPageSize)
			return
		}
	}

	feedbacks, pagination, err := service.GetFeedbackList(feedbackType, status, page, pageSize)
	if err != nil {
		switch err {
		case service.ErrInvalidPage,
			service.ErrInvalidPageSize,
			service.ErrInvalidFeedbackType,
			service.ErrInvalidStatus:
			writeErrorResponse(w, http.StatusBadRequest, err)
		default:
			writeErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	type FeedbackListResponse struct {
		Feedbacks  []*models.Feedback   `json:"feedbacks"`
		Pagination *models.Pagination    `json:"pagination"`
	}

	writeJSONResponse(w, http.StatusOK, "success", FeedbackListResponse{
		Feedbacks:  feedbacks,
		Pagination: pagination,
	})
}

func GetFeedbackDetailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorResponse(w, http.StatusMethodNotAllowed, nil)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/admin/feedbacks/")
	idStr := strings.TrimSuffix(path, "/")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, service.ErrFeedbackNotFound)
		return
	}

	feedback, err := service.GetFeedbackByID(id)
	if err != nil {
		if err == service.ErrFeedbackNotFound {
			writeErrorResponse(w, http.StatusNotFound, err)
		} else {
			writeErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSONResponse(w, http.StatusOK, "success", feedback)
}

func UpdateFeedbackStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeErrorResponse(w, http.StatusMethodNotAllowed, nil)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/admin/feedbacks/")
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		writeErrorResponse(w, http.StatusBadRequest, service.ErrFeedbackNotFound)
		return
	}

	idStr := parts[0]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, service.ErrFeedbackNotFound)
		return
	}

	var req models.UpdateFeedbackStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	err = service.UpdateFeedbackStatus(id, &req)
	if err != nil {
		switch err {
		case service.ErrInvalidStatus,
			service.ErrProcessingNoteRequired:
			writeErrorResponse(w, http.StatusBadRequest, err)
		case service.ErrFeedbackNotFound:
			writeErrorResponse(w, http.StatusNotFound, err)
		default:
			writeErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSONResponse(w, http.StatusOK, "status updated successfully", nil)
}

func AddInternalNoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, nil)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/admin/feedbacks/")
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		writeErrorResponse(w, http.StatusBadRequest, service.ErrFeedbackNotFound)
		return
	}

	idStr := parts[0]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, service.ErrFeedbackNotFound)
		return
	}

	var req models.AddInternalNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	err = service.AddInternalNote(id, &req)
	if err != nil {
		switch err {
		case service.ErrNoteTooLong:
			writeErrorResponse(w, http.StatusBadRequest, err)
		case service.ErrFeedbackNotFound:
			writeErrorResponse(w, http.StatusNotFound, err)
		default:
			writeErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSONResponse(w, http.StatusOK, "note added successfully", nil)
}

func GetStatisticsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorResponse(w, http.StatusMethodNotAllowed, nil)
		return
	}

	stats, err := service.GetStatistics()
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	writeJSONResponse(w, http.StatusOK, "success", stats)
}
