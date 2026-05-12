package handlers

import (
	"encoding/json"
	"net/http"
	"ohims/internal/models"
	"ohims/pkg/db"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type TodoHandler struct{}

func (h *TodoHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	status := r.URL.Query().Get("status")
	assignee := r.URL.Query().Get("assignee")

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}

	var todos []models.TodoItem
	var total int64

	query := db.GetDB().Model(&models.TodoItem{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if assignee != "" {
		query = query.Where("assignee = ?", assignee)
	}
	query.Count(&total)

	offset := (page - 1) * size
	query.Order("due_date asc, priority desc").Offset(offset).Limit(size).Find(&todos)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":       todos,
		"page":       page,
		"size":       size,
		"total":      total,
		"totalPages": (total + int64(size) - 1) / int64(size),
	})
}

func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var todo models.TodoItem
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, `{"error":"无效的请求体"}`, http.StatusBadRequest)
		return
	}

	if todo.DueDate.IsZero() {
		todo.DueDate = time.Now().AddDate(0, 0, 7)
	}
	if todo.Status == "" {
		todo.Status = "待处理"
	}
	if todo.Priority == "" {
		todo.Priority = "中"
	}
	todo.CreatedAt = time.Now()
	todo.UpdatedAt = time.Now()

	if err := db.GetDB().Create(&todo).Error; err != nil {
		http.Error(w, `{"error":"创建待办失败"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todo)
}

func (h *TodoHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/todos/")
	if strings.Contains(idStr, "/sub") {
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return
	}

	var todo models.TodoItem
	if err := db.GetDB().First(&todo, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, `{"error":"待办不存在"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"查询失败"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/todos/")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"无效的ID"}`, http.StatusBadRequest)
		return
	}

	var todo models.TodoItem
	if err := db.GetDB().First(&todo, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, `{"error":"待办不存在"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"查询失败"}`, http.StatusInternalServerError)
		return
	}

	var updateData models.TodoItem
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, `{"error":"无效的请求体"}`, http.StatusBadRequest)
		return
	}

	if updateData.Title != "" {
		todo.Title = updateData.Title
	}
	if updateData.Description != "" {
		todo.Description = updateData.Description
	}
	if updateData.Assignee != "" {
		todo.Assignee = updateData.Assignee
	}
	if !updateData.DueDate.IsZero() {
		todo.DueDate = updateData.DueDate
	}
	if updateData.Status != "" {
		todo.Status = updateData.Status
	}
	if updateData.Priority != "" {
		todo.Priority = updateData.Priority
	}
	todo.UpdatedAt = time.Now()

	db.GetDB().Save(&todo)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/todos/")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"无效的ID"}`, http.StatusBadRequest)
		return
	}

	db.GetDB().Delete(&models.TodoItem{}, id)
	w.WriteHeader(http.StatusNoContent)
}
