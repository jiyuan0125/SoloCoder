package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	baseURL := "http://localhost:8303"
	if envURL := os.Getenv("CSV_SERVER_URL"); envURL != "" {
		baseURL = envURL
	}

	client := NewAPIClient(baseURL)

	cmd := os.Args[1]

	switch cmd {
	case "parse":
		handleParse(client)
	case "serialize":
		handleSerialize(client)
	case "validate":
		handleValidate(client)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`CSV RFC 4180 Client - Command line tool for CSV parsing, serializing and validation

Usage:
  csv-client <command> [options]

Commands:
  parse      - Parse CSV text
  serialize  - Serialize data to CSV
  validate   - Validate CSV format
  help       - Show this help

Environment:
  CSV_SERVER_URL  - Server URL (default: http://localhost:8303)

Examples:
  csv-client parse --file data.csv
  csv-client parse --file data.csv --with-header
  csv-client parse --csv 'a,b,c'
  csv-client serialize --data '[["a","b"],["1","2"]]'
  cat data.csv | csv-client validate --with-header`)
}

func handleParse(client *APIClient) {
	fs := flag.NewFlagSet("parse", flag.ExitOnError)
	filePath := fs.String("file", "", "CSV file to parse")
	csvText := fs.String("csv", "", "CSV text to parse (inline)")
	withHeader := fs.Bool("with-header", false, "Parse first row as headers")
	fs.Parse(os.Args[2:])

	text, err := readInput(*filePath, *csvText)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.Parse(text, *withHeader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "parse failed: %s\n", resp.Error)
		os.Exit(1)
	}

	output, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(output))
}

func handleSerialize(client *APIClient) {
	fs := flag.NewFlagSet("serialize", flag.ExitOnError)
	dataText := fs.String("data", "", "JSON array of arrays (e.g., '[\"a\",\"b\"],[\"1\",\"2\"]')")
	fs.Parse(os.Args[2:])

	if *dataText == "" {
		fmt.Fprintln(os.Stderr, "error: --data is required")
		os.Exit(1)
	}

	var data [][]string
	if err := json.Unmarshal([]byte(*dataText), &data); err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid JSON data: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.Serialize(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "serialize failed: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Print(resp.CSV)
}

func handleValidate(client *APIClient) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	filePath := fs.String("file", "", "CSV file to validate")
	csvText := fs.String("csv", "", "CSV text to validate (inline)")
	withHeader := fs.Bool("with-header", false, "Check column count against header row")
	fs.Parse(os.Args[2:])

	text, err := readInput(*filePath, *csvText)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.Validate(text, *withHeader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "validate failed: %s\n", resp.Error)
		os.Exit(1)
	}

	if resp.Valid {
		fmt.Println("CSV is valid")
		os.Exit(0)
	}

	fmt.Printf("CSV has %d validation error(s):\n", len(resp.Errors))
	for _, e := range resp.Errors {
		fmt.Printf("  line %d, column %d: %s\n", e.Line, e.Column, e.Message)
	}
	os.Exit(1)
}

func readInput(filePath string, csvText string) (string, error) {
	if csvText != "" {
		return csvText, nil
	}

	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %v", err)
		}
		return string(data), nil
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read stdin: %v", err)
		}
		return strings.TrimSuffix(string(data), "\n"), nil
	}

	return "", fmt.Errorf("no input provided. Use --file, --csv, or pipe from stdin")
}
