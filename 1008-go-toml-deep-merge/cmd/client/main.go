package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"tomlmerge/common"
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

func (c *Client) postJSON(path string, body interface{}, resp interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request failed: %w", err)
	}
	req, err := http.NewRequest("POST", c.serverURL+path, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return fmt.Errorf("server returned status %d: %s", httpResp.StatusCode, string(bodyBytes))
	}
	if resp != nil {
		return json.NewDecoder(httpResp.Body).Decode(resp)
	}
	return nil
}

func (c *Client) get(path string, resp interface{}) error {
	httpResp, err := c.httpClient.Get(c.serverURL + path)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return fmt.Errorf("server returned status %d: %s", httpResp.StatusCode, string(bodyBytes))
	}
	if resp != nil {
		return json.NewDecoder(httpResp.Body).Decode(resp)
	}
	return nil
}

func readFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file %s failed: %w", path, err)
	}
	return string(content), nil
}

func writeFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func main() {
	var serverURL string
	var baseFile string
	var envFile string
	var outputFile string
	var uploadFile string
	var listFiles bool
	var mergeMultiple string
	var env string
	flag.StringVar(&serverURL, "server", "http://localhost:8100", "server URL")
	flag.StringVar(&baseFile, "base", "", "base TOML file path")
	flag.StringVar(&envFile, "env", "", "environment TOML file path")
	flag.StringVar(&outputFile, "output", "", "output file path (default: stdout)")
	flag.StringVar(&uploadFile, "upload", "", "upload TOML file to server")
	flag.BoolVar(&listFiles, "list", false, "list uploaded files")
	flag.StringVar(&mergeMultiple, "merge", "", "merge multiple uploaded files (format: baseID:envID1,envID2,...)")
	flag.StringVar(&env, "environment", "", "environment tag for uploaded file")
	flag.Parse()
	client := NewClient(serverURL)
	if uploadFile != "" {
		content, err := readFile(uploadFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		var resp common.UploadResponse
		if err := client.postJSON("/api/upload", common.UploadRequest{
			FileName:    filepath.Base(uploadFile),
			Content:     content,
			Environment: env,
		}, &resp); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if !resp.Success {
			fmt.Fprintln(os.Stderr, "Error:", resp.Error)
			os.Exit(1)
		}
		fmt.Printf("Uploaded successfully, file ID: %s\n", resp.FileID)
		return
	}
	if listFiles {
		var resp common.ListFilesResponse
		if err := client.get("/api/files", &resp); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if !resp.Success {
			fmt.Fprintln(os.Stderr, "Error:", resp.Error)
			os.Exit(1)
		}
		if len(resp.Files) == 0 {
			fmt.Println("No files uploaded")
			return
		}
		for _, f := range resp.Files {
			fmt.Printf("ID: %s, Name: %s, Env: %s, Size: %d\n", f.ID, f.Name, f.Environment, f.Size)
		}
		return
	}
	if mergeMultiple != "" {
		parts := strings.SplitN(mergeMultiple, ":", 2)
		if len(parts) != 2 {
			fmt.Fprintln(os.Stderr, "Invalid merge format, use: baseID:envID1,envID2,...")
			os.Exit(1)
		}
		baseID := parts[0]
		envIDs := strings.Split(parts[1], ",")
		var resp common.MergeResponse
		if err := client.postJSON("/api/merge-multiple", common.MergeMultipleRequest{
			BaseID: baseID,
			EnvIDs: envIDs,
		}, &resp); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if !resp.Success {
			fmt.Fprintln(os.Stderr, "Error:", resp.Error)
			os.Exit(1)
		}
		if outputFile != "" {
			if err := writeFile(outputFile, resp.Result); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			fmt.Printf("Wrote to %s\n", outputFile)
		} else {
			fmt.Println(resp.Result)
		}
		return
	}
	if baseFile == "" || envFile == "" {
		flag.Usage()
		os.Exit(1)
	}
	baseContent, err := readFile(baseFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	envContent, err := readFile(envFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var resp common.MergeResponse
	if err := client.postJSON("/api/merge", common.MergeRequest{
		BaseContent: baseContent,
		EnvContent:  envContent,
	}, &resp); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if !resp.Success {
		fmt.Fprintln(os.Stderr, "Error:", resp.Error)
		os.Exit(1)
	}
	if outputFile != "" {
		if err := writeFile(outputFile, resp.Result); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("Wrote to %s\n", outputFile)
	} else {
		fmt.Println(resp.Result)
	}
}
