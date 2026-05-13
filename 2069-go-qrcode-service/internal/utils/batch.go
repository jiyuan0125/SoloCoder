package utils

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

func ParseCSV(data []byte) ([]string, []int, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1
	
	var lines []string
	var skipped []int
	lineNum := 0
	
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		
		lineNum++
		
		if len(record) == 0 {
			skipped = append(skipped, lineNum)
			continue
		}
		
		content := strings.TrimSpace(record[0])
		if content == "" {
			skipped = append(skipped, lineNum)
			continue
		}
		
		lines = append(lines, content)
	}
	
	return lines, skipped, nil
}

func CreateZip(files map[string][]byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	
	for name, data := range files {
		f, err := writer.Create(name)
		if err != nil {
			return nil, err
		}
		if _, err := f.Write(data); err != nil {
			return nil, err
		}
	}
	
	if err := writer.Close(); err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

func SanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "*", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "_")
	name = strings.ReplaceAll(name, "|", "_")
	if len(name) > 50 {
		name = name[:50]
	}
	return name
}

func FormatFilename(index int, content, ext string) string {
	sanitized := SanitizeFilename(content)
	if sanitized == "" {
		sanitized = "qrcode"
	}
	return fmt.Sprintf("%04d_%s.%s", index+1, sanitized, strings.ToLower(ext))
}
