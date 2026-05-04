package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// 通知分发器
type NotificationDispatcher struct {
	storage Storage
}

// 创建新的通知分发器
func NewNotificationDispatcher(storage Storage) *NotificationDispatcher {
	return &NotificationDispatcher{
		storage: storage,
	}
}

// 创建通知处理函数
func (d *NotificationDispatcher) CreateNotificationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req NotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证请求
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 确定目标用户列表
	var targetUsers []string
	if req.SendToAll {
		// 发送给全部用户
		targetUsers = GetAllUsers()
	} else {
		// 过滤出存在的用户
		for _, userID := range req.UserIDs {
			if UserExists(userID) {
				// 去重
				exists := false
				for _, u := range targetUsers {
					if u == userID {
						exists = true
						break
					}
				}
				if !exists {
					targetUsers = append(targetUsers, userID)
				}
			}
		}
	}

	// 检查是否有目标用户
	if len(targetUsers) == 0 {
		http.Error(w, "No valid target users", http.StatusBadRequest)
		return
	}

	// 创建通知
	notification := &Notification{
		Title:      req.Title,
		Content:    req.Content,
		Channels:   req.Channels,
		SendToAll:  req.SendToAll,
		UserIDs:    req.UserIDs,
		TotalUsers: len(targetUsers),
	}

	// 保存通知
	if err := d.storage.SaveNotification(notification); err != nil {
		http.Error(w, "Failed to save notification", http.StatusInternalServerError)
		return
	}

	// 分发通知到各个渠道
	d.dispatchNotification(notification, targetUsers)

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":           notification.ID,
		"title":        notification.Title,
		"total_users":  notification.TotalUsers,
		"channels":     notification.Channels,
		"created_at":   notification.CreatedAt,
	})
}

// 获取通知处理函数
func (d *NotificationDispatcher) GetNotificationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从 URL 路径中提取通知 ID
	// 路径格式: /notifications/{id}
	path := strings.TrimPrefix(r.URL.Path, "/notifications/")
	if path == "" {
		http.Error(w, "Notification ID is required", http.StatusBadRequest)
		return
	}

	// 获取通知
	notification, err := d.storage.GetNotification(path)
	if err != nil {
		http.Error(w, "Notification not found", http.StatusNotFound)
		return
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notification)
}

// 分发通知到各个渠道
func (d *NotificationDispatcher) dispatchNotification(notification *Notification, targetUsers []string) {
	// 为每个渠道创建独立的分发流程
	for _, channel := range notification.Channels {
		// 每个渠道使用独立的 goroutine，确保一个渠道失败不影响其他渠道
		go func(ch Channel) {
			log.Printf("Starting to dispatch notification %s to channel %s", notification.ID, ch)
			
			for _, userID := range targetUsers {
				// 检查是否已经发送过（防止重复）
				existingRecords, err := d.storage.GetDeliveryRecordsByUserIDAndChannel(userID, ch)
				if err == nil {
					alreadySent := false
					for _, record := range existingRecords {
						if record.NotificationID == notification.ID {
							alreadySent = true
							break
						}
					}
					if alreadySent {
						log.Printf("Notification %s already sent to user %s via %s, skipping", 
							notification.ID, userID, ch)
						continue
					}
				}

				// 创建送达记录
				record := &DeliveryRecord{
					NotificationID: notification.ID,
					UserID:         userID,
					Channel:        ch,
					IsRead:         false,
				}

				// 根据渠道设置初始状态
				switch ch {
				case ChannelInbox:
					// 站内信直接标记为已送达
					record.Status = StatusDelivered
				case ChannelEmail, ChannelSMS:
					// 邮件和短信先标记为发送中
					record.Status = StatusPending
				}

				// 保存送达记录
				if err := d.storage.SaveDeliveryRecord(record); err != nil {
					log.Printf("Failed to save delivery record for user %s, channel %s: %v", 
						userID, ch, err)
					continue
				}

				// 模拟发送
				d.simulateSend(record, notification)
			}
			
			log.Printf("Finished dispatching notification %s to channel %s", notification.ID, ch)
		}(channel)
	}
}

// 模拟发送
func (d *NotificationDispatcher) simulateSend(record *DeliveryRecord, notification *Notification) {
	switch record.Channel {
	case ChannelInbox:
		// 站内信已经标记为已送达，无需额外处理
		log.Printf("[Inbox] Notification %s delivered to user %s", 
			notification.ID, record.UserID)
		
	case ChannelEmail:
		// 模拟邮件发送
		log.Printf("[Email] Sending notification %s to user %s: Title=%s", 
			notification.ID, record.UserID, notification.Title)
		
		// 这里可以添加实际的邮件发送逻辑，暂时用日志模拟
		// 实际应用中，这里会调用邮件服务，然后根据返回结果更新状态
		// 为了演示，我们假设邮件发送成功，直接更新状态
		// 但根据需求，邮件和短信应该先标记为发送中，后续通过回调更新
		// 所以这里我们只记录日志，不更新状态
		
	case ChannelSMS:
		// 模拟短信发送
		log.Printf("[SMS] Sending notification %s to user %s: Title=%s", 
			notification.ID, record.UserID, notification.Title)
		
		// 同样，这里只记录日志，不更新状态
	}

	// 为了测试方便，我们可以添加一个模拟的回调处理
	// 实际应用中，这个回调会由外部服务调用
	// 这里我们使用一个简单的模拟，假设邮件和短信在 2 秒后发送成功
	if record.Channel == ChannelEmail || record.Channel == ChannelSMS {
		// 注意：这里我们不自动更新状态，而是等待回调
		// 但为了演示，我们可以在日志中说明
		log.Printf("[%s] Notification %s to user %s is in pending state, waiting for callback", 
			strings.ToUpper(string(record.Channel)), notification.ID, record.UserID)
	}
}

// 辅助函数：打印错误
func logError(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	log.Printf("[ERROR] %s", msg)
}
