package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"rk-matcher/pkg/api"
)

const (
	defaultServerURL = "http://localhost:8402"
	version          = "1.0.0"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

func (c *Client) post(endpoint string, body interface{}, result interface{}) error {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.baseURL+endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (c *Client) get(endpoint string, result interface{}) error {
	resp, err := http.Get(c.baseURL + endpoint)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) Add(pattern string) error {
	req := api.AddRequest{Pattern: pattern}
	var resp api.AddResponse
	if err := c.post("/add", req, &resp); err != nil {
		return err
	}

	if resp.Success {
		fmt.Printf("Pattern added: %q\n", pattern)
	} else {
		fmt.Printf("Failed to add pattern: %s\n", resp.Message)
	}
	return nil
}

func (c *Client) Import(patterns []string) error {
	req := api.ImportRequest{Patterns: patterns}
	var resp api.ImportResponse
	if err := c.post("/import", req, &resp); err != nil {
		return err
	}

	fmt.Printf("Import complete: added %d, skipped %d\n", resp.Added, resp.Skipped)
	return nil
}

func (c *Client) Search(text string) error {
	req := api.SearchRequest{Text: text}
	var resp api.SearchResponse
	if err := c.post("/search", req, &resp); err != nil {
		return err
	}

	fmt.Printf("Found %d matches:\n", resp.Count)
	for i, m := range resp.Matches {
		fmt.Printf("  %d. Pattern: %q at index %d\n", i+1, m.Pattern, m.Index)
	}
	return nil
}

func (c *Client) Hash(text string, length int) error {
	req := api.HashRequest{Text: text, Length: length}
	var resp api.HashResponse
	if err := c.post("/hash", req, &resp); err != nil {
		return err
	}

	if resp.Success {
		fmt.Printf("Hash of %q", text)
		if length > 0 {
			fmt.Printf(" (first %d chars)", length)
		}
		fmt.Printf(": %s\n", resp.Hash)
	} else {
		fmt.Printf("Failed: %s\n", resp.Message)
	}
	return nil
}

func (c *Client) ListPatterns() error {
	var resp api.PatternsResponse
	if err := c.get("/patterns", &resp); err != nil {
		return err
	}

	fmt.Printf("Total patterns: %d\n", resp.Total)
	if resp.Total == 0 {
		return nil
	}

	fmt.Println("\nPatterns by length:")
	for length, patterns := range resp.ByLength {
		fmt.Printf("  Length %d (%d patterns):\n", length, len(patterns))
		for _, p := range patterns {
			fmt.Printf("    - %q\n", p)
		}
	}
	return nil
}

func (c *Client) ShowStats() error {
	var resp api.CollisionStatsResponse
	if err := c.get("/stats", &resp); err != nil {
		return err
	}

	fmt.Printf("Collision Statistics:\n")
	fmt.Printf("  Total hash checks:    %d\n", resp.TotalChecks)
	fmt.Printf("  False positives:      %d\n", resp.FalsePositives)
	if resp.TotalChecks > 0 {
		rate := float64(resp.FalsePositives) / float64(resp.TotalChecks) * 100
		fmt.Printf("  False positive rate:  %.4f%%\n", rate)
	}
	if len(resp.HashBuckets) > 0 {
		fmt.Printf("  Hash buckets with collisions: %d\n", len(resp.HashBuckets))
		for hash, count := range resp.HashBuckets {
			fmt.Printf("    - %s: %d patterns\n", hash, count)
		}
	}
	return nil
}

func (c *Client) ShowStatus() error {
	var resp api.StatusResponse
	if err := c.get("/status", &resp); err != nil {
		return err
	}

	fmt.Printf("Rabin-Karp Matcher Server v%s\n", resp.Version)
	fmt.Printf("Total patterns: %d\n", resp.Patterns.Total)
	fmt.Printf("\nCollision Stats:\n")
	fmt.Printf("  Total checks:    %d\n", resp.Stats.TotalChecks)
	fmt.Printf("  False positives: %d\n", resp.Stats.FalsePositives)
	return nil
}

func readPatternsFromFile(filename string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var patterns []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			patterns = append(patterns, line)
		}
	}
	return patterns, nil
}

func printUsage() {
	fmt.Println("Rabin-Karp Matcher Client v" + version)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client [global-flags] <command> [command-flags] [args]")
	fmt.Println()
	fmt.Println("Global Flags:")
	fmt.Println("  -server URL    Server URL (default: " + defaultServerURL + ")")
	fmt.Println("  -s URL         Server URL (shorthand)")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  add <pattern>           Add a single pattern")
	fmt.Println("  import <file|patterns>  Import patterns from file or comma-separated list")
	fmt.Println("  search <text>           Search text for matching patterns")
	fmt.Println("  hash <text> [length]    Compute hash of a string (debug)")
	fmt.Println("  list                    List all registered patterns")
	fmt.Println("  stats                   Show collision statistics")
	fmt.Println("  status                  Show server status")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  client add \"hello\"")
	fmt.Println("  client import patterns.txt")
	fmt.Println("  client import \"hello,world,test\"")
	fmt.Println("  client search \"hello world test\"")
	fmt.Println("  client hash \"hello\"")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := defaultServerURL
	if envURL := os.Getenv("SERVER_URL"); envURL != "" {
		serverURL = envURL
	}

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		if args[i] == "-server" || args[i] == "-s" {
			if i+1 < len(args) {
				serverURL = args[i+1]
				args = append(args[:i], args[i+2:]...)
				i--
			} else {
				fmt.Fprintln(os.Stderr, "Error: -server flag requires an argument")
				os.Exit(1)
			}
		}
	}

	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	client := NewClient(serverURL)
	command := args[0]

	switch command {
	case "add":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: add command requires a pattern")
			os.Exit(1)
		}
		if err := client.Add(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "import":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: import command requires a file or patterns")
			os.Exit(1)
		}
		input := args[1]
		var patterns []string
		var err error

		if _, statErr := os.Stat(input); statErr == nil {
			patterns, err = readPatternsFromFile(input)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
				os.Exit(1)
			}
		} else {
			for _, p := range strings.Split(input, ",") {
				p = strings.TrimSpace(p)
				if p != "" {
					patterns = append(patterns, p)
				}
			}
		}

		if len(patterns) == 0 {
			fmt.Fprintln(os.Stderr, "Error: no patterns to import")
			os.Exit(1)
		}

		if err := client.Import(patterns); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "search":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: search command requires text")
			os.Exit(1)
		}
		if err := client.Search(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "hash":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: hash command requires text")
			os.Exit(1)
		}
		length := 0
		if len(args) >= 3 {
			var err error
			length, err = strconv.Atoi(args[2])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid length: %v\n", err)
				os.Exit(1)
			}
		}
		if err := client.Hash(args[1], length); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "list":
		if err := client.ListPatterns(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "stats":
		if err := client.ShowStats(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "status":
		if err := client.ShowStatus(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "help", "-h", "--help":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command %q\n", command)
		fmt.Fprintln(os.Stderr, "Use 'client help' for usage information")
		os.Exit(1)
	}
}
