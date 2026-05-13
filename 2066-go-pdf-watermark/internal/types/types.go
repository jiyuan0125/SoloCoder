package types

import "time"

type Job struct {
	ID         string    `json:"id"`
	Filename   string    `json:"filename"`
	Filesize   int64     `json:"filesize"`
	Status     string    `json:"status"`
	Error      string    `json:"error,omitempty"`
	TotalPages int       `json:"total_pages"`
	OutputPath string    `json:"-"`
	CreatedAt  time.Time `json:"created_at"`
}

type WatermarkConfig struct {
	Mode      string // text or image
	Text      string
	FontSize  float64
	Color     string // R,G,B format
	Opacity   float64
	Rotation  float64
	Position  string // center, top-left, top-right, bottom-left, bottom-right, tile
	ImagePath string
	ImageMode string
}
