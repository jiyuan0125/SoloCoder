package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"syslog-parser/pkg/api"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	inputFile := flag.String("file", "", "Input file path (use - for stdin)")
	format := flag.String("format", "pretty", "Output format: pretty, json, or raw")
	command := flag.String("command", "parse", "Command: parse, batch, or list")
	flag.Parse()

	if envURL := os.Getenv("SYSLOG_SERVER_URL"); envURL != "" {
		*serverURL = envURL
	}

	switch *command {
	case "parse":
		if err := handleParse(*serverURL, *inputFile, *format); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "batch":
		if err := handleBatchParse(*serverURL, *inputFile, *format); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "list":
		if err := handleList(*serverURL, *format); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", *command)
		os.Exit(1)
	}
}

func handleParse(serverURL, inputFile, format string) error {
	var reader io.Reader

	if inputFile == "" || inputFile == "-" {
		reader = os.Stdin
	} else {
		f, err := os.Open(inputFile)
		if err != nil {
			return fmt.Errorf("failed to open file: %v", err)
		}
		defer f.Close()
		reader = f
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		result, err := parseSingleMessage(serverURL, line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
			continue
		}

		outputResult(result, format)
	}

	return scanner.Err()
}

func handleBatchParse(serverURL, inputFile, format string) error {
	var reader io.Reader

	if inputFile == "" || inputFile == "-" {
		reader = os.Stdin
	} else {
		f, err := os.Open(inputFile)
		if err != nil {
			return fmt.Errorf("failed to open file: %v", err)
		}
		defer f.Close()
		reader = f
	}

	var messages []string
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		messages = append(messages, line)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if len(messages) == 0 {
		return fmt.Errorf("no messages to parse")
	}

	result, err := batchParseMessages(serverURL, messages)
	if err != nil {
		return err
	}

	outputBatchResult(result, format)
	return nil
}

func handleList(serverURL, format string) error {
	resp, err := http.Get(serverURL + "/facility-severity")
	if err != nil {
		return fmt.Errorf("failed to call server: %v", err)
	}
	defer resp.Body.Close()

	var result api.FacilitySeverityResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if !result.Success {
		return fmt.Errorf("server error: %s", result.Error)
	}

	if format == "json" {
		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(jsonBytes))
	} else {
		fmt.Println("Facilities:")
		for _, f := range result.Facility {
			fmt.Printf("  %d: %s\n", f.Code, f.Name)
		}
		fmt.Println("\nSeverities:")
		for _, s := range result.Severity {
			fmt.Printf("  %d: %s\n", s.Code, s.Name)
		}
	}

	return nil
}

func parseSingleMessage(serverURL, message string) (*api.ParseResponse, error) {
	reqBody, _ := json.Marshal(api.ParseRequest{Message: message})

	resp, err := http.Post(serverURL+"/parse", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to call server: %v", err)
	}
	defer resp.Body.Close()

	var result api.ParseResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &result, nil
}

func batchParseMessages(serverURL string, messages []string) (*api.BatchParseResponse, error) {
	reqBody, _ := json.Marshal(api.BatchParseRequest{Messages: messages})

	resp, err := http.Post(serverURL+"/batch-parse", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to call server: %v", err)
	}
	defer resp.Body.Close()

	var result api.BatchParseResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &result, nil
}

func outputResult(result *api.ParseResponse, format string) {
	if format == "json" {
		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(jsonBytes))
		return
	}

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Error)
		return
	}

	printPrettyMessage(result.Message)
}

func outputBatchResult(result *api.BatchParseResponse, format string) {
	if format == "json" {
		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(jsonBytes))
		return
	}

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Error)
		return
	}

	for i, msg := range result.Messages {
		fmt.Printf("=== Message %d ===\n", i+1)
		if msg.Error != "" {
			fmt.Printf("Raw: %s\n", msg.Raw)
			fmt.Printf("Error: %s\n\n", msg.Error)
		} else {
			printPrettyMessage(msg.Parsed)
			fmt.Println()
		}
	}
}

func printPrettyMessage(msg *api.SyslogMsg) {
	if msg == nil {
		return
	}

	fmt.Printf("PRI: %d (Facility: %s[%d], Severity: %s[%d])\n",
		msg.PRI, msg.FacilityName, msg.Facility, msg.SeverityName, msg.Severity)
	fmt.Printf("Version: %d\n", msg.Version)
	fmt.Printf("Timestamp: %s (format: %s)\n", msg.Timestamp.Format("2006-01-02 15:04:05 MST"), msg.TimestampFormat)
	fmt.Printf("Hostname: %s\n", msg.Hostname)
	fmt.Printf("AppName: %s\n", msg.AppName)
	fmt.Printf("ProcID: %s\n", msg.ProcID)
	fmt.Printf("MsgID: %s\n", msg.MsgID)

	if len(msg.StructuredData) > 0 {
		fmt.Println("Structured Data:")
		for _, elem := range msg.StructuredData {
			fmt.Printf("  [%s]\n", elem.ID)
			for _, p := range elem.Params {
				fmt.Printf("    %s = %s\n", p.Name, p.Value)
			}
		}
	} else {
		fmt.Println("Structured Data: (none)")
	}

	fmt.Printf("Message: %s\n", msg.Message)
}
