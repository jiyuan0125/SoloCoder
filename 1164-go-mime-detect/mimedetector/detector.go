package mimedetector

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	MaxHeaderSize = 512
)

func DetectFile(filePath string) (*DetectionResult, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file: %s", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	header := make([]byte, MaxHeaderSize)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return nil, err
	}
	header = header[:n]

	extMIME := detectByExtension(filePath)
	magicMIME := detectByMagic(header)

	finalMIME, warning, confidence := combineResults(extMIME, magicMIME, filePath, header)

	return &DetectionResult{
		ExtensionMIME: extMIME,
		MagicMIME:     magicMIME,
		FinalMIME:     finalMIME,
		Warning:       warning,
		Confidence:    confidence,
	}, nil
}

func detectByExtension(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == "" {
		return "application/octet-stream"
	}

	for _, ft := range SupportedFileTypes {
		for _, e := range ft.Extensions {
			if e == ext {
				return ft.MIMEType
			}
		}
	}

	return "application/octet-stream"
}

func detectByMagic(header []byte) string {
	if len(header) == 0 {
		return "application/octet-stream"
	}

	for _, ft := range SupportedFileTypes {
		for _, magic := range ft.MagicNumbers {
			if len(header) >= len(magic) && bytes.Equal(header[:len(magic)], magic) {
				if ft.MIMEType == "image/webp" {
					if len(header) >= 12 && bytes.Equal(header[8:12], []byte("WEBP")) {
						return ft.MIMEType
					}
					continue
				}
				if ft.MIMEType == "video/x-msvideo" {
					if len(header) >= 12 && bytes.Equal(header[8:12], []byte("AVI ")) {
						return ft.MIMEType
					}
					continue
				}
				return ft.MIMEType
			}
		}
	}

	return "application/octet-stream"
}

func combineResults(extMIME, magicMIME, filePath string, header []byte) (string, string, string) {
	if magicMIME == "application/zip" {
		specificMIME := checkZipContainer(filePath)
		if specificMIME != "" {
			magicMIME = specificMIME
		}
	}

	if magicMIME == "application/octet-stream" {
		if extMIME != "application/octet-stream" {
			return extMIME, "无法检测到魔数，使用扩展名推断", "medium"
		}
		return "application/octet-stream", "", "low"
	}

	if extMIME == "application/octet-stream" {
		return magicMIME, "无扩展名或未知扩展名，使用魔数检测", "high"
	}

	if extMIME == magicMIME {
		return extMIME, "", "high"
	}

	return magicMIME, fmt.Sprintf("扩展名与魔数不匹配：扩展名推断为%s，魔数检测为%s，以魔数为准", extMIME, magicMIME), "high"
}

func checkZipContainer(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	header := make([]byte, MaxHeaderSize)
	_, err = file.Read(header)
	if err != nil {
		return ""
	}

	signatures := []struct {
		offset  int
		signature []byte
		mime    string
	}{
		{30, []byte("word/"), "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{30, []byte("xl/"), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		{30, []byte("ppt/"), "application/vnd.openxmlformats-officedocument.presentationml.presentation"},
		{30, []byte("META-INF/MANIFEST.MF"), "application/java-archive"},
		{30, []byte("AndroidManifest.xml"), "application/vnd.android.package-archive"},
	}

	for _, sig := range signatures {
		if len(header) >= sig.offset+len(sig.signature) {
			if bytes.Contains(header[sig.offset:], sig.signature) {
				return sig.mime
			}
		}
	}

	return ""
}

func (r *DetectionResult) IsConsistent() bool {
	return r.ExtensionMIME == r.MagicMIME && r.MagicMIME != "application/octet-stream"
}

func (r *DetectionResult) HasWarning() bool {
	return r.Warning != ""
}
