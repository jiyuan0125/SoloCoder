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

	"csvjson/pkg/api"
)

const defaultServer = "http://localhost:8105"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	subCmd := os.Args[1]
	if subCmd == "-h" || subCmd == "--help" {
		printUsage()
		return
	}
	switch subCmd {
	case "csv-to-json":
		runCSVToJSON(os.Args[2:])
	case "json-to-csv":
		runJSONToCSV(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", subCmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	prog := filepath.Base(os.Args[0])
	fmt.Printf(`Usage:
  %s csv-to-json [options] [<file>]
  %s json-to-csv [options] [<file>]

Options:
  -server URL    server URL (default: %s)
  -delimiter D   delimiter: comma or tab (default: comma)
  -pretty        pretty print JSON output
  -max-digits N  max digits for numeric conversion (default: 15)
`, prog, prog, defaultServer)
}

func readInput(path string) (string, error) {
	if path == "" || path == "-" {
		data, err := io.ReadAll(os.Stdin)
		return string(data), err
	}
	data, err := os.ReadFile(path)
	return string(data), err
}

func runCSVToJSON(args []string) {
	fs := flag.NewFlagSet("csv-to-json", flag.ExitOnError)
	server := fs.String("server", defaultServer, "server URL")
	delimiter := fs.String("delimiter", "comma", "delimiter")
	pretty := fs.Bool("pretty", false, "pretty JSON")
	maxDigits := fs.Int("max-digits", 15, "max digits")
	fs.Parse(args)
	path := ""
	if fs.NArg() > 0 {
		path = fs.Arg(0)
	}
	content, err := readInput(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read input: %v\n", err)
		os.Exit(1)
	}
	req := api.ConvertRequest{
		Content:   content,
		Delimiter: *delimiter,
		Pretty:    *pretty,
		MaxDigits: *maxDigits,
	}
	resp, err := callServer(*server, api.PathCSVToJSON, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
	if !resp.Success {
		fmt.Fprintf(os.Stderr, "error: %s\n", resp.Error)
		os.Exit(1)
	}
	fmt.Print(resp.Content)
}

func runJSONToCSV(args []string) {
	fs := flag.NewFlagSet("json-to-csv", flag.ExitOnError)
	server := fs.String("server", defaultServer, "server URL")
	delimiter := fs.String("delimiter", "comma", "delimiter")
	maxDigits := fs.Int("max-digits", 15, "max digits")
	fs.Parse(args)
	path := ""
	if fs.NArg() > 0 {
		path = fs.Arg(0)
	}
	content, err := readInput(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read input: %v\n", err)
		os.Exit(1)
	}
	req := api.ConvertRequest{
		Content:   content,
		Delimiter: *delimiter,
		MaxDigits: *maxDigits,
	}
	resp, err := callServer(*server, api.PathJSONToCSV, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
	if !resp.Success {
		fmt.Fprintf(os.Stderr, "error: %s\n", resp.Error)
		os.Exit(1)
	}
	fmt.Print(resp.Content)
}

func callServer(serverURL, path string, req api.ConvertRequest) (*api.ConvertResponse, error) {
	url := strings.TrimRight(serverURL, "/") + path
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	httpResp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	var resp api.ConvertResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &resp, nil
}
