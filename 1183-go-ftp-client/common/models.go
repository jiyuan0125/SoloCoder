package common

import "time"

type TransferMode string

const (
	ModeActive TransferMode = "active"
	ModePassive TransferMode = "passive"
)

type FTPConfig struct {
	Host     string       `json:"host"`
	Port     int          `json:"port"`
	Username string       `json:"username"`
	Password string       `json:"password"`
	Mode     TransferMode `json:"mode"`
}

type OperationType string

const (
	OpListDir     OperationType = "list_dir"
	OpChangeDir   OperationType = "change_dir"
	OpMakeDir     OperationType = "make_dir"
	OpRemoveDir   OperationType = "remove_dir"
	OpDownload    OperationType = "download"
	OpUpload      OperationType = "upload"
	OpGetCurrentDir OperationType = "get_current_dir"
)

type FileOperationRequest struct {
	Config      FTPConfig      `json:"config"`
	Operation   OperationType  `json:"operation"`
	RemotePath  string         `json:"remote_path,omitempty"`
	LocalPath   string         `json:"local_path,omitempty"`
	UseMLSD     bool           `json:"use_mlsd,omitempty"`
	Recursive   bool           `json:"recursive,omitempty"`
}

type FileEntry struct {
	Name        string    `json:"name"`
	IsDirectory bool      `json:"is_directory"`
	Size        int64     `json:"size"`
	Modified    time.Time `json:"modified,omitempty"`
	Permissions string    `json:"permissions,omitempty"`
}

type ListDirResponse struct {
	Success bool        `json:"success"`
	Entries []FileEntry `json:"entries,omitempty"`
	Error   string      `json:"error,omitempty"`
	CurrentDir string    `json:"current_dir,omitempty"`
}

type SimpleResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

type TransferStatus string

const (
	StatusQueued    TransferStatus = "queued"
	StatusRunning   TransferStatus = "running"
	StatusPaused    TransferStatus = "paused"
	StatusCompleted TransferStatus = "completed"
	StatusFailed    TransferStatus = "failed"
	StatusCancelled TransferStatus = "cancelled"
)

type TransferType string

const (
	TransferDownload TransferType = "download"
	TransferUpload   TransferType = "upload"
)

type TransferInfo struct {
	ID          string         `json:"id"`
	Type        TransferType   `json:"type"`
	RemotePath  string         `json:"remote_path"`
	LocalPath   string         `json:"local_path"`
	TotalSize   int64          `json:"total_size"`
	Transferred int64          `json:"transferred"`
	Status      TransferStatus `json:"status"`
	Error       string         `json:"error,omitempty"`
	StartTime   time.Time      `json:"start_time,omitempty"`
	EndTime     time.Time      `json:"end_time,omitempty"`
}

type QueueResponse struct {
	Success   bool           `json:"success"`
	Transfers []TransferInfo `json:"transfers,omitempty"`
	Error     string         `json:"error,omitempty"`
}

type TransferCancelRequest struct {
	TransferID string `json:"transfer_id"`
}

type TransferCancelResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
