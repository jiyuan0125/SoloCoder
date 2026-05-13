package model

import (
	"time"
)

type File struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	Filename        string    `json:"filename"`
	StoredPath      string    `json:"-"`
	Size            int64     `json:"size"`
	UploadTime      time.Time `json:"upload_time"`
	ExpireTime      time.Time `json:"expire_time"`
	MaxDownloads    int       `json:"max_downloads"`
	CurrentDownload int       `json:"current_download"`
}

type DownloadRecord struct {
	ID        int64     `json:"id"`
	FileID    string    `json:"file_id"`
	UserID    string    `json:"user_id"`
	Filename  string    `json:"filename"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Time      time.Time `json:"time"`
}
