package models

import "time"

type ProcessTask struct {
	ID        int64     `json:"id"`
	Status    string    `json:"status"`
	Total     int       `json:"total"`
	Success   int       `json:"success"`
	Failed    int       `json:"failed"`
	CreatedAt time.Time `json:"created_at"`
	FinishedAt time.Time `json:"finished_at,omitempty"`
	ZipPath   string    `json:"zip_path,omitempty"`
	Report    string    `json:"report,omitempty"`
}

type ImageOperation struct {
	Type       string  `json:"type"`
	Width      int     `json:"width,omitempty"`
	Height     int     `json:"height,omitempty"`
	X          int     `json:"x,omitempty"`
	Y          int     `json:"y,omitempty"`
	Format     string  `json:"format,omitempty"`
	Watermark  string  `json:"watermark,omitempty"`
	Opacity    float64 `json:"opacity,omitempty"`
	Quality    int     `json:"quality,omitempty"`
}

type ProcessRequest struct {
	Operations []ImageOperation `json:"operations"`
}

type ProcessResult struct {
	Total      int           `json:"total"`
	Success    int           `json:"success"`
	Failed     int           `json:"failed"`
	Successful []string      `json:"successful"`
	Skipped    []SkippedFile `json:"skipped"`
}

type SkippedFile struct {
	Filename string `json:"filename"`
	Reason   string `json:"reason"`
}
