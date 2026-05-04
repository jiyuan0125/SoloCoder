package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	AdminUsername = "admin"
	AdminPassword = "admin123"
)

type UpdateStatusRequest struct {
	Status         string `json:"status"`
	ProcessingNote string `json:"processing_note,omitempty"`
}

type UpdateNoteRequest struct {
	InternalNote string `json:"internal_note"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AdminHandler struct {
	storage *Storage
}

func NewAdminHandler(storage *Storage) *AdminHandler {
	return &AdminHandler{storage: storage}
}

func (h *AdminHandler) validateNote(note string) error {
	if len([]rune(note)) > MaxNoteLength {
		return fmt.Errorf("备注不能超过%d个字符", MaxNoteLength)
	}
	return nil
}

func (h *AdminHandler) validateStatus(status string) error {
	if !ValidStatuses[status] {
		return errors.New("无效的状态值，只能选择：待处理、处理中、已关闭")
	}
	return nil
}

func (h *AdminHandler) checkAuth(r *http.Request) bool {
	username, password, ok := r.BasicAuth()
	if !ok {
		return false
	}
	return username == AdminUsername && password == AdminPassword
}

func (h *AdminHandler) requireAuth(w http.ResponseWriter, r *http.Request) bool {
	if !h.checkAuth(r) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Admin Area"`)
		http.Error(w, `{"error": "未授权，请使用管理员账号登录"}`, http.StatusUnauthorized)
		return false
	}
	return true
}

func (h *AdminHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "请求体解析失败"}`, http.StatusBadRequest)
		return
	}

	if req.Username == AdminUsername && req.Password == AdminPassword {
		response := map[string]interface{}{
			"success": true,
			"message": "登录成功",
		}
		json.NewEncoder(w).Encode(response)
	} else {
		http.Error(w, `{"error": "用户名或密码错误"}`, http.StatusUnauthorized)
	}
}

func (h *AdminHandler) HandleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !h.requireAuth(w, r) {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/api/admin/feedbacks/")
	id := strings.TrimSuffix(path, "/status")
	id = strings.TrimSpace(id)

	if id == "" {
		http.Error(w, `{"error": "反馈ID不能为空"}`, http.StatusBadRequest)
		return
	}

	feedback, exists := h.storage.GetFeedbackByID(id)
	if !exists {
		http.Error(w, `{"error": "反馈不存在"}`, http.StatusNotFound)
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "请求体解析失败"}`, http.StatusBadRequest)
		return
	}

	if err := h.validateStatus(req.Status); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	if feedback.Type == FeedbackTypeBugReport && req.Status == StatusClosed {
		if strings.TrimSpace(req.ProcessingNote) == "" {
			http.Error(w, `{"error": "Bug报告关闭时必须填写处理说明"}`, http.StatusBadRequest)
			return
		}
		feedback.ProcessingNote = strings.TrimSpace(req.ProcessingNote)
	}

	feedback.Status = req.Status
	feedback.UpdatedAt = time.Now()

	if err := h.storage.UpdateFeedback(*feedback); err != nil {
		http.Error(w, `{"error": "更新反馈状态失败"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "状态更新成功",
	}
	json.NewEncoder(w).Encode(response)
}

func (h *AdminHandler) HandleUpdateNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !h.requireAuth(w, r) {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/api/admin/feedbacks/")
	id := strings.TrimSuffix(path, "/note")
	id = strings.TrimSpace(id)

	if id == "" {
		http.Error(w, `{"error": "反馈ID不能为空"}`, http.StatusBadRequest)
		return
	}

	feedback, exists := h.storage.GetFeedbackByID(id)
	if !exists {
		http.Error(w, `{"error": "反馈不存在"}`, http.StatusNotFound)
		return
	}

	var req UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "请求体解析失败"}`, http.StatusBadRequest)
		return
	}

	if err := h.validateNote(req.InternalNote); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	feedback.InternalNote = strings.TrimSpace(req.InternalNote)
	feedback.UpdatedAt = time.Now()

	if err := h.storage.UpdateFeedback(*feedback); err != nil {
		http.Error(w, `{"error": "更新备注失败"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "备注更新成功",
	}
	json.NewEncoder(w).Encode(response)
}
