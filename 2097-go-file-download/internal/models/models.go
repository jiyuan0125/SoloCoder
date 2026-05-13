package models

import (
	"time"
)

type File struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	FileName    string    `json:"file_name"`
	OriginalName string   `json:"original_name"`
	Size        int64     `json:"size"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

type DownloadLink struct {
	ID           string    `json:"id"`
	FileID       string    `json:"file_id"`
	UserID       string    `json:"user_id"`
	MaxDownloads int       `json:"max_downloads"`
	Downloaded   int       `json:"downloaded"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type DownloadRecord struct {
	ID          string    `json:"id"`
	LinkID      string    `json:"link_id"`
	FileID      string    `json:"file_id"`
	UserID      string    `json:"user_id"`
	DownloadedAt time.Time `json:"downloaded_at"`
	IPAddress   string    `json:"ip_address"`
}

type UploadResult struct {
	OriginalName string `json:"original_name"`
	Success      bool   `json:"success"`
	FileID       string `json:"file_id,omitempty"`
	Error        string `json:"error,omitempty"`
}

type GenerateLinkRequest struct {
	FileID       string `json:"file_id"`
	MaxDownloads int    `json:"max_downloads"`
	ExpiresIn    int64  `json:"expires_in"`
}

type UserFilesRequest struct {
	UserID string `json:"user_id"`
	Keyword string `json:"keyword,omitempty"`
}
