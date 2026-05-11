package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gotestgen/api"
)

type Client struct {
	serverURL string
	client    *http.Client
}

func NewClient(serverURL string) *Client {
	return &Client{
		serverURL: strings.TrimSuffix(serverURL, "/"),
		client:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) GenerateTestCode(fileName string, sourceCode string) (*api.GenerateResponse, error) {
	req := api.GenerateRequest{
		FileName:   fileName,
		SourceCode: sourceCode,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(c.serverURL+"/api/generate", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var result api.GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) PreviewTestCode(fileName string, sourceCode string) (*api.PreviewResponse, error) {
	req := api.PreviewRequest{
		FileName:   fileName,
		SourceCode: sourceCode,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(c.serverURL+"/api/preview", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var result api.PreviewResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func readFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func confirmOverwrite(filePath string) bool {
	fmt.Printf("File %s already exists. Overwrite? [y/N]: ", filePath)
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func writeFile(filePath string, content string, force bool) error {
	if _, err := os.Stat(filePath); err == nil {
		if !force {
			if !confirmOverwrite(filePath) {
				fmt.Printf("Skipped: %s\n", filePath)
				return nil
			}
		}
	}

	dir := filepath.Dir(filePath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Generated: %s\n", filePath)
	return nil
}

func printPreview(preview *api.PreviewResponse) {
	fmt.Println("=== Preview ===")
	fmt.Printf("Target file: %s\n", preview.FileName)
	fmt.Printf("Structs found: %v\n", preview.Structs)
	fmt.Printf("Test functions to be generated:\n")
	for _, tf := range preview.TestFuncs {
		fmt.Printf("  - %s (receiver: %s, cases: %d)\n", tf.Name, tf.Receiver, len(tf.Cases))
		for _, tc := range tf.Cases {
			fmt.Printf("    * %s\n", tc.Name)
		}
	}
	fmt.Println("===============")
}

func processFile(client *Client, filePath string, printOnly bool, force bool) error {
	content, err := readFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	fileName := filepath.Base(filePath)
	resp, err := client.GenerateTestCode(fileName, content)
	if err != nil {
		return fmt.Errorf("failed to generate test code: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("server error: %s", resp.Error)
	}

	if printOnly {
		fmt.Printf("=== %s ===\n", resp.FileName)
		fmt.Println(resp.Code)
		return nil
	}

	testFilePath := strings.TrimSuffix(filePath, ".go") + "_test.go"
	return writeFile(testFilePath, resp.Code, force)
}

func processFiles(client *Client, filePaths []string, printOnly bool, force bool) {
	for _, filePath := range filePaths {
		fmt.Printf("Processing: %s\n", filePath)
		if err := processFile(client, filePath, printOnly, force); err != nil {
			fmt.Printf("Error processing %s: %v\n", filePath, err)
		}
	}
}

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	printOnly := flag.Bool("print", false, "Print generated code to stdout instead of writing to files")
	force := flag.Bool("force", false, "Force overwrite existing test files")
	preview := flag.Bool("preview", false, "Preview test functions without generating code")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <file1.go> [file2.go ...]\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s main.go\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "  %s -print main.go\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "  %s -preview main.go\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "  %s -force *.go\n", filepath.Base(os.Args[0]))
	}

	flag.Parse()

	files := flag.Args()
	if len(files) == 0 {
		fmt.Println("Error: No files specified")
		flag.Usage()
		os.Exit(1)
	}

	client := NewClient(*serverURL)

	if *preview {
		for _, filePath := range files {
			content, err := readFile(filePath)
			if err != nil {
				fmt.Printf("Error reading file %s: %v\n", filePath, err)
				continue
			}

			fileName := filepath.Base(filePath)
			resp, err := client.PreviewTestCode(fileName, content)
			if err != nil {
				fmt.Printf("Error previewing %s: %v\n", filePath, err)
				continue
			}

			if !resp.Success {
				fmt.Printf("Server error for %s: %s\n", filePath, resp.Error)
				continue
			}

			printPreview(resp)
		}
		return
	}

	processFiles(client, files, *printOnly, *force)
}
