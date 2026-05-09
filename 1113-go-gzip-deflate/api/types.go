package api

import "time"

type ArchiveRequest struct {
	Files         []string       `json:"files"`
	Recursive     bool           `json:"recursive"`
	IncludeHidden bool           `json:"include_hidden"`
	GlobPatterns  []string       `json:"glob_patterns"`
	ExcludePaths  []string       `json:"exclude_paths"`
	TimeRange     *TimeRange     `json:"time_range"`
	Compression   CompressionOpt `json:"compression"`
	RenameRules   []RenameRule   `json:"rename_rules"`
	ArchiveName   string         `json:"archive_name"`
}

type TimeRange struct {
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
}

type CompressionOpt struct {
	Level  int    `json:"level"`
	Method string `json:"method"`
}

type RenameRule struct {
	Pattern     string `json:"pattern"`
	Replacement string `json:"replacement"`
}

type ArchiveResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	FileCount int    `json:"file_count"`
	TotalSize int64  `json:"total_size"`
	Archived  bool   `json:"archived,omitempty"`
}

type FileInfo struct {
	Path    string    `json:"path"`
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
	Mode    uint32    `json:"mode"`
	IsDir   bool      `json:"is_dir"`
}
