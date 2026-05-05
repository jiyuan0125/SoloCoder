// Package main implements the CSV join client.
// This is a command-line tool that calls the CSV join server via HTTP.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"csvjoin/pkg/api"
)

// Client represents the CSV join HTTP client.
type Client struct {
	serverURL string
}

// NewClient creates a new CSV join client.
func NewClient(serverURL string) *Client {
	return &Client{serverURL: strings.TrimRight(serverURL, "/")}
}

// HealthCheck checks if the server is running.
func (c *Client) HealthCheck() error {
	resp, err := http.Get(c.serverURL + "/health")
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %s", resp.Status)
	}

	var healthResp api.HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		return fmt.Errorf("failed to parse health response: %w", err)
	}

	if healthResp.Status != "ok" {
		return fmt.Errorf("server not healthy: %s", healthResp.Status)
	}

	return nil
}

// UploadFile uploads a file to the server.
func (c *Client) UploadFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("files", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.serverURL+"/upload", body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed: %s - %s", resp.Status, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	files, ok := result["files"].([]interface{})
	if !ok || len(files) == 0 {
		return "", fmt.Errorf("no files returned in response")
	}

	uploadedPath, ok := files[0].(string)
	if !ok {
		return "", fmt.Errorf("invalid file path in response")
	}

	return uploadedPath, nil
}

