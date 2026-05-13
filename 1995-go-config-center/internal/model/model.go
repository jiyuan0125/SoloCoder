package model

import "time"

type Config struct {
	Env       string    `json:"env"`
	Project   string    `json:"project"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type VersionSnapshot struct {
	Version    int       `json:"version"`
	Env        string    `json:"env"`
	Project    string    `json:"project"`
	Key        string    `json:"key"`
	Value      string    `json:"value"`
	Operation  string    `json:"operation"`
	PreviousVersion int `json:"previous_version"`
	CreatedAt  time.Time `json:"created_at"`
}

type WatchRegistration struct {
	ID        string    `json:"id"`
	Env       string    `json:"env"`
	Project   string    `json:"project"`
	Key       string    `json:"key"`
	Callback  string    `json:"callback_url"`
	IsKeyWatch bool     `json:"is_key_watch"`
	Status    string    `json:"status"`
	ConsecutiveFailures int `json:"consecutive_failures"`
	CreatedAt time.Time `json:"created_at"`
}

type DiffItem struct {
	Key        string `json:"key"`
	OldValue   string `json:"old_value"`
	NewValue   string `json:"new_value"`
	ChangeType string `json:"change_type"`
}

type WatchPayload struct {
	Env      string    `json:"env"`
	Project  string    `json:"project"`
	Key      string    `json:"key"`
	OldValue string    `json:"old_value"`
	NewValue string    `json:"new_value"`
	Version  int       `json:"version"`
	Time     time.Time `json:"time"`
}
