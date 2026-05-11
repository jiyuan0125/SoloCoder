package api

import "config-hotload/pkg/config"

type GetConfigResponse struct {
	Success bool                   `json:"success"`
	Config  map[string]interface{} `json:"config,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

type GetFieldRequest struct {
	Path string `json:"path"`
}

type GetFieldResponse struct {
	Success bool        `json:"success"`
	Value   interface{} `json:"value,omitempty"`
	Exists  bool        `json:"exists,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type ReloadResponse struct {
	Success bool         `json:"success"`
	Changes []ChangeInfo `json:"changes,omitempty"`
	Error   string       `json:"error,omitempty"`
}

type ChangeInfo struct {
	Path string      `json:"path"`
	Old  interface{} `json:"old,omitempty"`
	New  interface{} `json:"new,omitempty"`
	Type string      `json:"type"`
}

type HistoryResponse struct {
	Success bool               `json:"success"`
	History []HistoryRecord    `json:"history,omitempty"`
	Error   string             `json:"error,omitempty"`
}

type HistoryRecord struct {
	Timestamp string       `json:"timestamp"`
	Changes   []ChangeInfo `json:"changes"`
}

type WatchEvent struct {
	Timestamp string       `json:"timestamp"`
	Changes   []ChangeInfo `json:"changes"`
}

func ConvertChanges(changes []config.Change) []ChangeInfo {
	result := make([]ChangeInfo, len(changes))
	for i, c := range changes {
		result[i] = ChangeInfo{
			Path: c.Path,
			Old:  c.Old,
			New:  c.New,
			Type: string(c.Type),
		}
	}
	return result
}

func ConvertHistory(records []config.ChangeRecord) []HistoryRecord {
	result := make([]HistoryRecord, len(records))
	for i, r := range records {
		result[i] = HistoryRecord{
			Timestamp: r.Timestamp.Format("2006-01-02T15:04:05.000Z07:00"),
			Changes:   ConvertChanges(r.Changes),
		}
	}
	return result
}
