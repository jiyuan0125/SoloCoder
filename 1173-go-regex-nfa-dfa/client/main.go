package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regex-engine/api"
	"strings"
)

type Client struct {
	BaseURL string
}

func NewClient(baseURL string) *Client {
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	return &Client{BaseURL: baseURL}
}

func (c *Client) postJSON(path string, req interface{}, resp interface{}) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpResp, err := http.Post(c.BaseURL+path, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if err := json.Unmarshal(respBody, resp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w, body: %s", err, string(respBody))
	}

	return nil
}

func (c *Client) Compile(pattern string) error {
	req := api.CompileRequest{Pattern: pattern}
	resp := api.CompileResponse{}

	if err := c.postJSON("/compile", &req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("compile error: %s", resp.Error)
	}

	fmt.Printf("Pattern: %s\n", resp.Pattern)
	fmt.Printf("NFA States:     %d\n", resp.NFAStates)
	fmt.Printf("DFA States:     %d\n", resp.DFAStates)
	fmt.Printf("Minimized DFA:  %d\n", resp.MinDFAStates)

	return nil
}

func (c *Client) Match(pattern, text string) error {
	req := api.MatchRequest{Pattern: pattern, Text: text}
	resp := api.MatchResponse{}

	if err := c.postJSON("/match", &req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("match error: %s", resp.Error)
	}

	if resp.Match {
		fmt.Printf("MATCH: \"%s\" matches pattern \"%s\"\n", text, pattern)
	} else {
		fmt.Printf("NO MATCH: \"%s\" does not match pattern \"%s\"\n", text, pattern)
	}

	return nil
}

func (c *Client) Viz(pattern string) error {
	req := api.VizRequest{Pattern: pattern}
	resp := api.VizResponse{}

	if err := c.postJSON("/viz", &req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("viz error: %s", resp.Error)
	}

	fmt.Printf("=== Full Visualization for: %s ===\n\n", resp.Pattern)

	fmt.Printf("--- AST ---\n%s\n\n", resp.AST)

	fmt.Printf("--- NFA (%d states) ---\n", resp.NFAStates)
	fmt.Printf("Start: %d, Finals: %v\n", resp.NFA.StartState, resp.NFA.FinalStates)
	for _, t := range resp.NFA.Transitions {
		fmt.Printf("  %d --[%s]--> %v\n", t.From, t.Symbol, t.To)
	}
	fmt.Println()

	fmt.Printf("--- DFA (%d states) ---\n", resp.DFAStates)
	fmt.Printf("Start: %d, Finals: %v\n", resp.DFA.StartState, resp.DFA.FinalStates)
	for _, t := range resp.DFA.Transitions {
		fmt.Printf("  %d --[%s]--> %d\n", t.From, t.Symbol, t.To)
	}
	fmt.Println()

	fmt.Printf("--- Minimized DFA (%d states) ---\n", resp.MinDFAStates)
	fmt.Printf("Start: %d, Finals: %v\n", resp.MinimizedDFA.StartState, resp.MinimizedDFA.FinalStates)
	for _, t := range resp.MinimizedDFA.Transitions {
		fmt.Printf("  %d --[%s]--> %d\n", t.From, t.Symbol, t.To)
	}
	fmt.Println()

	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, `Regex Engine Client

Usage:
  %s compile <pattern>               - Compile regex and show state counts
  %s match <pattern> <text>          - Test if text matches pattern
  %s viz <pattern>                   - Show full visualization (AST, NFA, DFA, min-DFA)

Options:
  -server <url>                      - Server URL (default: http://localhost:8200)
  -help                              - Show this help
`, os.Args[0], os.Args[0], os.Args[0])
}

func main() {
	serverURL := flag.String("server", "http://localhost:8200", "Server URL")
	help := flag.Bool("help", false, "Show help")
	flag.Parse()

	if *help || flag.NArg() < 1 {
		usage()
		os.Exit(0)
	}

	cmd := flag.Arg(0)
	client := NewClient(*serverURL)

	var err error
	switch cmd {
	case "compile":
		if flag.NArg() < 2 {
			fmt.Fprintln(os.Stderr, "error: compile requires a pattern")
			usage()
			os.Exit(1)
		}
		err = client.Compile(flag.Arg(1))

	case "match":
		if flag.NArg() < 3 {
			fmt.Fprintln(os.Stderr, "error: match requires pattern and text")
			usage()
			os.Exit(1)
		}
		err = client.Match(flag.Arg(1), flag.Arg(2))

	case "viz":
		if flag.NArg() < 2 {
			fmt.Fprintln(os.Stderr, "error: viz requires a pattern")
			usage()
			os.Exit(1)
		}
		err = client.Viz(flag.Arg(1))

	default:
		fmt.Fprintf(os.Stderr, "error: unknown command '%s'\n", cmd)
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
