package main

import (
	"errors"
	"time"
)

// 发送渠道类型
type Channel string

const (
	ChannelInbox  Channel = "inbox"  // 站内信
	ChannelEmail  Channel = "email"  // 邮件
	ChannelSMS    Channel = "sms"    // 短信
)

// 送达状态
type DeliveryStatus string

const (
	StatusPending    DeliveryStatus = "pending"    // 发送中
	StatusDelivered  DeliveryStatus = "delivered"  // 已送达
	StatusFailed     DeliveryStatus = "failed"     // 发送失败
)

// 通知请求
type NotificationRequest struct {
	Title       string   `json:"title"`
	Content     string   `json:"content"`
	Channels    []Channel `json:"channels"`
	UserIDs     []string  `json:"user_ids,omitempty"`
	SendToAll   bool      `json:"send_to_all"`
}

// 验证通知请求
func (r *NotificationRequest) Validate() error {
	if r.Title == "" {
		return errors.New("标题不能为空")
	}
	if len(r.Title) > 50 {
		return errors.New("标题不能超过50个字符")
	}
	if len(r.Channels) == 0 {
		return errors.New("至少需要选择一个发送渠道")
	}
	if !r.SendToAll && len(r.UserIDs) == 0 {
		return errors.New("需要指定目标用户或选择发送给全部用户")
	}
	return nil
}

// 通知
type Notification struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Channels    []Channel `json:"channels"`
	SendToAll   bool      `json:"send_to_all"`
	UserIDs     []string  `json:"user_ids,omitempty"`
	TotalUsers  int       `json:"total_users"` // 实际发送的总人数
	CreatedAt   time.Time `json:"created_at"`
}

// 送达记录
type DeliveryRecord struct {
	ID             string         `json:"id"`
	NotificationID string         `json:"notification_id"`
	UserID         string         `json:"user_id"`
	Channel        Channel        `json:"channel"`
	Status         DeliveryStatus `json:"status"`
	IsRead         bool           `json:"is_read"` // 仅站内信有效
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// 统计信息
type ChannelStats struct {
	Channel    Channel `json:"channel"`
	Total      int     `json:"total"`
	Delivered  int     `json:"delivered"`
	Failed     int     `json:"failed"`
	Pending    int     `json:"pending"`
}

type NotificationStats struct {
	NotificationID string         `json:"notification_id"`
	TotalUsers     int            `json:"total_users"`
	ChannelStats   []ChannelStats `json:"channel_stats"`
}

// 用户通知查询参数
type UserNotificationQuery struct {
	UserID string `json:"user_id"`
	IsRead *bool  `json:"is_read,omitempty"` // nil 表示全部
}

// 模拟用户数据 - 实际应用中应该从用户服务获取
var mockUsers = []string{
	"user_001", "user_002", "user_003", "user_004", "user_005",
	"user_006", "user_007", "user_008", "user_009", "user_010",
}

// 获取所有用户ID
func GetAllUsers() []string {
	// 返回副本，防止外部修改
	result := make([]string, len(mockUsers))
	copy(result, mockUsers)
	return result
}

// 检查用户是否存在
func UserExists(userID string) bool {
	for _, u := range mockUsers {
		if u == userID {
			return true
		}
	}
	return false
}
