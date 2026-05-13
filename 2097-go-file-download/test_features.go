//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

func main() {
	client := &http.Client{Timeout: 10 * time.Second}

	fmt.Println("=== Test 1: Upload with path traversal filename (../etc/passwd) ===")
	{
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "../etc/passwd")
		io.WriteString(part, "test content")
		writer.WriteField("expire_hours", "24")
		writer.Close()

		req, _ := http.NewRequest("POST", "http://localhost:8888/api/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("X-User-ID", "user1")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			b, _ := io.ReadAll(resp.Body)
			fmt.Printf("Status: %d, Response: %s\n", resp.StatusCode, string(b))
			resp.Body.Close()
		}
	}

	fmt.Println("\n=== Test 2: Upload with / in filename ===")
	{
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "subdir/file.txt")
		io.WriteString(part, "test content")
		writer.WriteField("expire_hours", "24")
		writer.Close()

		req, _ := http.NewRequest("POST", "http://localhost:8888/api/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("X-User-ID", "user1")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			b, _ := io.ReadAll(resp.Body)
			fmt.Printf("Status: %d, Response: %s\n", resp.StatusCode, string(b))
			resp.Body.Close()
		}
	}

	fmt.Println("\n=== Test 3: Admin list all files ===")
	{
		req, _ := http.NewRequest("GET", "http://localhost:8888/api/admin/files", nil)
		req.Header.Set("X-User-ID", "admin")
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			b, _ := io.ReadAll(resp.Body)
			fmt.Printf("Status: %d, Body length: %d\n", resp.StatusCode, len(b))
			resp.Body.Close()
		}
	}

	fmt.Println("\n=== Test 4: Non-admin accessing admin endpoint (should be 403) ===")
	{
		req, _ := http.NewRequest("GET", "http://localhost:8888/api/admin/files", nil)
		req.Header.Set("X-User-ID", "user2")
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			b, _ := io.ReadAll(resp.Body)
			fmt.Printf("Status: %d, Response: %s\n", resp.StatusCode, string(b))
			resp.Body.Close()
		}
	}

	fmt.Println("\n=== Test 5: Download by path ===")
	{
		req, _ := http.NewRequest("GET", "http://localhost:8888/api/files/testfile2.txt", nil)
		req.Header.Set("X-User-ID", "user1")
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			b, _ := io.ReadAll(resp.Body)
			fmt.Printf("Status: %d, Body: %s\n", resp.StatusCode, strings.TrimSpace(string(b)))
			resp.Body.Close()
		}
	}

	fmt.Println("\n=== Test 6: Search with no results ===")
	{
		req, _ := http.NewRequest("GET", "http://localhost:8888/api/files?q=nonexistentxyz", nil)
		req.Header.Set("X-User-ID", "user1")
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			b, _ := io.ReadAll(resp.Body)
			fmt.Printf("Status: %d, Response: %s\n", resp.StatusCode, string(b))
			resp.Body.Close()
		}
	}

	fmt.Println("\n=== Test 7: Admin download records ===")
	{
		req, _ := http.NewRequest("GET", "http://localhost:8888/api/admin/downloads", nil)
		req.Header.Set("X-User-ID", "admin")
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			b, _ := io.ReadAll(resp.Body)
			fmt.Printf("Status: %d, Body length: %d\n", resp.StatusCode, len(b))
			resp.Body.Close()
		}
	}
}
