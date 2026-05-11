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

	"go-dep-analyzer/pkg/api"
)

const defaultServerURL = "http://localhost:8080"

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func uploadFiles(serverURL string, goModContent string, goSumContent string) error {
	reqBody := api.UploadRequest{
		GoModContent: goModContent,
		GoSumContent: goSumContent,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(serverURL+"/upload", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned error: %s", string(body))
	}

	var uploadResp api.UploadResponse
	json.NewDecoder(resp.Body).Decode(&uploadResp)

	if !uploadResp.Success {
		return fmt.Errorf("upload failed: %s", uploadResp.Message)
	}

	return nil
}

func getReport(serverURL string) (*api.FullReportResponse, error) {
	resp, err := http.Get(serverURL + "/report")
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned error: %s", string(body))
	}

	var report api.FullReportResponse
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		return nil, fmt.Errorf("failed to parse report: %v", err)
	}

	return &report, nil
}

func printReport(report *api.FullReportResponse) {
	fmt.Println("==================================")
	fmt.Println("  Go Dependency Analysis Report")
	fmt.Println("==================================")
	fmt.Println()

	fmt.Printf("Total Dependencies:     %d\n", report.TotalDeps)
	fmt.Printf("Direct Dependencies:    %d\n", report.DirectDeps)
	fmt.Printf("Indirect Dependencies:  %d\n", report.IndirectDeps)
	fmt.Println()

	fmt.Println("--- Cyclic Dependencies ---")
	if len(report.Cycles) == 0 {
		fmt.Println("  No cyclic dependencies detected.")
	} else {
		for i, cycle := range report.Cycles {
			fmt.Printf("  Cycle %d: %s\n", i+1, strings.Join(cycle.Modules, " -> "))
		}
	}
	fmt.Println()

	fmt.Println("--- Version Conflicts ---")
	if len(report.Conflicts) == 0 {
		fmt.Println("  No version conflicts detected.")
	} else {
		for _, conflict := range report.Conflicts {
			fmt.Printf("  Module: %s\n", conflict.Module)
			fmt.Printf("    Required: %s\n", conflict.RequiredVersion)
			fmt.Printf("    Actual:   %s\n", conflict.ActualVersion)
			fmt.Printf("    Reason:   %s\n", conflict.Reason)
			fmt.Println()
		}
	}

	fmt.Println("--- Deepest Dependencies (Top 5) ---")
	if len(report.DeepestModules) == 0 {
		fmt.Println("  No depth data available.")
	} else {
		for i, md := range report.DeepestModules {
			fmt.Printf("  %d. %s (depth: %d)\n", i+1, md.Module, md.Depth)
		}
	}
	fmt.Println()

	fmt.Println("--- Slimming Suggestions ---")
	if len(report.SlimmingSuggestions) == 0 {
		fmt.Println("  No slimming suggestions available.")
	} else {
		for _, s := range report.SlimmingSuggestions {
			fmt.Printf("  Module: %s\n", s.Module)
			fmt.Printf("  Suggestion: %s\n", s.Suggestion)
			fmt.Println()
		}
	}

	fmt.Println("==================================")
}

func main() {
	dirFlag := flag.String("dir", ".", "Project directory containing go.mod and go.sum")
	serverFlag := flag.String("server", defaultServerURL, "Server URL (default: http://localhost:8080)")
	flag.Parse()

	dir, err := filepath.Abs(*dirFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Invalid directory path: %v\n", err)
		os.Exit(1)
	}

	goModPath := filepath.Join(dir, "go.mod")
	goSumPath := filepath.Join(dir, "go.sum")

	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: go.mod not found in %s\n", dir)
		os.Exit(1)
	}

	goModContent, err := readFile(goModPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to read go.mod: %v\n", err)
		os.Exit(1)
	}

	goSumContent := ""
	if _, err := os.Stat(goSumPath); err == nil {
		goSumContent, err = readFile(goSumPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to read go.sum: %v\n", err)
		}
	} else {
		fmt.Println("Warning: go.sum not found. Some analysis features may be limited.")
	}

	fmt.Printf("Analyzing project in: %s\n", dir)
	fmt.Printf("Connecting to server: %s\n", *serverFlag)
	fmt.Println()

	fmt.Println("Uploading files to server...")
	if err := uploadFiles(*serverFlag, goModContent, goSumContent); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Fetching analysis report...")
	report, err := getReport(*serverFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	printReport(report)
}
