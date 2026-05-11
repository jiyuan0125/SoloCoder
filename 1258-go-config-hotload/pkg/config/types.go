package config

import (
	"time"
)

type ChangeType string

const (
	ChangeTypeAdd    ChangeType = "add"
	ChangeTypeModify ChangeType = "modify"
	ChangeTypeDelete ChangeType = "delete"
)

type Change struct {
	Path    string      `json:"path"`
	Old     interface{} `json:"old"`
	New     interface{} `json:"new"`
	Type    ChangeType  `json:"type"`
	Affected []string   `json:"affected,omitempty"`
}

type ChangeRecord struct {
	Timestamp time.Time `json:"timestamp"`
	Changes   []Change  `json:"changes"`
}

type ListenerFunc func(change Change)

type ConfigFormat string

const (
	FormatYAML ConfigFormat = "yaml"
	FormatJSON ConfigFormat = "json"
)

type Options struct {
	FilePath       string
	Format         ConfigFormat
	EnableRef      bool
	StableWaitTime time.Duration
}
