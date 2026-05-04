package models

import (
	"time"
)

type Channel string

const (
	ChannelInSite  Channel = "insite"
	ChannelEmail   Channel = "email"
	ChannelSMS     Channel = "sms"
)

type DeliveryStatus string

const (
	StatusPending    DeliveryStatus = "pending"
	StatusSending    DeliveryStatus = "sending"
	StatusDelivered  DeliveryStatus = "delivered"
	StatusFailed     DeliveryStatus = "failed"
)

type Notification struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Channels    []Channel `json:"channels"`
	TargetAll   bool      `json:"target_all"`
	TargetUsers []string  `json:"target_users"`
	TotalUsers  int       `json:"total_users"`
	CreatedAt   time.Time `json:"created_at"`
}

type UserDelivery struct {
	ID             string         `json:"id"`
	NotificationID string         `json:"notification_id"`
	UserID         string         `json:"user_id"`
	Channel        Channel        `json:"channel"`
	Status         DeliveryStatus `json:"status"`
	IsRead         bool           `json:"is_read"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type Statistics struct {
	NotificationID string                `json:"notification_id"`
	Title          string                `json:"title"`
	TotalUsers     int                   `json:"total_users"`
	ChannelStats   map[Channel]ChannelStat `json:"channel_stats"`
}

type ChannelStat struct {
	Channel   Channel `json:"channel"`
	Total     int     `json:"total"`
	Delivered int     `json:"delivered"`
	Failed    int     `json:"failed"`
	Pending   int     `json:"pending"`
}

type CreateNotificationRequest struct {
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Channels    []Channel `json:"channels"`
	TargetAll   bool      `json:"target_all"`
	TargetUsers []string  `json:"target_users"`
}

type CreateNotificationResponse struct {
	NotificationID string `json:"notification_id"`
	TotalUsers     int    `json:"total_users"`
}

type UserNotification struct {
	ID             string    `json:"id"`
	NotificationID string    `json:"notification_id"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	Channel        Channel   `json:"channel"`
	Status         DeliveryStatus `json:"status"`
	IsRead         bool      `json:"is_read"`
	CreatedAt      time.Time `json:"created_at"`
}

type StatisticsResponse struct {
	Statistics []Statistics `json:"statistics"`
}

type UserListNotificationsRequest struct {
	UserID string `json:"user_id"`
	IsRead *bool  `json:"is_read,omitempty"`
}

type UserListNotificationsResponse struct {
	Notifications []UserNotification `json:"notifications"`
}

type MarkAsReadRequest struct {
	UserID       string `json:"user_id"`
	DeliveryID   string `json:"delivery_id"`
}

type MarkAsReadResponse struct {
	Success bool `json:"success"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
