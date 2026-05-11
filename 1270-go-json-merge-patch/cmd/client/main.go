package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"jsonmerge/pkg/api"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "apply":
		handleApply()
	case "diff":
		handleDiff()
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  merge apply --original <file> --patch <file> [--server <url>]")
	fmt.Println("  merge diff <file1> <file2> [--server <url>]")
}

func handleApply() {
	fs := flag.NewFlagSet("apply", flag.ExitOnError)
	originalFile := fs.String("original", "", "Original JSON file")
	patchFile := fs.String("patch", "", "Patch JSON file")
	serverURL := fs.String("server", "http://localhost:8080", "Server URL")
	fs.Parse(os.Args[2:])

	if *originalFile == "" || *patchFile == "" {
		fmt.Fprintln(os.Stderr, "Error: --original and --patch are required")
		fs.Usage()
		os.Exit(1)
	}

	original, err := os.ReadFile(*originalFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading original file: %v\n", err)
		os.Exit(1)
	}

	patch, err := os.ReadFile(*patchFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading patch file: %v\n", err)
		os.Exit(1)
	}

	if !json.Valid(original) {
		fmt.Fprintf(os.Stderr, "Error: original file contains invalid JSON\n")
		os.Exit(1)
	}

	if !json.Valid(patch) {
		fmt.Fprintf(os.Stderr, "Error: patch file contains invalid JSON\n")
		os.Exit(1)
	}

	req := api.ApplyRequest{
		Original: json.RawMessage(original),
		Patch:    json.RawMessage(patch),
	}

	var resp api.ApplyResponse
	if err := callServer(*serverURL+"/merge/apply", req, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if resp.Error != "" {
		fmt.Fprintf(os.Stderr, "Server error: %s\n", resp.Error)
		os.Exit(1)
	}

	var pretty bytes.Buffer
	json.Indent(&pretty, resp.Result, "", "  ")
	fmt.Println(pretty.String())
}

func handleDiff() {
	fs := flag.NewFlagSet("diff", flag.ExitOnError)
	serverURL := fs.String("server", "http://localhost:8080", "Server URL")
	fs.Parse(os.Args[2:])

	args := fs.Args()
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: two files required")
		fs.Usage()
		os.Exit(1)
	}

	file1, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", args[0], err)
		os.Exit(1)
	}

	file2, err := os.ReadFile(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", args[1], err)
		os.Exit(1)
	}

	if !json.Valid(file1) {
		fmt.Fprintf(os.Stderr, "Error: %s contains invalid JSON\n", args[0])
		os.Exit(1)
	}

	if !json.Valid(file2) {
		fmt.Fprintf(os.Stderr, "Error: %s contains invalid JSON\n", args[1])
		os.Exit(1)
	}

	req := api.DiffRequest{
		Source: json.RawMessage(file1),
		Target: json.RawMessage(file2),
	}

	var resp api.DiffResponse
	if err := callServer(*serverURL+"/merge/diff", req, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if resp.Error != "" {
		fmt.Fprintf(os.Stderr, "Server error: %s\n", resp.Error)
		os.Exit(1)
	}

	var pretty bytes.Buffer
	json.Indent(&pretty, resp.Patch, "", "  ")
	fmt.Println(pretty.String())
}

func callServer(url string, reqBody interface{}, respBody interface{}) error {
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return json.Unmarshal(body, respBody)
}
