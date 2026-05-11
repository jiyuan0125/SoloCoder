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

	"go-fuzz-corpus/api"
)

func getServerURL() string {
	url := os.Getenv("FUZZ_CORPUS_SERVER_URL")
	if url == "" {
		url = "http://localhost:8410"
	}
	return url
}

func makeRequest(method, endpoint string, reqBody, respBody interface{}) error {
	serverURL := getServerURL()
	url := fmt.Sprintf("%s%s", serverURL, endpoint)

	var body io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp api.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("server error: %d", resp.StatusCode)
		}
		return fmt.Errorf("server error: %s", errResp.Error)
	}

	if respBody != nil {
		return json.NewDecoder(resp.Body).Decode(respBody)
	}
	return nil
}

func cmdScan(args []string) error {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	path := fs.String("path", ".", "Root path to scan")
	if err := fs.Parse(args); err != nil {
		return err
	}

	absPath, err := filepath.Abs(*path)
	if err != nil {
		absPath = *path
	}

	req := api.ScanRequest{RootPath: absPath}
	var resp api.ScanResponse
	if err := makeRequest("POST", "/api/scan", &req, &resp); err != nil {
		return err
	}

	if len(resp.Targets) == 0 {
		fmt.Println("No fuzz targets found.")
		return nil
	}

	fmt.Println("Found fuzz targets:")
	fmt.Println("==================")
	for _, target := range resp.Targets {
		fmt.Printf("\nTarget: %s\n", target.TargetName)
		fmt.Printf("  Path: %s\n", target.FilePath)
		fmt.Printf("  Files: %d\n", target.FileCount)
		fmt.Printf("  Total Size: %d bytes\n", target.TotalSize)

		var manualCount int
		for _, f := range target.Files {
			if f.IsManual {
				manualCount++
			}
		}
		fmt.Printf("  Manual Seeds: %d\n", manualCount)
		fmt.Printf("  Auto Seeds: %d\n", target.FileCount-manualCount)
	}

	return nil
}

func cmdDedup(args []string) error {
	fs := flag.NewFlagSet("dedup", flag.ExitOnError)
	target := fs.String("target", "", "Target fuzz function name (required)")
	path := fs.String("path", ".", "Root path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *target == "" {
		return fmt.Errorf("--target is required")
	}

	absPath, err := filepath.Abs(*path)
	if err != nil {
		absPath = *path
	}

	req := api.DedupRequest{
		TargetName: *target,
		RootPath:   absPath,
	}
	var resp api.DedupResponse
	if err := makeRequest("POST", "/api/dedup", &req, &resp); err != nil {
		return err
	}

	fmt.Printf("Deduplication result for %s:\n", resp.TargetName)
	fmt.Printf("  Removed: %d files\n", resp.RemovedCount)
	fmt.Printf("  Retained: %d files\n", resp.RetainedCount)

	if len(resp.RemovedFiles) > 0 {
		fmt.Println("\nRemoved files:")
		for _, f := range resp.RemovedFiles {
			fmt.Printf("  - %s\n", f)
		}
	}

	return nil
}

func cmdMutate(args []string) error {
	fs := flag.NewFlagSet("mutate", flag.ExitOnError)
	target := fs.String("target", "", "Target fuzz function name (required)")
	path := fs.String("path", ".", "Root path")
	count := fs.Int("count", 100, "Number of mutations to generate")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *target == "" {
		return fmt.Errorf("--target is required")
	}

	absPath, err := filepath.Abs(*path)
	if err != nil {
		absPath = *path
	}

	req := api.MutateRequest{
		TargetName: *target,
		RootPath:   absPath,
		Count:      *count,
	}
	var resp api.MutateResponse
	if err := makeRequest("POST", "/api/mutate", &req, &resp); err != nil {
		return err
	}

	fmt.Printf("Mutation result for %s:\n", resp.TargetName)
	fmt.Printf("  Generated: %d new seeds\n", resp.Generated)
	fmt.Printf("  Failed: %d\n", resp.Failed)

	if len(resp.NewFiles) > 0 {
		fmt.Println("\nNew files:")
		for _, f := range resp.NewFiles {
			fmt.Printf("  - %s\n", f)
		}
	}

	return nil
}

func cmdImport(args []string) error {
	fs := flag.NewFlagSet("import", flag.ExitOnError)
	target := fs.String("target", "", "Target fuzz function name (required)")
	crashDir := fs.String("crashdir", "", "Directory containing crash files (required)")
	path := fs.String("path", ".", "Root path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *target == "" {
		return fmt.Errorf("--target is required")
	}
	if *crashDir == "" {
		return fmt.Errorf("--crashdir is required")
	}

	absPath, err := filepath.Abs(*path)
	if err != nil {
		absPath = *path
	}

	absCrashDir, err := filepath.Abs(*crashDir)
	if err != nil {
		absCrashDir = *crashDir
	}

	req := api.ImportRequest{
		CrashDir:   absCrashDir,
		TargetName: *target,
		RootPath:   absPath,
	}
	var resp api.ImportResponse
	if err := makeRequest("POST", "/api/import", &req, &resp); err != nil {
		return err
	}

	fmt.Printf("Import result for %s:\n", resp.TargetName)
	fmt.Printf("  Imported: %d files\n", resp.Imported)
	fmt.Printf("  Skipped (duplicates): %d\n", resp.Skipped)

	if len(resp.NewFiles) > 0 {
		fmt.Println("\nImported files:")
		for _, f := range resp.NewFiles {
			fmt.Printf("  - %s\n", f)
		}
	}

	return nil
}

func printUsage() {
	fmt.Println("Usage: fuzz-corpus-client <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  scan    - Scan for fuzz corpus directories")
	fmt.Println("  dedup   - Deduplicate corpus files by content hash")
	fmt.Println("  mutate  - Generate new corpus files via mutation")
	fmt.Println("  import  - Import crash files into corpus")
	fmt.Println("\nEnvironment Variables:")
	fmt.Println("  FUZZ_CORPUS_SERVER_URL - Server URL (default: http://localhost:8410)")
	fmt.Println("\nUse 'fuzz-corpus-client <command> -h' for help with a specific command.")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	var err error
	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "scan":
		err = cmdScan(args)
	case "dedup":
		err = cmdDedup(args)
	case "mutate":
		err = cmdMutate(args)
	case "import":
		err = cmdImport(args)
	case "-h", "--help", "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
