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
	"time"

	"archiver/pkg/api"
)

type Client struct {
	config api.ClientConfig
}

func NewClient(config api.ClientConfig) *Client {
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Minute
	}
	return &Client{config: config}
}

func (c *Client) doRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	url := c.config.ServerURL + endpoint

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{
		Timeout: c.config.Timeout,
	}

	return client.Do(req)
}

func (c *Client) HealthCheck() (bool, error) {
	resp, err := c.doRequest(http.MethodGet, "/health", nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

func (c *Client) ListFiles(path string, patterns []string, recursive bool) (*api.ListFilesResponse, error) {
	req := api.ListFilesRequest{
		Path:      path,
		Patterns:  patterns,
		Recursive: recursive,
	}

	resp, err := c.doRequest(http.MethodPost, "/list", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result api.ListFilesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) Archive(req *api.ArchiveRequest, outputPath string) (*http.Response, error) {
	resp, err := c.doRequest(http.MethodPost, "/archive", req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	if outputPath != "" {
		out, err := os.Create(outputPath)
		if err != nil {
			resp.Body.Close()
			return nil, err
		}
		defer out.Close()

		if _, err := io.Copy(out, resp.Body); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()
	}

	return resp, nil
}

func main() {
	var (
		serverURL   = flag.String("server", "http://localhost:8430", "server URL")
		outputFile  = flag.String("output", "archive.tar.gz", "output file path")
		directories = flag.String("dirs", "", "comma-separated directories to archive")
		files       = flag.String("files", "", "comma-separated files to archive")
		include     = flag.String("include", "", "comma-separated include patterns")
		exclude     = flag.String("exclude", "", "comma-separated exclude patterns")
		recursive   = flag.Bool("recursive", true, "recursively include subdirectories")
		compression = flag.Int("compression", 6, "compression level (1-9, 0=no compression)")
		stripPrefix = flag.String("strip-prefix", "", "strip prefix from paths in archive")
		addPrefix   = flag.String("add-prefix", "", "add prefix to paths in archive")
		healthCheck = flag.Bool("health", false, "check server health")
		list        = flag.String("list", "", "list files in directory (format: path)")
	)

	flag.Parse()

	config := api.ClientConfig{
		ServerURL: *serverURL,
		Timeout:   5 * time.Minute,
	}

	client := NewClient(config)

	if *healthCheck {
		ok, err := client.HealthCheck()
		if err != nil {
			fmt.Fprintf(os.Stderr, "health check failed: %v\n", err)
			os.Exit(1)
		}
		if ok {
			fmt.Println("server is healthy")
			return
		}
		fmt.Println("server is not healthy")
		os.Exit(1)
	}

	if *list != "" {
		patterns := []string{}
		if *include != "" {
			patterns = strings.Split(*include, ",")
		}
		resp, err := client.ListFiles(*list, patterns, *recursive)
		if err != nil {
			fmt.Fprintf(os.Stderr, "list failed: %v\n", err)
			os.Exit(1)
		}
		if !resp.Success {
			fmt.Fprintf(os.Stderr, "server error: %s\n", resp.Error)
			os.Exit(1)
		}
		for _, f := range resp.Files {
			fmt.Printf("%s\t%d\t%s\n", f.Path, f.Size, f.ModTime.Format(time.RFC3339))
		}
		return
	}

	req := &api.ArchiveRequest{
		Compression:    *compression,
		StripPrefix:    *stripPrefix,
		AddPrefix:      *addPrefix,
		FollowSymlinks: false,
		IncludeEmpty:   false,
	}

	if *include != "" {
		req.IncludePattern = strings.Split(*include, ",")
	}
	if *exclude != "" {
		req.ExcludePattern = strings.Split(*exclude, ",")
	}

	allFiles := make(map[string]string)

	if *directories != "" {
		for _, dir := range strings.Split(*directories, ",") {
			dir = strings.TrimSpace(dir)
			if dir == "" {
				continue
			}
			req.Directories = append(req.Directories, dir)
		}
	}

	if *files != "" {
		for _, filePattern := range strings.Split(*files, ",") {
			filePattern = strings.TrimSpace(filePattern)
			if filePattern == "" {
				continue
			}

			absPath, err := filepath.Abs(filePattern)
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid path %s: %v\n", filePattern, err)
				continue
			}

			info, err := os.Stat(absPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "cannot stat %s: %v\n", filePattern, err)
				continue
			}

			if info.IsDir() {
				req.Directories = append(req.Directories, absPath)
			} else {
				allFiles[absPath] = filepath.Base(absPath)
			}
		}
	}

	if len(req.Directories) == 0 && len(allFiles) == 0 {
		fmt.Fprintln(os.Stderr, "error: no files or directories specified")
		fmt.Fprintln(os.Stderr, "usage: -files=file1,file2 or -dirs=dir1,dir2")
		flag.Usage()
		os.Exit(1)
	}

	if len(allFiles) > 0 {
		req.Files = allFiles
	}

	fmt.Printf("requesting archive...\n")
	fmt.Printf("  directories: %v\n", req.Directories)
	fmt.Printf("  files: %d\n", len(req.Files))
	fmt.Printf("  compression: %d\n", req.Compression)
	if len(req.IncludePattern) > 0 {
		fmt.Printf("  include: %v\n", req.IncludePattern)
	}
	if len(req.ExcludePattern) > 0 {
		fmt.Printf("  exclude: %v\n", req.ExcludePattern)
	}

	resp, err := client.Archive(req, *outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "archive failed: %v\n", err)
		os.Exit(1)
	}
	if resp != nil {
		defer resp.Body.Close()
	}

	fileCount := resp.Header.Get("X-File-Count")
	writtenSize := resp.Header.Get("X-Written-Size")

	fmt.Printf("archive created: %s\n", *outputFile)
	if fileCount != "" {
		fmt.Printf("  files: %s\n", fileCount)
	}
	if writtenSize != "" {
		fmt.Printf("  size: %s bytes\n", writtenSize)
	}
}
