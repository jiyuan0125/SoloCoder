package api

import "time"

type MetadataRequest struct {
	FilePath string `json:"file_path"`
}

type MetadataResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Metadata  *Metadata   `json:"metadata,omitempty"`
}

type Metadata struct {
	Format      string            `json:"format"`
	Duration    float64           `json:"duration_seconds"`
	Bitrate     int               `json:"bitrate_bps"`
	FileSize    int64             `json:"file_size_bytes"`
	FormatTags  map[string]string `json:"format_tags,omitempty"`
	AudioTags   map[string]string `json:"audio_tags,omitempty"`
	LastUpdated time.Time         `json:"last_updated,omitempty"`
}
