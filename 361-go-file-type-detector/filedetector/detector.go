package filedetector

import (
	"io"
	"os"
	"bytes"
)

type fileType struct {
	mime   string
	signature []byte
}

var fileTypes = []fileType{
	{"image/jpeg", []byte{0xFF, 0xD8, 0xFF}},
	{"image/png", []byte{0x89, 0x50, 0x4E, 0x47}},
	{"image/gif", []byte{0x47, 0x49, 0x46, 0x38}},
	{"application/pdf", []byte{0x25, 0x50, 0x44, 0x46}},
	{"application/zip", []byte{0x50, 0x4B, 0x03, 0x04}},
	{"application/vnd.rar", []byte{0x52, 0x61, 0x72, 0x21}},
}

func DetectFromBytes(data []byte) string {
	if len(data) == 0 {
		return "unknown"
	}

	for _, ft := range fileTypes {
		if len(data) >= len(ft.signature) && bytes.HasPrefix(data, ft.signature) {
			return ft.mime
		}
	}

	return "unknown"
}

func DetectFromPath(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return "", err
	}

	return DetectFromBytes(header[:n]), nil
}
