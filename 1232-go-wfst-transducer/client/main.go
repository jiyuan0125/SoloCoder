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

	"wfst/api"
)

type Client struct {
	serverURL string
}

func NewClient(serverURL string) *Client {
	if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
		serverURL = "http://" + serverURL
	}
	return &Client{serverURL: serverURL}
}

func (c *Client) doPost(endpoint string, reqBody interface{}, respBody interface{}) error {
	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.serverURL+endpoint, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		if json.Unmarshal(body, &errResp) == nil {
			return fmt.Errorf("server error [%d]: %s", resp.StatusCode, errResp.Error)
		}
		return fmt.Errorf("server error [%d]: %s", resp.StatusCode, string(body))
	}

	if respBody != nil {
		if err := json.Unmarshal(body, respBody); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

func (c *Client) Create(arcs []api.Arc) (*api.CreateResponse, error) {
	req := api.CreateRequest{Arcs: arcs}
	var resp api.CreateResponse
	if err := c.doPost("/wfst/create", &req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) Search(wfstID string, input []string) (*api.SearchResponse, error) {
	req := api.SearchRequest{
		WFSTID: wfstID,
		Input:  input,
	}
	var resp api.SearchResponse
	if err := c.doPost("/wfst/search", &req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) Compose(wfstIDA, wfstIDB string) (*api.ComposeResponse, error) {
	req := api.ComposeRequest{
		WFSTIDA: wfstIDA,
		WFSTIDB: wfstIDB,
	}
	var resp api.ComposeResponse
	if err := c.doPost("/wfst/compose", &req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func loadArcsFromFile(filename string) ([]api.Arc, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var req api.CreateRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return req.Arcs, nil
}

func printCreateUsage() {
	fmt.Println("Usage: wfst-client create <json-file>")
	fmt.Println()
	fmt.Println("Create a new WFST from a JSON file.")
	fmt.Println()
	fmt.Println("JSON file format:")
	fmt.Println("  {")
	fmt.Println("    \"arcs\": [")
	fmt.Println("      {\"from\": 0, \"input\": \"a\", \"output\": \"b\", \"weight\": 0.5, \"next\": 1}")
	fmt.Println("    ]")
	fmt.Println("  }")
}

func printSearchUsage() {
	fmt.Println("Usage: wfst-client search <wfst-id> <input1> <input2> ...")
	fmt.Println()
	fmt.Println("Search for the best path in a WFST for the given input sequence.")
}

func printComposeUsage() {
	fmt.Println("Usage: wfst-client compose <wfst-id-a> <wfst-id-b>")
	fmt.Println()
	fmt.Println("Compose two WFSTs into a new WFST.")
}

func printUsage() {
	fmt.Println("WFST Client - A command-line tool for interacting with WFST server")
	fmt.Println()
	fmt.Println("Usage: wfst-client [global-options] <command> [command-options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create    - Create a new WFST from JSON file")
	fmt.Println("  search    - Search for best path in a WFST")
	fmt.Println("  compose   - Compose two WFSTs")
	fmt.Println()
	fmt.Println("Global Options:")
	fmt.Println("  -server <url>   Server URL (default: localhost:8213)")
	fmt.Println()
	fmt.Println("Use 'wfst-client <command> -h' for more information about a command.")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	var serverURL string
	fs := flag.NewFlagSet("global", flag.ContinueOnError)
	fs.StringVar(&serverURL, "server", "localhost:8213", "Server URL")
	
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		if args[i] == "-server" && i+1 < len(args) {
			serverURL = args[i+1]
			args = append(args[:i], args[i+2:]...)
			break
		}
	}

	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	client := NewClient(serverURL)
	command := args[0]

	switch command {
	case "create":
		if len(args) < 2 {
			printCreateUsage()
			os.Exit(1)
		}
		filename := args[1]
		arcs, err := loadArcsFromFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		resp, err := client.Create(arcs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("WFST created successfully!")
		fmt.Printf("  WFST ID: %s\n", resp.WFSTID)
		fmt.Printf("  States:  %d\n", resp.NumStates)
		fmt.Printf("  Arcs:    %d\n", resp.NumArcs)
		if resp.Message != "" {
			fmt.Printf("  Message: %s\n", resp.Message)
		}

	case "search":
		if len(args) < 3 {
			printSearchUsage()
			os.Exit(1)
		}
		wfstID := args[1]
		input := args[2:]

		resp, err := client.Search(wfstID, input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if !resp.Success {
			fmt.Println("No valid path found for input sequence.")
			os.Exit(0)
		}

		fmt.Println("Best path found!")
		fmt.Printf("  Path:          %v\n", resp.Path)
		fmt.Printf("  Input Seq:     %v\n", resp.InputSeq)
		fmt.Printf("  Output Seq:    %v\n", resp.OutputSeq)
		fmt.Printf("  Weights:       %v\n", resp.Weights)
		fmt.Printf("  Total Weight:  %.4f\n", resp.TotalWeight)
		if resp.Message != "" {
			fmt.Printf("  Message:      %s\n", resp.Message)
		}

	case "compose":
		if len(args) < 3 {
			printComposeUsage()
			os.Exit(1)
		}
		wfstIDA := args[1]
		wfstIDB := args[2]

		resp, err := client.Compose(wfstIDA, wfstIDB)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("WFST composed successfully!")
		fmt.Printf("  New WFST ID: %s\n", resp.WFSTID)
		fmt.Printf("  States:      %d\n", resp.NumStates)
		fmt.Printf("  Arcs:        %d\n", resp.NumArcs)
		if resp.Message != "" {
			fmt.Printf("  Message:    %s\n", resp.Message)
		}

	case "-h", "--help", "help":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		fmt.Println()
		printUsage()
		os.Exit(1)
	}
}
