package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// 送达状态管理器
type DeliveryStatusManager struct {
	storage Storage
}

// 创建新的送达状态管理器
func NewDeliveryStatusManager(storage Storage) *DeliveryStatusManager {
	return &DeliveryStatusManager{
		storage: storage,
	}
}

// 状态更新请求
type StatusUpdateRequest struct {
	RecordID string         `json:"record_id"`
	Status   DeliveryStatus `json:"status"`
}

// 更新送达状态处理函数（供回调使用）
func (m *DeliveryStatusManager) UpdateStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StatusUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证状态值
	if req.Status != StatusDelivered && req.Status != StatusFailed {
		http.Error(w, "Invalid status value. Only 'delivered' or 'failed' are allowed", http.StatusBadRequest)
		return
	}

	// 更新状态
	if err := m.storage.UpdateDeliveryRecordStatus(req.RecordID, req.Status); err != nil {
		log.Printf("Failed to update delivery record status: %v", err)
		http.Error(w, "Failed to update status", http.StatusInternalServerError)
		return
	}

	log.Printf("Delivery record %s status updated to %s", req.RecordID, req.Status)

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Status updated successfully",
	})
}

// 获取统计信息处理函数
func (m *DeliveryStatusManager) GetStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从查询参数中获取通知 ID
	notificationID := r.URL.Query().Get("notification_id")
	if notificationID == "" {
		http.Error(w, "Notification ID is required", http.StatusBadRequest)
		return
	}

	// 获取通知信息
	notification, err := m.storage.GetNotification(notificationID)
	if err != nil {
		http.Error(w, "Notification not found", http.StatusNotFound)
		return
	}

	// 获取该通知的所有送达记录
	records, err := m.storage.GetDeliveryRecordsByNotificationID(notificationID)
	if err != nil {
		log.Printf("Failed to get delivery records: %v", err)
		http.Error(w, "Failed to get statistics", http.StatusInternalServerError)
		return
	}

	// 按渠道统计
	statsMap := make(map[Channel]*ChannelStats)
	for _, ch := range notification.Channels {
		statsMap[ch] = &ChannelStats{
			Channel:   ch,
			Total:     0,
			Delivered: 0,
			Failed:    0,
			Pending:   0,
		}
	}

	// 统计各状态
	for _, record := range records {
		stats, exists := statsMap[record.Channel]
		if !exists {
			continue
		}

		stats.Total++
		switch record.Status {
		case StatusDelivered:
			stats.Delivered++
		case StatusFailed:
			stats.Failed++
		case StatusPending:
			stats.Pending++
		}
	}

	// 构建统计结果
	channelStats := make([]ChannelStats, 0, len(statsMap))
	for _, stats := range statsMap {
		channelStats = append(channelStats, *stats)
	}

	// 构建最终响应
	notificationStats := &NotificationStats{
		NotificationID: notificationID,
		TotalUsers:     notification.TotalUsers,
		ChannelStats:   channelStats,
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notificationStats)
}

// 批量更新状态（用于测试或内部调用）
func (m *DeliveryStatusManager) BatchUpdateStatus(notificationID string, channel Channel, status DeliveryStatus) error {
	// 获取该通知和渠道的所有记录
	records, err := m.storage.GetDeliveryRecordsByNotificationID(notificationID)
	if err != nil {
		return err
	}

	// 过滤出指定渠道的记录
	for _, record := range records {
		if record.Channel == channel {
			if err := m.storage.UpdateDeliveryRecordStatus(record.ID, status); err != nil {
				log.Printf("Failed to update record %s: %v", record.ID, err)
			}
		}
	}

	return nil
}

// 辅助函数：解析渠道字符串
func parseChannel(channelStr string) (Channel, bool) {
	channelStr = strings.ToLower(channelStr)
	switch channelStr {
	case "inbox":
		return ChannelInbox, true
	case "email":
		return ChannelEmail, true
	case "sms":
		return ChannelSMS, true
	default:
		return "", false
	}
}

// 辅助函数：解析状态字符串
func parseStatus(statusStr string) (DeliveryStatus, bool) {
	statusStr = strings.ToLower(statusStr)
	switch statusStr {
	case "pending":
		return StatusPending, true
	case "delivered":
		return StatusDelivered, true
	case "failed":
		return StatusFailed, true
	default:
		return "", false
	}
}
