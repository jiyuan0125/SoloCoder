package main

import "time"

type Department struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type User struct {
	ID          int64  `json:"id"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Username    string `json:"username"`
	DepartmentID *int64 `json:"department_id,omitempty"`
	CreatedAt   string `json:"created_at"`
}

type NotificationPreference struct {
	ID         int64  `json:"id"`
	UserID       int64  `json:"user_id"`
	MessageType  string `json:"message_type"`
	Channel      string `json:"channel"`
}

type NotificationStatus string

const (
	StatusSuccess    NotificationStatus = "成功"
	StatusFailed     NotificationStatus = "发送失败"
	StatusDowngraded NotificationStatus = "已降级"
	StatusPending    NotificationStatus = "待重试"
)

type NotificationRecord struct {
	ID             int64              `json:"id"`
	UserID         int64              `json:"user_id"`
	MessageType    string             `json:"message_type"`
	Content        string             `json:"content"`
	Channel        string             `json:"channel"`
	Status         NotificationStatus `json:"status"`
	RetryCount     int                `json:"retry_count"`
	NextRetryAt    *string            `json:"next_retry_at,omitempty"`
	CreatedAt      string             `json:"created_at"`
	UpdatedAt      string             `json:"updated_at"`
	DegradedFrom   *string            `json:"degraded_from,omitempty"`
}

type SendNotificationRequest struct {
	UserID      int64  `json:"user_id"`
	MessageType string `json:"message_type"`
	Content     string `json:"content"`
}

type BatchNotificationRequest struct {
	DepartmentID int64  `json:"department_id"`
	MessageType  string `json:"message_type"`
	Content      string `json:"content"`
}

type NotificationChannel string

const (
	ChannelEmail     NotificationChannel = "email"
	ChannelSMS       NotificationChannel = "sms"
	ChannelSite      NotificationChannel = "site"
)

func (c NotificationChannel) String() string {
	return string(c)
}

func ValidChannel(channel string) bool {
	switch channel {
	case ChannelEmail.String(), ChannelSMS.String(), ChannelSite.String():
		return true
	default:
		return false
	}
}

func GetRetryInterval() time.Duration {
	return 5 * time.Minute
}

func MaxRetryCount() int {
	return 3
}