// Join sends a join request to the server.
func (c *Client) Join(req api.JoinRequest) (*api.JoinResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.serverURL+"/join", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var joinResp api.JoinResponse
	if err := json.NewDecoder(resp.Body).Decode(&joinResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &joinResp, nil
}

// DownloadFile downloads a file from the server.
func (c *Client) DownloadFile(serverPath, localPath string) error {
	filename := filepath.Base(serverPath)
	downloadURL := fmt.Sprintf("%s/download?file=%s", c.serverURL, filename)

	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed: %s - %s", resp.Status, string(bodyBytes))
	}

	outFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

func printUsage() {
	fmt.Println("CSV Join Client - Command-line tool for joining CSV files")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  csvjoin-client [options] <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  health    - Check if the server is running")
	fmt.Println("  join      - Join CSV files")
	fmt.Println("  upload    - Upload CSV files to server")
	fmt.Println("  download  - Download result file from server")
	fmt.Println()
	fmt.Println("Global Options:")
	fmt.Println("  -server <url>   - Server URL (default: http://localhost:8080)")
	fmt.Println()
	fmt.Println("Join Options:")
	fmt.Println("  -key <column>        - Join key column name (required)")
	fmt.Println("  -type <inner|left>   - Join type: inner or left (default: left)")
	fmt.Println("  -notrim              - Do not trim whitespace from join keys")
	fmt.Println("  -output <path>       - Output file path (default: output.csv)")
	fmt.Println("  -skip-header         - Treat first row as data (not header)")
	fmt.Println("  -columns <names>     - Custom column names (comma-separated, use with -skip-header)")
	fmt.Println("  -suffixes <suffixes> - File suffixes to distinguish columns (comma-separated)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  csvjoin-client health")
	fmt.Println("  csvjoin-client join -key id -type left file1.csv file2.csv")
	fmt.Println("  csvjoin-client join -key name -output result.csv a.csv b.csv c.csv")
	fmt.Println("  csvjoin-client upload file1.csv file2.csv")
	fmt.Println("  csvjoin-client download /uploads/result_xxx.csv ./local_output.csv")
	os.Exit(1)
}

func main() {
	// Parse global flags
	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	flag.Usage = printUsage

	// Custom flag parsing to handle subcommands
	flag.CommandLine.Init(os.Args[0], flag.ExitOnError)

	// Find first non-flag argument as command
	var command string
	var args []string

	for i := 1; i < len(os.Args); i++ {
		if strings.HasPrefix(os.Args[i], "-") {
			continue
		}
		command = os.Args[i]
		if i+1 < len(os.Args) {
			args = os.Args[i+1:]
		}
		break
	}

	if command == "" {
		printUsage()
	}

	// Create client
	client := NewClient(*serverURL)

	switch command {
	case "health":
		if err := client.HealthCheck(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Server is running")

	case "join":
		joinCmd := flag.NewFlagSet("join", flag.ExitOnError)
		joinKey := joinCmd.String("key", "", "Join key column name")
		joinType := joinCmd.String("type", "left", "Join type: inner or left")
		noTrim := joinCmd.Bool("notrim", false, "Do not trim whitespace from join keys")
		outputPath := joinCmd.String("output", "output.csv", "Output file path")
		skipHeader := joinCmd.Bool("skip-header", false, "Treat first row as data")
		columns := joinCmd.String("columns", "", "Custom column names (comma-separated)")
		suffixes := joinCmd.String("suffixes", "", "File suffixes (comma-separated)")

		joinCmd.Parse(args)
		files := joinCmd.Args()

		if *joinKey == "" {
			fmt.Fprintln(os.Stderr, "Error: -key is required")
			os.Exit(1)
		}

		if len(files) < 2 {
			fmt.Fprintln(os.Stderr, "Error: at least 2 files required")
			os.Exit(1)
		}

		// Parse column names if provided
		var columnNames []string
		if *columns != "" {
			columnNames = strings.Split(*columns, ",")
		}

		// Parse suffixes if provided
		var fileSuffixes []string
		if *suffixes != "" {
			fileSuffixes = strings.Split(*suffixes, ",")
		}

		// Prepare join request
		req := api.JoinRequest{
			JoinKey:  *joinKey,
			TrimKeys: !*noTrim,
		}

		switch *joinType {
		case "inner":
			req.JoinType = api.InnerJoin
		case "left":
			req.JoinType = api.LeftJoin
		default:
			fmt.Fprintf(os.Stderr, "Error: invalid join type '%s'\n", *joinType)
			os.Exit(1)
		}

		// Add files (use local paths directly, server will read them)
		for i, file := range files {
			fileConfig := api.FileConfig{
				Path:       file,
				SkipHeader: *skipHeader,
			}

			if *skipHeader && len(columnNames) > 0 {
				fileConfig.ColumnNames = columnNames
			}

			if len(fileSuffixes) > i {
				fileConfig.FileSuffix = fileSuffixes[i]
			}

			req.Files = append(req.Files, fileConfig)
		}

		fmt.Println("Sending join request to server...")
		resp, err := client.Join(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if !resp.Success {
			fmt.Fprintf(os.Stderr, "Join failed: %s\n", resp.Error)
			os.Exit(1)
		}

		fmt.Printf("Join successful!\n")
		fmt.Printf("  Output path on server: %s\n", resp.OutputPath)
		fmt.Printf("  Row count: %d\n", resp.RowCount)
		fmt.Printf("  Columns: %v\n", resp.Columns)

		// Download the result
		fmt.Printf("\nDownloading result to %s...\n", *outputPath)
		if err := client.DownloadFile(resp.OutputPath, *outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to download result: %v\n", err)
			fmt.Fprintf(os.Stderr, "The result is still available on the server at: %s\n", resp.OutputPath)
			os.Exit(1)
		}

		fmt.Printf("Result saved to: %s\n", *outputPath)

	case "upload":
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "Error: no files specified for upload")
			os.Exit(1)
		}

		for _, file := range args {
			fmt.Printf("Uploading %s...\n", file)
			path, err := client.UploadFile(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error uploading %s: %v\n", file, err)
				os.Exit(1)
			}
			fmt.Printf("  Uploaded to: %s\n", path)
		}

	case "download":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: download requires server path and local path")
			os.Exit(1)
		}

		serverPath := args[0]
		localPath := args[1]

		fmt.Printf("Downloading %s to %s...\n", serverPath, localPath)
		if err := client.DownloadFile(serverPath, localPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Download successful")

	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n", command)
		printUsage()
	}
}
