package models

import (
	"time"
)

type BackupStatus string

const (
	StatusSuccess      BackupStatus = "success"
	StatusPartialFail  BackupStatus = "partial_fail"
	StatusInterrupted  BackupStatus = "interrupted"
)

type FileInfo struct {
	RelativePath string    `json:"relative_path"`
	Size         int64     `json:"size"`
	ModTime      time.Time `json:"mod_time"`
}

type BackupRecord struct {
	ID          string       `json:"id"`
	StartTime   time.Time    `json:"start_time"`
	EndTime     time.Time    `json:"end_time"`
	SourcePath  string       `json:"source_path"`
	TargetPath  string       `json:"target_path"`
	IsCompressed bool        `json:"is_compressed"`
	Status      BackupStatus `json:"status"`
	Files       []FileInfo   `json:"files"`
	FailedFiles []string     `json:"failed_files"`
}

type ResumeInfo struct {
	BackupID       string    `json:"backup_id"`
	CompletedFiles []string  `json:"completed_files"`
	LastModified   time.Time `json:"last_modified"`
}

type Metadata struct {
	BackupRecords []BackupRecord `json:"backup_records"`
	LatestID      string         `json:"latest_id"`
}
