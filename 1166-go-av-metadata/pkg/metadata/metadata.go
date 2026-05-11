package metadata

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type AudioMetadata struct {
	Format      string
	Duration    float64
	Bitrate     int
	FileSize    int64
	FormatTags  map[string]string
	AudioTags   map[string]string
	LastUpdated time.Time
}

type Parser interface {
	Parse(filePath string) (*AudioMetadata, error)
	Supports(extension string) bool
}

func ParseFile(filePath string) (*AudioMetadata, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("文件不存在: %v", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("路径是目录，不是文件")
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	ext = strings.TrimPrefix(ext, ".")

	switch ext {
	case "mp3":
		return parseMP3(filePath)
	case "mp4", "m4a", "m4v":
		return parseMP4(filePath)
	case "flac":
		return parseFLAC(filePath)
	case "wav", "wave":
		return parseWAV(filePath)
	case "avi":
		return parseAVI(filePath)
	case "mkv", "webm":
		return parseMKV(filePath)
	default:
		return parseGeneric(filePath, ext)
	}
}

func DetectFormat(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	ext = strings.TrimPrefix(ext, ".")

	switch ext {
	case "mp3":
		return "MP3"
	case "mp4", "m4a", "m4v":
		return "MP4"
	case "flac":
		return "FLAC"
	case "wav", "wave":
		return "WAV"
	case "avi":
		return "AVI"
	case "mkv", "webm":
		return "MKV"
	default:
		return strings.ToUpper(ext)
	}
}
