package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/example/callgraph/internal/api"
)

const (
	defaultServerURL = "http://localhost:8300"
	envServerURL     = "CALLGRAPH_SERVER"
)

func collectGoFiles(dir string) ([]api.File, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("abs path: %w", err)
	}

	var files []api.File
	err = filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.Contains(path, "_test.go") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		relPath, err := filepath.Rel(absDir, path)
		if err != nil {
			relPath = filepath.Base(path)
		}

		files = append(files, api.File{
			Name:    relPath,
			Content: string(content),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func sendRequest(serverURL string, req api.AnalyzeRequest) (*api.AnalyzeResponse, error) {
	data, err := api.EncodeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}

	url := strings.TrimRight(serverURL, "/") + "/analyze"
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("server error %d: %s", resp.StatusCode, string(body))
	}

	apiResp, err := api.DecodeResponse(body)
	if err != nil {
		return nil, fmt.Errorf("decode response: %w, body=%s", err, string(body))
	}

	return &apiResp, nil
}

func getServerURL() string {
	if v := os.Getenv(envServerURL); v != "" {
		return v
	}
	return defaultServerURL
}

func main() {
	serverURLFlag := flag.String("server", "", "server URL (default: http://localhost:8300, or $CALLGRAPH_SERVER)")
	formatFlag := flag.String("format", "dot", "output format: dot or json")
	outputFlag := flag.String("output", "", "output file (default: stdout)")
	dirFlag := flag.String("dir", "", "directory containing Go files (default: current directory)")
	flag.Parse()

	serverURL := getServerURL()
	if *serverURLFlag != "" {
		serverURL = *serverURLFlag
	}

	dir := *dirFlag
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			log.Fatalf("get working directory: %v", err)
		}
	}

	format := strings.ToLower(strings.TrimSpace(*formatFlag))
	if format == "" {
		format = "dot"
	}
	if format != "dot" && format != "json" {
		log.Fatalf("unsupported format: %s (expected dot or json)", format)
	}

	files, err := collectGoFiles(dir)
	if err != nil {
		log.Fatalf("collect files: %v", err)
	}
	if len(files) == 0 {
		log.Fatalf("no .go files found in %s", dir)
	}

	absDir, _ := filepath.Abs(dir)
	pkgPath := filepath.Base(absDir)

	fmt.Printf("Package: %s\n", pkgPath)
	fmt.Printf("Dir: %s\n", absDir)
	fmt.Printf("Files: %d\n", len(files))
	fmt.Printf("Format: %s\n", format)
	fmt.Printf("Server: %s\n", serverURL)
	fmt.Println()

	req := api.AnalyzeRequest{
		PackagePath: pkgPath,
		Files:       files,
		Format:      format,
	}

	resp, err := sendRequest(serverURL, req)
	if err != nil {
		log.Fatalf("request failed: %v", err)
	}

	if !resp.Success {
		log.Fatalf("server returned error: %s", resp.Message)
	}

	output := resp.Output
	if *outputFlag != "" {
		if err := os.WriteFile(*outputFlag, []byte(output), 0644); err != nil {
			log.Fatalf("write output: %v", err)
		}
		fmt.Printf("\nSaved to: %s\n", *outputFlag)
	} else {
		fmt.Println(output)
	}
}
