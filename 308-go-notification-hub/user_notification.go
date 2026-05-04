package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"
)

// 用户通知处理器
type UserNotificationHandler struct {
	storage Storage
}

// 创建新的用户通知处理器
func NewUserNotificationHandler(storage Storage) *UserNotificationHandler {
	return &UserNotificationHandler{
		storage: storage,
	}
}

// 用户通知视图（用于返回给用户）
type UserNotificationView struct {
	RecordID       string    `json:"record_id"`
	NotificationID string    `json:"notification_id"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	Channel        Channel   `json:"channel"`
	Status         string    `json:"status"`
	IsRead         bool      `json:"is_read"`
	CreatedAt      time.Time `json:"created_at"`
}

// 标记已读请求
type MarkAsReadRequest struct {
	RecordID string `json:"record_id"`
}

// 获取用户通知列表处理函数
func (h *UserNotificationHandler) GetUserNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从查询参数中获取用户 ID
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// 检查用户是否存在
	if !UserExists(userID) {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// 解析已读未读筛选参数
	isReadParam := r.URL.Query().Get("is_read")
	var isReadFilter *bool
	if isReadParam != "" {
		val, err := strconv.ParseBool(isReadParam)
		if err != nil {
			http.Error(w, "Invalid is_read parameter", http.StatusBadRequest)
			return
		}
		isReadFilter = &val
	}

	// 获取用户的所有送达记录
	records, err := h.storage.GetDeliveryRecordsByUserID(userID)
	if err != nil {
		log.Printf("Failed to get delivery records for user %s: %v", userID, err)
		http.Error(w, "Failed to get notifications", http.StatusInternalServerError)
		return
	}

	// 构建用户通知视图列表
	var notifications []UserNotificationView
	for _, record := range records {
		// 应用已读未读筛选（仅对站内信有效）
		if isReadFilter != nil && record.Channel == ChannelInbox {
			if record.IsRead != *isReadFilter {
				continue
			}
		}

		// 获取通知详情
		notification, err := h.storage.GetNotification(record.NotificationID)
		if err != nil {
			log.Printf("Failed to get notification %s: %v", record.NotificationID, err)
			continue
		}

		// 构建视图
		view := UserNotificationView{
			RecordID:       record.ID,
			NotificationID: record.NotificationID,
			Title:          notification.Title,
			Content:        notification.Content,
			Channel:        record.Channel,
			Status:         string(record.Status),
			IsRead:         record.IsRead,
			CreatedAt:      record.CreatedAt,
		}

		notifications = append(notifications, view)
	}

	// 按创建时间倒序排列
	sort.Slice(notifications, func(i, j int) bool {
		return notifications[i].CreatedAt.After(notifications[j].CreatedAt)
	})

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":       userID,
		"notifications": notifications,
		"count":         len(notifications),
	})
}

// 标记通知为已读处理函数
func (h *UserNotificationHandler) MarkAsReadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MarkAsReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.RecordID == "" {
		http.Error(w, "Record ID is required", http.StatusBadRequest)
		return
	}

	// 标记为已读
	if err := h.storage.MarkAsRead(req.RecordID); err != nil {
		log.Printf("Failed to mark record %s as read: %v", req.RecordID, err)
		if err.Error() == "only inbox notifications can be marked as read" {
			http.Error(w, "Only inbox notifications can be marked as read", http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to mark as read", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("Record %s marked as read", req.RecordID)

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Notification marked as read successfully",
	})
}

// 批量标记已读（可选功能）
func (h *UserNotificationHandler) BatchMarkAsReadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RecordIDs []string `json:"record_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.RecordIDs) == 0 {
		http.Error(w, "Record IDs are required", http.StatusBadRequest)
		return
	}

	successCount := 0
	failCount := 0
	for _, recordID := range req.RecordIDs {
		if err := h.storage.MarkAsRead(recordID); err != nil {
			log.Printf("Failed to mark record %s as read: %v", recordID, err)
			failCount++
		} else {
			successCount++
		}
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Batch mark as read completed",
		"success_count": successCount,
		"fail_count":    failCount,
	})
}
