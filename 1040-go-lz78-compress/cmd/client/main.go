package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"lz78-tool/pkg/api"
)

const (
	defaultServerURL = "http://localhost:8310"
)

type Client struct {
	serverURL string
	httpClient *http.Client
}

func NewClient(serverURL string) *Client {
	return &Client{
		serverURL:  strings.TrimRight(serverURL, "/"),
		httpClient: &http.Client{},
	}
}

func (c *Client) Compress(text string) ([]int, error) {
	reqBody := api.CompressRequest{
		Text: text,
	}
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.serverURL+"/api/compress", "application/json", bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result api.CompressResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("server error: %s", result.Error)
	}

	return result.Indexes, nil
}

func (c *Client) Decompress(indexes []int) (string, error) {
	reqBody := api.DecompressRequest{
		Indexes: indexes,
	}
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.serverURL+"/api/decompress", "application/json", bytes.NewBuffer(reqJSON))
	if err != nil {
		return "", fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result api.DecompressResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		return "", fmt.Errorf("server error: %s", result.Error)
	}

	return result.Text, nil
}

func (c *Client) GetDictStatus() (*api.DictStatusResponse, error) {
	resp, err := c.httpClient.Get(c.serverURL + "/api/dict")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result api.DictStatusResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("server error: %s", result.Error)
	}

	return &result, nil
}

func compressFile(client *Client, inputPath, outputPath string) error {
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	originalSize := len(content)

	indexes, err := client.Compress(string(content))
	if err != nil {
		return fmt.Errorf("compression failed: %w", err)
	}

	compressedData, err := json.Marshal(indexes)
	if err != nil {
		return fmt.Errorf("failed to marshal compressed data: %w", err)
	}

	if outputPath == "" {
		outputPath = inputPath + ".lz78"
	}

	if err := os.WriteFile(outputPath, compressedData, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	compressedSize := len(compressedData)

	fmt.Printf("Compression complete!\n")
	fmt.Printf("  Original size: %d bytes\n", originalSize)
	fmt.Printf("  Compressed size: %d bytes\n", compressedSize)
	fmt.Printf("  Compression ratio: %.2f%%\n", float64(compressedSize)/float64(originalSize)*100)
	fmt.Printf("  Output saved to: %s\n", outputPath)

	return nil
}

func decompressFile(client *Client, inputPath, outputPath string) error {
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	compressedSize := len(content)

	var indexes []int
	if err := json.Unmarshal(content, &indexes); err != nil {
		return fmt.Errorf("failed to parse compressed file: %w", err)
	}

	text, err := client.Decompress(indexes)
	if err != nil {
		return fmt.Errorf("decompression failed: %w", err)
	}

	if outputPath == "" {
		outputPath = strings.TrimSuffix(inputPath, filepath.Ext(inputPath))
		if outputPath == inputPath {
			outputPath = inputPath + ".decompressed"
		}
	}

	if err := os.WriteFile(outputPath, []byte(text), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	decompressedSize := len(text)

	fmt.Printf("Decompression complete!\n")
	fmt.Printf("  Compressed size: %d bytes\n", compressedSize)
	fmt.Printf("  Decompressed size: %d bytes\n", decompressedSize)
	fmt.Printf("  Output saved to: %s\n", outputPath)

	return nil
}

func showDictStatus(client *Client, showEntries bool) error {
	status, err := client.GetDictStatus()
	if err != nil {
		return fmt.Errorf("failed to get dictionary status: %w", err)
	}

	fmt.Printf("Dictionary Status:\n")
	fmt.Printf("  Total entries: %d\n", status.Count)

	if showEntries {
		fmt.Printf("\nDictionary Entries:\n")
		for i := 0; i < status.Count; i++ {
			if entry, ok := status.Entries[i]; ok {
				display := entry
				if entry == "" {
					display = "(empty string)"
				}
				fmt.Printf("  %4d: %q\n", i, display)
			}
		}
	}

	return nil
}

func printUsage() {
	fmt.Println(`LZ78 Compression Tool - Command Line Client

Usage:
  client <command> [options]

Commands:
  compress <input_file> [output_file]  Compress a file
  decompress <input_file> [output_file]  Decompress a file
  dict [--show-entries]                 Show dictionary status

Options:
  --server <url>  Server URL (default: http://localhost:8310)
  --help          Show this help message

Examples:
  client compress document.txt
  client compress document.txt document.txt.lz78
  client decompress document.txt.lz78
  client decompress document.txt.lz78 document_restored.txt
  client dict
  client dict --show-entries
  client --server http://localhost:8310 compress document.txt`)
}

func main() {
	serverURL := defaultServerURL

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	args := os.Args[1:]

	for i := 0; i < len(args); i++ {
		if args[i] == "--help" || args[i] == "-h" {
			printUsage()
			os.Exit(0)
		}
		if args[i] == "--server" && i+1 < len(args) {
			serverURL = args[i+1]
			args = append(args[:i], args[i+2:]...)
			i--
		}
	}

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]
	client := NewClient(serverURL)

	var err error
	switch command {
	case "compress":
		if len(args) < 2 {
			fmt.Println("Error: compress command requires input file")
			printUsage()
			os.Exit(1)
		}
		inputPath := args[1]
		outputPath := ""
		if len(args) >= 3 {
			outputPath = args[2]
		}
		err = compressFile(client, inputPath, outputPath)

	case "decompress":
		if len(args) < 2 {
			fmt.Println("Error: decompress command requires input file")
			printUsage()
			os.Exit(1)
		}
		inputPath := args[1]
		outputPath := ""
		if len(args) >= 3 {
			outputPath = args[2]
		}
		err = decompressFile(client, inputPath, outputPath)

	case "dict":
		showEntries := false
		for _, arg := range args[1:] {
			if arg == "--show-entries" {
				showEntries = true
			}
		}
		err = showDictStatus(client, showEntries)

	default:
		fmt.Printf("Error: unknown command '%s'\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
