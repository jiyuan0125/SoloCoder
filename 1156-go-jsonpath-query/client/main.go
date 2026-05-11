package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"jsonpath/api"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "query":
		runQuery(os.Args[2:])
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`jsonpath client
Usage:
  client query --json <file-or-string> --path <jsonpath> [--compact] [--server <url>]

Flags:
  --json     JSON string or path to JSON file (required)
  --path     JSONPath expression (required)
  --compact  Output compact JSON instead of indented
  --server   Server URL (default: http://localhost:8600/query)`)
}

func runQuery(args []string) {
	fs := flag.NewFlagSet("query", flag.ExitOnError)
	jsonFlag := fs.String("json", "", "JSON string or file path")
	pathFlag := fs.String("path", "", "JSONPath expression")
	compactFlag := fs.Bool("compact", false, "Compact output")
	serverFlag := fs.String("server", "http://localhost:8600/query", "Server URL")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, "error parsing flags:", err)
		os.Exit(1)
	}

	if *jsonFlag == "" || *pathFlag == "" {
		printUsage()
		os.Exit(1)
	}

	jsonData, err := readJSON(*jsonFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading JSON:", err)
		os.Exit(1)
	}

	req := api.QueryRequest{
		JSON: string(jsonData),
		Path: *pathFlag,
	}
	reqBytes, _ := json.Marshal(req)

	resp, err := http.Post(*serverFlag, "application/json", bytes.NewReader(reqBytes))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error calling server:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var apiResp api.QueryResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		fmt.Fprintln(os.Stderr, "error parsing response:", err)
		fmt.Fprintln(os.Stderr, string(body))
		os.Exit(1)
	}

	if apiResp.Error != "" {
		fmt.Fprintln(os.Stderr, "error:", apiResp.Error)
		os.Exit(1)
	}

	if apiResp.MatchCount == 0 {
		fmt.Println("no matches")
		return
	}

	var output []byte
	if *compactFlag {
		output, err = json.Marshal(apiResp.Results)
	} else {
		output, err = json.MarshalIndent(apiResp.Results, "", "  ")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error marshaling results:", err)
		os.Exit(1)
	}
	fmt.Println(string(output))
}

func readJSON(input string) ([]byte, error) {
	info, err := os.Stat(input)
	if err == nil && !info.IsDir() {
		return os.ReadFile(input)
	}
	if isLikelyJSON(input) {
		return []byte(input), nil
	}
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", input)
	}
	return nil, err
}

func isLikelyJSON(s string) bool {
	trimmed := []byte{}
	for _, b := range s {
		if b != ' ' && b != '\t' && b != '\n' && b != '\r' {
			trimmed = append(trimmed, byte(b))
		}
	}
	if len(trimmed) == 0 {
		return false
	}
	first := trimmed[0]
	last := trimmed[len(trimmed)-1]
	return (first == '{' && last == '}') || (first == '[' && last == ']')
}
