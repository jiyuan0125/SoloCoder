//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

func main() {
	client := &http.Client{Timeout: 10 * time.Second}

	fmt.Println("=== Upload file with max_downloads=1 ===")
	{
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "limited.txt")
		io.WriteString(part, "limited content")
		writer.WriteField("expire_hours", "24")
		writer.WriteField("max_downloads", "1")
		writer.Close()

		req, _ := http.NewRequest("POST", "http://localhost:8888/api/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("X-User-ID", "user1")

		resp, _ := client.Do(req)
		b, _ := io.ReadAll(resp.Body)
		fmt.Printf("Upload: %d %s\n", resp.StatusCode, string(b))
		resp.Body.Close()

		var result map[string]interface{}
		json.Unmarshal(b, &result)
		results := result["results"].([]interface{})
		first := results[0].(map[string]interface{})
		fileID := first["file_id"].(string)

		fmt.Println("\n=== Download 1 (should succeed) ===")
		req2, _ := http.NewRequest("GET", "http://localhost:8888/api/download/"+fileID, nil)
		resp2, _ := client.Do(req2)
		b2, _ := io.ReadAll(resp2.Body)
		fmt.Printf("Download 1: %d, Body: %s\n", resp2.StatusCode, string(b2))
		resp2.Body.Close()

		fmt.Println("\n=== Download 2 (should be 410) ===")
		req3, _ := http.NewRequest("GET", "http://localhost:8888/api/download/"+fileID, nil)
		resp3, _ := client.Do(req3)
		b3, _ := io.ReadAll(resp3.Body)
		fmt.Printf("Download 2: %d, Body: %s\n", resp3.StatusCode, string(b3))
		resp3.Body.Close()
	}
}
