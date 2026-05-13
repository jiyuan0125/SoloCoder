package models

import "time"

type WatchDirectory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Path      string    `gorm:"uniqueIndex;not null" json:"path"`
	Recursive bool      `gorm:"not null" json:"recursive"`
	Callback  string    `gorm:"not null" json:"callback"`
	Status    string    `gorm:"default:active" json:"status"` // active, paused
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FileEvent struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Path      string    `gorm:"not null" json:"path"`
	Event     string    `gorm:"not null" json:"event"` // create, modify, delete, rename
	Status    string    `gorm:"default:pending" json:"status"`
	WatchID   uint      `gorm:"not null" json:"watch_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NotificationEvent struct {
	WatchID   uint   `json:"watch_id"`
	Path      string `json:"path"`
	Event     string `json:"event"`
	Timestamp int64  `json:"timestamp"`
}
