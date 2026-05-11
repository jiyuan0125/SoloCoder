package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/example/icaltool/api"
)

const defaultServerURL = "http://localhost:8104"

type Client struct {
	ServerURL string
}

func NewClient(serverURL string) *Client {
	if serverURL == "" {
		serverURL = defaultServerURL
	}
	return &Client{
		ServerURL: serverURL,
	}
}

func (c *Client) post(endpoint string, reqBody interface{}, respBody interface{}) error {
	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	resp, err := http.Post(c.ServerURL+endpoint, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
			return fmt.Errorf("server error: %s", errResp.Error)
		}
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, respBody); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	return nil
}

func (c *Client) Parse(data string) (*api.ParseResponse, error) {
	req := api.ParseRequest{
		ICalData: data,
	}
	var resp api.ParseResponse
	if err := c.post("/parse", &req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) Generate(events []api.EventDTO) (*api.GenerateResponse, error) {
	req := api.GenerateRequest{
		Events: events,
	}
	var resp api.GenerateResponse
	if err := c.post("/generate", &req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ExpandRules(rrule, dtstart, tzid, start, end string) (*api.ExpandRulesResponse, error) {
	req := api.ExpandRulesRequest{
		RRULE:   rrule,
		DTStart: dtstart,
		TZID:    tzid,
		Start:   start,
		End:     end,
	}
	var resp api.ExpandRulesResponse
	if err := c.post("/expand-rules", &req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func printHelp() {
	fmt.Println(`iCal Tool - A tool for parsing, generating, and expanding iCal calendar files

Usage:
  ical-tool <command> [options]

Commands:
  parse          Parse an iCal file
  generate       Generate an iCal file
  expand-rules   Expand RRULE to get event instances

Global Options:
  --server <url>    Server URL (default: http://localhost:8104)
  --help          Show this help message

Examples:
  ical-tool parse --input calendar.ics
  ical-tool expand-rules --rrule "FREQ=WEEKLY;BYDAY=MO,WE,FR --dtstart 20240101T090000 --start 2024-01-01T00:00:00Z --end 2024-06-30T23:59:59Z`)
}

func readFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	return string(data), nil
}

func writeFile(path string, data string) error {
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create dir: %w", err)
		}
	}
	return ioutil.WriteFile(path, []byte(data), 0644)
}

func handleParseCommand(args []string, serverURL string) {
	fs := flag.NewFlagSet("parse", flag.ExitOnError)
	inputPath := fs.String("input", "", "Input iCal file path")
	outputPath := fs.String("output", "", "Output file path (optional, prints to stdout if not specified")
	fs.Parse(args)

	if *inputPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --input is required")
		os.Exit(1)
	}

	data, err := readFile(*inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	client := NewClient(serverURL)
	resp, err := client.Parse(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	output, _ := json.MarshalIndent(resp, "", "  ")

	if *outputPath != "" {
		if err := writeFile(*outputPath, string(output)); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Parsed %d events\n", len(resp.Events))
	} else {
		fmt.Println(string(output))
	}
}

func handleGenerateCommand(args []string, serverURL string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	inputPath := fs.String("input", "", "Input JSON file path")
	outputPath := fs.String("output", "output.ics", "Output iCal file path")
	fs.Parse(args)

	if *inputPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --input is required")
		os.Exit(1)
	}

	data, err := readFile(*inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var req api.GenerateRequest
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	client := NewClient(serverURL)
	resp, err := client.Generate(req.Events)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := writeFile(*outputPath, resp.ICalData); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated iCal file: %s\n", *outputPath)
}

func handleExpandRulesCommand(args []string, serverURL string) {
	fs := flag.NewFlagSet("expand-rules", flag.ExitOnError)
	rrule := fs.String("rrule", "", "RRULE string (required)")
	dtstart := fs.String("dtstart", "", "DTSTART in format 20060102T150405 or RFC3339 (required)")
	tzid := fs.String("tzid", "", "Timezone ID, e.g. America/New_York")
	start := fs.String("start", "", "Start time in RFC3339 (required)")
	end := fs.String("end", "", "End time in RFC3339 (required)")
	outputPath := fs.String("output", "", "Output file path (optional)")
	fs.Parse(args)

	if *rrule == "" || *dtstart == "" || *start == "" || *end == "" {
		fmt.Fprintln(os.Stderr, "Error: --rrule, --dtstart, --start, and --end are required")
		os.Exit(1)
	}

	client := NewClient(serverURL)
	resp, err := client.ExpandRules(*rrule, *dtstart, *tzid, *start, *end)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var output strings.Builder
	for i, inst := range resp.Instances {
		output.WriteString(fmt.Sprintf("%d. %s\n", i+1, inst))
	}

	if *outputPath != "" {
		if err := writeFile(*outputPath, output.String()); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Expanded %d instances\n", len(resp.Instances))
	} else {
		fmt.Print(output.String())
	}
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	serverURL := defaultServerURL
	if envURL := os.Getenv("ICAL_SERVER_URL"); envURL != "" {
		serverURL = envURL
	}

	args := os.Args[1:]
	command := args[0]
	remaining := args[1:]

	for i := 0; i < len(args); i++ {
		if args[i] == "--server" && i+1 < len(args) {
			serverURL = args[i+1]
			remaining = append(remaining[2:])
			break
		}
	}

	switch command {
	case "parse":
		handleParseCommand(remaining, serverURL)
	case "generate":
		handleGenerateCommand(remaining, serverURL)
	case "expand-rules":
		handleExpandRulesCommand(remaining, serverURL)
	case "--help", "-h":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}
