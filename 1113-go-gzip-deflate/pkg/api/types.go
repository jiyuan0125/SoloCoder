package api

import "time"

type ArchiveRequest struct {
	Files          map[string]string   `json:"files"`
	Directories    []string            `json:"directories"`
	Compression    int                 `json:"compression"`
	IncludePattern []string            `json:"include_pattern"`
	ExcludePattern []string            `json:"exclude_pattern"`
	TimeStart      *time.Time          `json:"time_start"`
	TimeEnd        *time.Time          `json:"time_end"`
	StripPrefix    string              `json:"strip_prefix"`
	AddPrefix      string              `json:"add_prefix"`
	FollowSymlinks bool                `json:"follow_symlinks"`
	IncludeEmpty   bool                `json:"include_empty"`
	Renames        map[string]string   `json:"renames"`
}

type ArchiveResponse struct {
	Success       bool   `json:"success"`
	FileCount     int    `json:"file_count"`
	WrittenSize   int64  `json:"written_size"`
	ErrorMessage  string `json:"error_message,omitempty"`
}

type FileInfo struct {
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	Mode    uint32    `json:"mode"`
	ModTime time.Time `json:"mod_time"`
	IsDir   bool      `json:"is_dir"`
}

type ListFilesRequest struct {
	Path       string   `json:"path"`
	Patterns   []string `json:"patterns"`
	Recursive  bool     `json:"recursive"`
}

type ListFilesResponse struct {
	Success bool       `json:"success"`
	Files   []FileInfo `json:"files"`
	Error   string     `json:"error,omitempty"`
}

type ServerConfig struct {
	Address     string
	MaxFileSize int64
	MaxFiles    int
	Timeout     time.Duration
}

type ClientConfig struct {
	ServerURL string
	Timeout   time.Duration
}
