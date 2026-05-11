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

	"mdtableparser/pkg/common"
)

const (
	defaultServerURL = "http://localhost:8080"
)

func getServerURL() string {
	serverURL := defaultServerURL

	if envURL := os.Getenv("SERVER_URL"); envURL != "" {
		serverURL = envURL
	}

	return serverURL
}

func printUsage() {
	fmt.Println("Markdown Table Parser CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  mdtable <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  extract-tables   Extract tables from Markdown text")
	fmt.Println("  markdown-to-csv  Convert Markdown tables to CSV")
	fmt.Println("  csv-to-markdown  Convert CSV to Markdown table")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  SERVER_URL       Server URL (default: http://localhost:8080)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  mdtable extract-tables -f input.md")
	fmt.Println("  mdtable markdown-to-csv -f input.md -o output.csv")
	fmt.Println("  mdtable csv-to-markdown -f input.csv -o output.md")
	fmt.Println("  echo \"| A | B |\n| --- | --- |\n| 1 | 2 |\" | mdtable markdown-to-csv")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := strings.ToLower(os.Args[1])
	serverURL := getServerURL()

	switch command {
	case "extract-tables":
		handleExtractTables(serverURL, os.Args[2:])
	case "markdown-to-csv":
		handleMarkdownToCSV(serverURL, os.Args[2:])
	case "csv-to-markdown":
		handleCSVToMarkdown(serverURL, os.Args[2:])
	case "-h", "--help", "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleExtractTables(serverURL string, args []string) {
	fs := flag.NewFlagSet("extract-tables", flag.ExitOnError)
	inputFile := fs.String("f", "", "Input Markdown file (optional, reads from stdin if not provided)")
	outputFile := fs.String("o", "", "Output file (optional, writes to stdout if not provided)")
	fs.Parse(args)

	var content string
	var err error

	if *inputFile != "" {
		content, err = readFile(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
			os.Exit(1)
		}
	} else {
		content, err = readStdin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			os.Exit(1)
		}
	}

	reqBody, err := json.Marshal(common.ExtractTablesRequest{
		Markdown: content,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/api/extract-tables", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending request to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp common.ErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err == nil {
			fmt.Fprintf(os.Stderr, "Server error: %s\n", errorResp.Error)
		} else {
			fmt.Fprintf(os.Stderr, "Server returned status %d: %s\n", resp.StatusCode, string(respBody))
		}
		os.Exit(1)
	}

	var result common.ExtractTablesResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}

	output := formatExtractTablesResult(result)

	if *outputFile != "" {
		if err := writeFile(*outputFile, output); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Extracted %d tables to %s\n", len(result.Tables), *outputFile)
	} else {
		fmt.Print(output)
	}
}

func handleMarkdownToCSV(serverURL string, args []string) {
	fs := flag.NewFlagSet("markdown-to-csv", flag.ExitOnError)
	inputFile := fs.String("f", "", "Input Markdown file (optional, reads from stdin if not provided)")
	outputFile := fs.String("o", "", "Output file (optional, writes to stdout if not provided)")
	fs.Parse(args)

	var content string
	var err error

	if *inputFile != "" {
		content, err = readFile(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
			os.Exit(1)
		}
	} else {
		content, err = readStdin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			os.Exit(1)
		}
	}

	reqBody, err := json.Marshal(common.MarkdownToCSVRequest{
		Markdown: content,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/api/markdown-to-csv", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending request to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp common.ErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err == nil {
			fmt.Fprintf(os.Stderr, "Server error: %s\n", errorResp.Error)
		} else {
			fmt.Fprintf(os.Stderr, "Server returned status %d: %s\n", resp.StatusCode, string(respBody))
		}
		os.Exit(1)
	}

	var result common.MarkdownToCSVResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if *outputFile != "" {
		if err := writeFile(*outputFile, result.CSVContent); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Converted %d tables to CSV and saved to %s\n", result.TableCount, *outputFile)
	} else {
		fmt.Print(result.CSVContent)
	}
}

func handleCSVToMarkdown(serverURL string, args []string) {
	fs := flag.NewFlagSet("csv-to-markdown", flag.ExitOnError)
	inputFile := fs.String("f", "", "Input CSV file (optional, reads from stdin if not provided)")
	outputFile := fs.String("o", "", "Output file (optional, writes to stdout if not provided)")
	hasHeader := fs.Bool("header", true, "CSV has header row (default: true)")
	fs.Parse(args)

	var content string
	var err error

	if *inputFile != "" {
		content, err = readFile(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
			os.Exit(1)
		}
	} else {
		content, err = readStdin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			os.Exit(1)
		}
	}

	reqBody, err := json.Marshal(common.CSVToMarkdownRequest{
		CSVContent: content,
		HasHeader:  *hasHeader,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/api/csv-to-markdown", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending request to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp common.ErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err == nil {
			fmt.Fprintf(os.Stderr, "Server error: %s\n", errorResp.Error)
		} else {
			fmt.Fprintf(os.Stderr, "Server returned status %d: %s\n", resp.StatusCode, string(respBody))
		}
		os.Exit(1)
	}

	var result common.CSVToMarkdownResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if *outputFile != "" {
		if err := writeFile(*outputFile, result.Markdown); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Converted CSV to Markdown and saved to %s\n", *outputFile)
	} else {
		fmt.Print(result.Markdown)
	}
}

func readFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func writeFile(filename string, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

func readStdin() (string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func formatExtractTablesResult(result common.ExtractTablesResponse) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Found %d table(s):\n\n", len(result.Tables)))

	for i, table := range result.Tables {
		builder.WriteString(fmt.Sprintf("=== Table %d ===\n", i+1))
		builder.WriteString(fmt.Sprintf("Headers: %v\n", table.Headers))
		builder.WriteString(fmt.Sprintf("Alignments: %v\n", table.Alignments))
		builder.WriteString(fmt.Sprintf("Rows (%d):\n", len(table.Rows)))

		for j, row := range table.Rows {
			builder.WriteString(fmt.Sprintf("  %d: %v\n", j+1, row))
		}
		builder.WriteString("\n")
	}

	return builder.String()
}
