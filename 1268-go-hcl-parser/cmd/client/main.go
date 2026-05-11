package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"hcl-parser/pkg/api"
)

var (
	serverURL = "http://localhost:8080"
)

func main() {
	if envURL := os.Getenv("HCL_SERVER_URL"); envURL != "" {
		serverURL = envURL
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "parse":
		err = cmdParse(args)
	case "format":
		err = cmdFormat(args)
	case "get":
		err = cmdGet(args)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`HCL Parser CLI

Usage:
  hcl parse <file>                Parse HCL file and print JSON
  hcl format <file>               Format HCL file
  hcl get <file> --path <path>    Extract value by dot path

Options:
  -h, --help                      Show this help message`)
}

func readFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", path, err)
	}
	return string(content), nil
}

func httpPost(endpoint string, reqBody interface{}, respBody interface{}) error {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(serverURL+endpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if err := json.Unmarshal(respBytes, respBody); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	return nil
}

func cmdParse(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: hcl parse <file>")
	}

	content, err := readFile(args[0])
	if err != nil {
		return err
	}

	var resp api.ParseResponse
	if err := httpPost("/hcl/parse", &api.ParseRequest{Content: content}, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}

	jsonBytes, err := json.MarshalIndent(resp.Body, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format JSON: %v", err)
	}
	fmt.Println(string(jsonBytes))

	return nil
}

func cmdFormat(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: hcl format <file>")
	}

	content, err := readFile(args[0])
	if err != nil {
		return err
	}

	var resp api.FormatResponse
	if err := httpPost("/hcl/format", &api.FormatRequest{Content: content}, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}

	fmt.Print(resp.Content)

	return nil
}

func cmdGet(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: hcl get <file> --path <path>")
	}

	filePath := args[0]
	var path string

	fs := flag.NewFlagSet("get", flag.ContinueOnError)
	fs.StringVar(&path, "path", "", "Dot-separated path to extract")
	
	remainingArgs := args[1:]
	if err := fs.Parse(remainingArgs); err != nil {
		return err
	}

	if path == "" {
		for i, arg := range remainingArgs {
			if (arg == "--path" || arg == "-path") && i+1 < len(remainingArgs) {
				path = remainingArgs[i+1]
				break
			}
			if strings.HasPrefix(arg, "--path=") {
				path = strings.TrimPrefix(arg, "--path=")
				break
			}
		}
	}

	if path == "" {
		return fmt.Errorf("--path is required")
	}

	content, err := readFile(filePath)
	if err != nil {
		return err
	}

	var resp api.GetResponse
	if err := httpPost("/hcl/get", &api.GetRequest{Content: content, Path: path}, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}

	jsonBytes, err := json.MarshalIndent(resp.Value, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format JSON: %v", err)
	}
	fmt.Println(string(jsonBytes))

	return nil
}
