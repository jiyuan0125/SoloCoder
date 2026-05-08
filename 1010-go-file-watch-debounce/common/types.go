package common

import "time"

type EventType string

const (
	EventCreate  EventType = "create"
	EventModify  EventType = "modify"
	EventDelete  EventType = "delete"
	EventRename  EventType = "rename"
	EventReplace EventType = "replace"
)

type Event struct {
	Path      string    `json:"path"`
	OldPath   string    `json:"old_path,omitempty"`
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

type AddWatchRequest struct {
	Path        string        `json:"path"`
	Debounce    time.Duration `json:"debounce"`
	MaxWait     time.Duration `json:"max_wait"`
	CallbackURL string        `json:"callback_url"`
}

type AddWatchResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
	Error   string `json:"error,omitempty"`
}

type RemoveWatchRequest struct {
	ID string `json:"id"`
}

type RemoveWatchResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type ListWatchesResponse struct {
	Success bool               `json:"success"`
	Watches []WatchInformation `json:"watches"`
	Error   string             `json:"error,omitempty"`
}

type WatchInformation struct {
	ID          string        `json:"id"`
	Path        string        `json:"path"`
	Debounce    time.Duration `json:"debounce"`
	MaxWait     time.Duration `json:"max_wait"`
	CallbackURL string        `json:"callback_url"`
	Active      bool          `json:"active"`
}

type ListEventsResponse struct {
	Success bool    `json:"success"`
	Events  []Event `json:"events"`
	Error   string  `json:"error,omitempty"`
}
