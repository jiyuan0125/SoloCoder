package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/comment-moderator/pkg/common"
	"github.com/gorilla/mux"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) SetupRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/health", h.HealthCheck).Methods("GET")

	r.HandleFunc("/comments/submit", h.SubmitComment).Methods("POST")
	r.HandleFunc("/comments/edit", h.EditComment).Methods("PUT")
	r.HandleFunc("/comments/report", h.ReportComment).Methods("POST")
	r.HandleFunc("/comments/rejected/{comment_id}", h.ViewRejectedComment).Methods("GET")

	r.HandleFunc("/moderate/pending", h.GetPendingComments).Methods("GET")
	r.HandleFunc("/moderate/approve", h.BatchApprove).Methods("POST")
	r.HandleFunc("/moderate/reject", h.BatchReject).Methods("POST")

	r.HandleFunc("/sensitive-words", h.GetSensitiveWords).Methods("GET")
	r.HandleFunc("/sensitive-words", h.AddSensitiveWord).Methods("POST")
	r.HandleFunc("/sensitive-words", h.RemoveSensitiveWord).Methods("DELETE")

	r.HandleFunc("/moderators", h.RegisterModerator).Methods("POST")
	r.HandleFunc("/moderators/stats", h.GetModeratorStats).Methods("GET")

	r.HandleFunc("/audit-logs", h.GetAuditLogs).Methods("GET")

	return r
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) SubmitComment(w http.ResponseWriter, r *http.Request) {
	var req common.SubmitCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success":false,"message":"无效的请求体"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.service.SubmitComment(req.UserID, req.Content)
	if err != nil {
		http.Error(w, `{"success":false,"message":"服务器错误"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) EditComment(w http.ResponseWriter, r *http.Request) {
	var req common.EditCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success":false,"message":"无效的请求体"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.service.EditComment(req.CommentID, req.UserID, req.Content)
	if err != nil {
		http.Error(w, `{"success":false,"message":"服务器错误"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) ReportComment(w http.ResponseWriter, r *http.Request) {
	var req common.ReportCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success":false,"message":"无效的请求体"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.service.ReportComment(req.CommentID, req.UserID, req.Reason)
	if err != nil {
		http.Error(w, `{"success":false,"message":"服务器错误"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) ViewRejectedComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["comment_id"]
	userID := r.URL.Query().Get("user_id")

	if userID == "" {
		http.Error(w, `{"success":false,"message":"需要提供user_id参数"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.service.ViewRejectedComment(userID, commentID)
	if err != nil {
		http.Error(w, `{"success":false,"message":"服务器错误"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetPendingComments(w http.ResponseWriter, r *http.Request) {
	moderatorID := r.URL.Query().Get("moderator_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	resp, err := h.service.GetPendingComments(moderatorID, limit, offset)
	if err != nil {
		http.Error(w, `{"success":false,"message":"服务器错误"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) BatchApprove(w http.ResponseWriter, r *http.Request) {
	var req common.BatchApproveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success":false,"message":"无效的请求体"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.service.BatchApprove(req.ModeratorID, req.CommentIDs)
	if err != nil {
		http.Error(w, `{"success":false,"message":"服务器错误"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) BatchReject(w http.ResponseWriter, r *http.Request) {
	var req common.BatchRejectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success":false,"message":"无效的请求体"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.service.BatchReject(req.ModeratorID, req.CommentIDs, req.Reason)
	if err != nil {
		http.Error(w, `{"success":false,"message":"服务器错误"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetSensitiveWords(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.GetSensitiveWords()
	if err != nil {
		http.Error(w, `{"success":false,"message":"服务器错误"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) AddSensitiveWord(w http.ResponseWriter, r *http.Request) {
	var req common.AddSensitiveWordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success":false,"message":"无效的请求体"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.service.AddSensitiveWord(req.AdminID, req.Word, req.Level)
	if err != nil {
		http.Error(w, `{"success":false,"message":"服务器错误"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) RemoveSensitiveWord(w http.ResponseWriter, r *http.Request) {
	var req common.RemoveSensitiveWordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success":false,"message":"无效的请求体"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.service.RemoveSensitiveWord(req.AdminID, req.Word)
	if err != nil {
		http.Error(w, `{"success":false,"message":"服务器错误"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) RegisterModerator(w http.ResponseWriter, r *http.Request) {
	var req common.RegisterModeratorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success":false,"message":"无效的请求体"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.service.RegisterModerator(req.AdminID, req.Name)
	if err != nil {
		http.Error(w, `{"success":false,"message":"服务器错误"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetModeratorStats(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.GetModeratorStats()
	if err != nil {
		http.Error(w, `{"success":false,"message":"服务器错误"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.GetAuditLogs()
	if err != nil {
		http.Error(w, `{"success":false,"message":"服务器错误"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
