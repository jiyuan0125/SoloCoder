package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"godoc-generator/internal/api"
)

type Client struct {
	serverURL string
	httpClient *http.Client
}

func NewClient(serverURL string) *Client {
	return &Client{
		serverURL: strings.TrimSuffix(serverURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) UploadFiles(files []string, includeUnexported bool) (*api.UploadResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("include_unexported", fmt.Sprintf("%t", includeUnexported))

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file, err)
		}

		part, err := writer.CreateFormFile("files", filepath.Base(file))
		if err != nil {
			return nil, fmt.Errorf("failed to create form file: %w", err)
		}
		_, err = part.Write(content)
		if err != nil {
			return nil, fmt.Errorf("failed to write file content: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.serverURL+"/api/upload", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload files: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var uploadResp api.UploadResponse
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &uploadResp, nil
}

func (c *Client) Export(format string, outputFile string) error {
	url := fmt.Sprintf("%s/api/export?format=%s", c.serverURL, format)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to export: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("export failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if outputFile != "" {
		if err := os.WriteFile(outputFile, content, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		log.Printf("Exported to %s", outputFile)
	} else {
		fmt.Println(string(content))
	}

	return nil
}

func (c *Client) Search(query string, includeUnexported bool) (*api.SearchResponse, error) {
	url := fmt.Sprintf("%s/api/search?q=%s&include_unexported=%t",
		c.serverURL, query, includeUnexported)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var searchResp api.SearchResponse
	if err := json.Unmarshal(respBody, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &searchResp, nil
}

func collectGoFiles(dir string) ([]string, error) {
	var files []string

	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		if strings.HasSuffix(dir, ".go") && !strings.HasSuffix(filepath.Base(dir), "_test.go") {
			return []string{dir}, nil
		}
		return nil, fmt.Errorf("not a Go file: %s", dir)
	}

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(filepath.Base(path), "_test.go") {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

func getFileModTimes(files []string) map[string]time.Time {
	times := make(map[string]time.Time)
	for _, file := range files {
		if info, err := os.Stat(file); err == nil {
			times[file] = info.ModTime()
		}
	}
	return times
}

func filesChanged(oldTimes, newTimes map[string]time.Time) bool {
	if len(oldTimes) != len(newTimes) {
		return true
	}
	for file, newTime := range newTimes {
		oldTime, exists := oldTimes[file]
		if !exists || !oldTime.Equal(newTime) {
			return true
		}
	}
	return false
}

func watchMode(client *Client, dir string, includeUnexported bool, format string, outputFile string) error {
	log.Printf("Watching directory: %s", dir)
	log.Printf("Polling every 2 seconds...")

	files, err := collectGoFiles(dir)
	if err != nil {
		return fmt.Errorf("failed to collect files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no Go files found in directory")
	}

	log.Printf("Found %d Go files", len(files))

	modTimes := getFileModTimes(files)

	uploadResp, err := client.UploadFiles(files, includeUnexported)
	if err != nil {
		return fmt.Errorf("initial upload failed: %w", err)
	}
	log.Printf("Initial upload successful: %s", uploadResp.Message)

	if err := client.Export(format, outputFile); err != nil {
		log.Printf("Initial export failed: %v", err)
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		newFiles, err := collectGoFiles(dir)
		if err != nil {
			log.Printf("Warning: failed to collect files: %v", err)
			continue
		}

		newModTimes := getFileModTimes(newFiles)
		if filesChanged(modTimes, newModTimes) {
			log.Printf("Changes detected, regenerating...")

			uploadResp, err := client.UploadFiles(newFiles, includeUnexported)
			if err != nil {
				log.Printf("Upload failed: %v", err)
				continue
			}
			log.Printf("Upload successful: %s", uploadResp.Message)

			if err := client.Export(format, outputFile); err != nil {
				log.Printf("Export failed: %v", err)
				continue
			}

			modTimes = newModTimes
			files = newFiles
			log.Printf("Documentation regenerated successfully")
		}
	}

	return nil
}

func printUsage() {
	fmt.Println(`Go Doc Generator Client

Usage:
  godoc-client [flags] <directory-or-file>

Flags:
  -server      Server URL (default: http://localhost:8080)
  -format      Output format: json or markdown (default: json)
  -output      Output file (default: stdout)
  -include-unexported  Include unexported declarations
  -watch       Watch mode: monitor files for changes
  -search      Search query (search mode)

Examples:
  # Generate documentation for a package
  godoc-client ./mypackage -format markdown -output docs.md

  # Watch mode
  godoc-client ./mypackage -watch -format markdown -output docs.md

  # Search documentation
  godoc-client -search "NewClient"

  # Include unexported declarations
  godoc-client ./mypackage -include-unexported
`)
}

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	format := flag.String("format", "json", "Output format: json or markdown")
	output := flag.String("output", "", "Output file")
	includeUnexported := flag.Bool("include-unexported", false, "Include unexported declarations")
	watch := flag.Bool("watch", false, "Watch mode")
	search := flag.String("search", "", "Search query")

	flag.Usage = printUsage
	flag.Parse()

	client := NewClient(*serverURL)

	if *search != "" {
		resp, err := client.Search(*search, *includeUnexported)
		if err != nil {
			log.Fatalf("Search failed: %v", err)
		}

		fmt.Printf("Found %d results:\n\n", len(resp.Results))
		for i, result := range resp.Results {
			fmt.Printf("%d. [%s] %s\n", i+1, result.Kind, result.Name)
			if result.Summary != "" {
				fmt.Printf("   %s\n", result.Summary)
			}
			if !result.Exported {
				fmt.Printf("   [Unexported]\n")
			}
			fmt.Println()
		}
		return
	}

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	target := args[0]

	files, err := collectGoFiles(target)
	if err != nil {
		log.Fatalf("Failed to collect files: %v", err)
	}

	if len(files) == 0 {
		log.Fatalf("No Go files found in: %s", target)
	}

	log.Printf("Found %d Go files", len(files))

	if *watch {
		if err := watchMode(client, target, *includeUnexported, *format, *output); err != nil {
			log.Fatalf("Watch mode error: %v", err)
		}
		return
	}

	uploadResp, err := client.UploadFiles(files, *includeUnexported)
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}
	log.Printf("Upload successful: %s", uploadResp.Message)

	if err := client.Export(*format, *output); err != nil {
		log.Fatalf("Export failed: %v", err)
	}
}
