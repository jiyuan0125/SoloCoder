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

	"go-log-parser/protocol"
)

var (
	serverAddr string
	format     string
	fields     string
	inputFile  string
	listFormats bool
	registerName string
	registerFormat string
	registerFields string
	outputFormat string
)

func init() {
	flag.StringVar(&serverAddr, "server", "http://localhost:8080", "Server address")
	flag.StringVar(&format, "format", "nginx_combined", "Log format name or custom format string")
	flag.StringVar(&fields, "fields", "", "Field names (comma-separated, for custom format)")
	flag.StringVar(&inputFile, "file", "", "Input log file path")
	flag.BoolVar(&listFormats, "list", false, "List supported formats")
	flag.StringVar(&registerName, "register-name", "", "Register format: name")
	flag.StringVar(&registerFormat, "register-format", "", "Register format: format string")
	flag.StringVar(&registerFields, "register-fields", "", "Register format: field names (comma-separated)")
	flag.StringVar(&outputFormat, "output", "json", "Output format: json or text")
}

func main() {
	flag.Parse()

	if listFormats {
		if err := listSupportedFormats(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if registerName != "" {
		if registerFormat == "" {
			fmt.Fprintf(os.Stderr, "Error: -register-format is required when using -register-name\n")
			os.Exit(1)
		}
		if err := registerFormatToServer(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if inputFile == "" && flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Error: input file or log content is required\n")
		fmt.Fprintf(os.Stderr, "\nUsage:\n")
		fmt.Fprintf(os.Stderr, "  client -file access.log -format nginx_combined\n")
		fmt.Fprintf(os.Stderr, "  client -list\n")
		fmt.Fprintf(os.Stderr, "  client -register-name myformat -register-format \"$field1 - $field2\" -register-fields \"field1,field2\"\n")
		flag.Usage()
		os.Exit(1)
	}

	if err := parseLogs(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func listSupportedFormats() error {
	resp, err := http.Get(strings.TrimRight(serverAddr, "/") + "/formats")
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned error: %s", string(body))
	}

	var formatsResp protocol.ListFormatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&formatsResp); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	fmt.Println("Supported formats:")
	for _, f := range formatsResp.Formats {
		fmt.Printf("  - %s\n", f)
	}
	return nil
}

func registerFormatToServer() error {
	var fieldList []string
	if registerFields != "" {
		fieldList = strings.Split(registerFields, ",")
		for i, f := range fieldList {
			fieldList[i] = strings.TrimSpace(f)
		}
	}

	req := protocol.FormatRegisterRequest{
		Name:         registerName,
		FormatString: registerFormat,
		Fields:       fieldList,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(
		strings.TrimRight(serverAddr, "/")+"/register",
		"application/json",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned error: %s", string(body))
	}

	var respData protocol.FormatRegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	fmt.Println(respData.Message)
	return nil
}

func parseLogs() error {
	var req protocol.ParseRequest
	req.Format = format

	if fields != "" {
		fieldList := strings.Split(fields, ",")
		for i, f := range fieldList {
			fieldList[i] = strings.TrimSpace(f)
		}
		req.Fields = fieldList
	}

	if inputFile != "" {
		req.FilePath = inputFile
	} else {
		req.LogContent = strings.Join(flag.Args(), " ")
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(
		strings.TrimRight(serverAddr, "/")+"/parse",
		"application/json",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned error: %s", string(body))
	}

	var respData protocol.ParseResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if outputFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(respData)
	} else {
		printTextOutput(&respData)
	}

	return nil
}

func printTextOutput(resp *protocol.ParseResponse) {
	fmt.Printf("Parsed %d entries\n\n", resp.TotalCount)

	for i, entry := range resp.Entries {
		fmt.Printf("=== Entry %d (line %d) ===\n", i+1, entry.LineNum)
		if entry.Timestamp != "" {
			fmt.Printf("  Timestamp: %s\n", entry.Timestamp)
		}
		for key, value := range entry.Fields {
			fmt.Printf("  %s: %s\n", key, value)
		}
		fmt.Println()
	}

	if len(resp.Errors) > 0 {
		fmt.Printf("Errors (%d):\n", len(resp.Errors))
		for _, e := range resp.Errors {
			fmt.Printf("  Line %d: %s\n", e.LineNum, e.Reason)
		}
	}
}
