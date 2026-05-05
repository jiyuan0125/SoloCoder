package indexer

import (
	"os"
	"path/filepath"
	"strings"
)

var binaryMagicNumbers = map[string][]byte{
	"png":   {0x89, 0x50, 0x4E, 0x47},
	"jpg":   {0xFF, 0xD8, 0xFF},
	"jpeg":  {0xFF, 0xD8, 0xFF},
	"gif":   {0x47, 0x49, 0x46, 0x38},
	"bmp":   {0x42, 0x4D},
	"pdf":   {0x25, 0x50, 0x44, 0x46},
	"zip":   {0x50, 0x4B, 0x03, 0x04},
	"rar":   {0x52, 0x61, 0x72, 0x21},
	"exe":   {0x4D, 0x5A},
	"dll":   {0x4D, 0x5A},
	"class": {0xCA, 0xFE, 0xBA, 0xBE},
	"jar":   {0x50, 0x4B, 0x03, 0x04},
	"war":   {0x50, 0x4B, 0x03, 0x04},
	"ear":   {0x50, 0x4B, 0x03, 0x04},
	"tar":   {0x75, 0x73, 0x74, 0x61, 0x72},
	"gz":    {0x1F, 0x8B},
	"bz2":   {0x42, 0x5A, 0x68},
	"7z":    {0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C},
	"iso":   {0x43, 0x44, 0x30, 0x30, 0x31},
	"mp3":   {0x49, 0x44, 0x33},
	"ogg":   {0x4F, 0x67, 0x67, 0x53},
	"wav":   {0x52, 0x49, 0x46, 0x46},
	"avi":   {0x52, 0x49, 0x46, 0x46},
	"mp4":   {0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70, 0x6D, 0x70, 0x34, 0x32},
	"mov":   {0x6D, 0x6F, 0x6F, 0x76},
	"mkv":   {0x1A, 0x45, 0xDF, 0xA3},
	"webm":  {0x1A, 0x45, 0xDF, 0xA3},
	"wmv":   {0x30, 0x26, 0xB2, 0x75, 0x8E, 0x66, 0xCF, 0x11, 0xA6, 0xD9, 0x00, 0xAA, 0x00, 0x62, 0xCE, 0x6C},
}

var excludedExtensions = map[string]bool{
	".png":   true,
	".jpg":   true,
	".jpeg":  true,
	".gif":   true,
	".bmp":   true,
	".pdf":   true,
	".zip":   true,
	".rar":   true,
	".exe":   true,
	".dll":   true,
	".class": true,
	".jar":   true,
	".war":   true,
	".ear":   true,
	".tar":   true,
	".gz":    true,
	".bz2":   true,
	".7z":    true,
	".iso":   true,
	".mp3":   true,
	".ogg":   true,
	".wav":   true,
	".avi":   true,
	".mp4":   true,
	".mov":   true,
	".mkv":   true,
	".webm":  true,
	".wmv":   true,
	".swf":   true,
	".obj":   true,
	".o":     true,
	".a":     true,
	".so":    true,
	".lib":   true,
	".pyc":   true,
	".pyo":   true,
	".pyd":   true,
}

var excludedDirs = map[string]bool{
	".git":         true,
	".svn":         true,
	".hg":          true,
	"node_modules": true,
	"vendor":       true,
	"target":       true,
	"__pycache__":  true,
	".idea":        true,
	".vscode":      true,
	"bin":          true,
	"obj":          true,
}

func IsBinaryFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	if excludedExtensions[ext] {
		return true
	}

	file, err := os.Open(filePath)
	if err != nil {
		return true
	}
	defer file.Close()

	header := make([]byte, 100)
	n, err := file.Read(header)
	if err != nil || n == 0 {
		return true
	}
	header = header[:n]

	for _, magic := range binaryMagicNumbers {
		if len(magic) <= len(header) && hasPrefix(header, magic) {
			return true
		}
	}

	return false
}

func hasPrefix(data, prefix []byte) bool {
	if len(prefix) > len(data) {
		return false
	}
	for i := range prefix {
		if data[i] != prefix[i] {
			return false
		}
	}
	return true
}

func isExcluded(filePath string) bool {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return false
	}

	parts := strings.Split(absPath, string(filepath.Separator))
	for _, part := range parts {
		if excludedDirs[part] {
			return true
		}
	}

	return false
}
