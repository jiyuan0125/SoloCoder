package models

import "time"

type Param struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	In          string `json:"in"`
}

type ReturnValue struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type Example struct {
	Request  string `json:"request"`
	Response string `json:"response"`
}

type API struct {
	ID              int64       `json:"id"`
	VersionID       int64       `json:"version_id"`
	Module          string      `json:"module"`
	Path            string      `json:"path"`
	Method          string      `json:"method"`
	HandlerName     string      `json:"handler_name"`
	Description     string      `json:"description"`
	Params          []Param     `json:"params"`
	Returns         []ReturnValue `json:"returns"`
	Example         Example     `json:"example"`
	IsComplete      bool        `json:"is_complete"`
	MissingFields   []string    `json:"missing_fields"`
	CreatedAt       time.Time   `json:"created_at"`
}

type DocumentVersion struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
	SourceDir string    `json:"source_dir"`
}

type DiffItem struct {
	Module      string `json:"module"`
	Path        string `json:"path"`
	Method      string `json:"method"`
	HandlerName string `json:"handler_name"`
}

type DiffResult struct {
	Added    []DiffItem `json:"added"`
	Removed  []DiffItem `json:"removed"`
	Modified []DiffItem `json:"modified"`
}
