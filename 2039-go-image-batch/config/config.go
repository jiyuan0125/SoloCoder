package config

const (
	ServerPort         = "8080"
	DatabasePath       = "./data/image_batch.db"
	UploadDir          = "./uploads"
	OutputDir          = "./outputs"
	TempDir            = "./temp"
	MaxMemoryPerImage  = 200 * 1024 * 1024
	MaxUploadSize      = 500 * 1024 * 1024
	SupportedFormats   = "JPEG, PNG, GIF, BMP, WebP"
)
