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

	"github.com/example/openapi-parser/api"
)

func getServerURL() string {
	url := os.Getenv("OAPI_SERVER")
	if url == "" {
		url = "http://localhost:8080"
	}
	url = strings.TrimRight(url, "/")
	return url
}

func readFileContent(filename string) (string, string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", "", fmt.Errorf("failed to read file: %w", err)
	}
	ext := strings.ToLower(filepath.Ext(filename))
	format := "yaml"
	if ext == ".json" {
		format = "json"
	}
	return string(data), format, nil
}

func sendRequest(endpoint string, payload interface{}) ([]byte, error) {
	serverURL := getServerURL()
	url := serverURL + endpoint

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func cmdCheck(filename string) error {
	content, format, err := readFileContent(filename)
	if err != nil {
		return err
	}

	payload := api.ValidateRequest{
		Format:   format,
		Filename: filename,
		Content:  content,
	}

	respBody, err := sendRequest("/openapi/validate", payload)
	if err != nil {
		return err
	}

	var resp api.ValidateResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("server error: %s", resp.Error)
	}

	if resp.Valid {
		fmt.Printf("✓ File '%s' is valid\n", filename)
		return nil
	}

	fmt.Printf("✗ File '%s' has %d issue(s):\n", filename, len(resp.Errors))
	for i, e := range resp.Errors {
		fmt.Printf("  %d. [%s] %s\n", i+1, e.Field, e.Message)
	}
	os.Exit(1)
	return nil
}

func cmdListPaths(filename string) error {
	content, format, err := readFileContent(filename)
	if err != nil {
		return err
	}

	payload := api.ListPathsRequest{
		Format:   format,
		Filename: filename,
		Content:  content,
	}

	respBody, err := sendRequest("/openapi/list-paths", payload)
	if err != nil {
		return err
	}

	var resp api.ListPathsResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("server error: %s", resp.Error)
	}

	fmt.Printf("Paths in '%s':\n", filename)
	for _, p := range resp.Paths {
		fmt.Printf("  %s: %s\n", p.Path, strings.Join(p.Methods, ", "))
	}
	return nil
}

func cmdResolve(filename string, ref string, all bool) error {
	content, format, err := readFileContent(filename)
	if err != nil {
		return err
	}

	payload := api.ResolveRequest{
		Format:   format,
		Filename: filename,
		Content:  content,
		Ref:      ref,
		All:      all,
	}

	respBody, err := sendRequest("/openapi/resolve", payload)
	if err != nil {
		return err
	}

	var resp api.ResolveResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("server error: %s", resp.Error)
	}

	prettyJSON, err := json.MarshalIndent(resp.Data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}
	fmt.Println(string(prettyJSON))
	return nil
}

func printUsage() {
	fmt.Println("OpenAPI 3.0 Parser CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  oapi check <file>           Validate OpenAPI file")
	fmt.Println("  oapi list-paths <file>      List all paths and operations")
	fmt.Println("  oapi resolve <file> --ref <ref>  Resolve and expand a $ref")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  OAPI_SERVER   Server URL (default: http://localhost:8080)")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "check":
		if len(args) < 1 {
			fmt.Println("Error: missing file argument")
			printUsage()
			os.Exit(1)
		}
		if err := cmdCheck(args[0]); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	case "list-paths":
		if len(args) < 1 {
			fmt.Println("Error: missing file argument")
			printUsage()
			os.Exit(1)
		}
		if err := cmdListPaths(args[0]); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	case "resolve":
		if len(args) < 1 {
			fmt.Println("Error: missing file argument")
			printUsage()
			os.Exit(1)
		}

		fs := flag.NewFlagSet("resolve", flag.ExitOnError)
		refFlag := fs.String("ref", "", "The $ref to resolve (e.g. #/components/schemas/Pet)")
		allFlag := fs.Bool("all", false, "Resolve all references in the entire document")
		fs.Parse(args[1:])

		filename := args[0]
		if !*allFlag && *refFlag == "" {
			fmt.Println("Error: --ref is required (or use --all to resolve everything)")
			printUsage()
			os.Exit(1)
		}

		if err := cmdResolve(filename, *refFlag, *allFlag); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}
