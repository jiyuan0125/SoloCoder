package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/exif-reader/pkg/api"
)

const baseURL = "http://localhost:8203"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "parse":
		if len(os.Args) < 3 {
			fmt.Println("Error: missing image file path")
			printUsage()
			os.Exit(1)
		}
		handleParse(os.Args[2])
	case "help":
		printUsage()
	default:
		fmt.Printf("Error: unknown command '%s'\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("EXIF Reader Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client parse <image-file>   Parse EXIF data from an image file")
	fmt.Println("  client help                 Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  client parse photo.jpg")
	fmt.Println("  client parse scan.tiff")
}

func handleParse(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error: failed to open file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		fmt.Printf("Error: failed to create form file: %v\n", err)
		os.Exit(1)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Printf("Error: failed to copy file content: %v\n", err)
		os.Exit(1)
	}

	writer.Close()

	req, err := http.NewRequest("POST", baseURL+"/parse", body)
	if err != nil {
		fmt.Printf("Error: failed to create request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var errResp api.ErrorResponse
		json.Unmarshal(body, &errResp)
		fmt.Printf("Error: server returned %d - %s\n", resp.StatusCode, errResp.Error)
		os.Exit(1)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: failed to read response: %v\n", err)
		os.Exit(1)
	}

	var parseResp api.ParseResponse
	if err := json.Unmarshal(bodyBytes, &parseResp); err != nil {
		fmt.Printf("Error: failed to parse response: %v\n", err)
		os.Exit(1)
	}

	sessionID := resp.Header.Get("X-Session-Id")

	printFormattedEXIF(&parseResp, sessionID)
}

func printFormattedEXIF(resp *api.ParseResponse, sessionID string) {
	fmt.Println("========================================")
	fmt.Println("          EXIF Metadata")
	fmt.Println("========================================")
	fmt.Println()

	if len(resp.Metadata) == 0 {
		fmt.Println("No EXIF metadata found in the image.")
	} else {
		for _, key := range []string{
			api.FieldDateTime,
			api.FieldExposureTime,
			api.FieldFNumber,
			api.FieldISO,
			api.FieldGPSLatitude,
			api.FieldGPSLongitude,
			api.FieldGPSAltitude,
		} {
			if value, ok := resp.Metadata[key]; ok {
				printField(key, value)
			}
		}

		fmt.Println()
		fmt.Println("--- Raw Metadata ---")
		for k, v := range resp.Metadata {
			if !isStandardField(k) {
				printField(k, v)
			}
		}
	}

	fmt.Println()
	if resp.HasThumbnail {
		fmt.Println("[✓] Thumbnail available")
	} else {
		fmt.Println("[ ] No thumbnail embedded")
	}

	if sessionID != "" {
		fmt.Printf("Session ID: %s\n", sessionID)
	}
	fmt.Println("========================================")
}

func isStandardField(key string) bool {
	standardFields := []string{
		api.FieldDateTime,
		api.FieldExposureTime,
		api.FieldFNumber,
		api.FieldISO,
		api.FieldGPSLatitude,
		api.FieldGPSLongitude,
		api.FieldGPSAltitude,
	}
	for _, f := range standardFields {
		if f == key {
			return true
		}
	}
	return false
}

func printField(key string, value interface{}) {
	label := formatLabel(key)
	fmt.Printf("%-20s %v\n", label+":", value)
}

func formatLabel(key string) string {
	labels := map[string]string{
		api.FieldDateTime:     "拍摄时间",
		api.FieldExposureTime: "快门速度",
		api.FieldFNumber:      "光圈值",
		api.FieldISO:          "ISO感光度",
		api.FieldGPSLatitude:  "GPS纬度",
		api.FieldGPSLongitude: "GPS经度",
		api.FieldGPSAltitude:  "GPS海拔",
	}

	if label, ok := labels[key]; ok {
		return label
	}

	words := splitCamelCase(key)
	return strings.Join(words, " ")
}

func splitCamelCase(s string) []string {
	var words []string
	var current strings.Builder

	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			if current.Len() > 0 {
				words = append(words, current.String())
			}
			current.Reset()
		}
		current.WriteRune(r)
	}

	if current.Len() > 0 {
		words = append(words, current.String())
	}

	return words
}
