package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"regex-timeout-engine/internal/protocol"
)

const defaultServer = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	server := defaultServer
	if envServer := os.Getenv("SERVER"); envServer != "" {
		server = envServer
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "compile":
		runCompile(server, args)
	case "match":
		runMatch(server, args)
	case "purge":
		runPurge(server, args)
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func runCompile(server string, args []string) {
	fs := flag.NewFlagSet("compile", flag.ExitOnError)
	pattern := fs.String("pattern", "", "regex pattern to compile")
	timeout := fs.Int("timeout", 0, "compile timeout in milliseconds (default: 5000)")
	fs.Parse(args)

	if *pattern == "" {
		fmt.Fprintln(os.Stderr, "error: -pattern is required")
		os.Exit(1)
	}

	req := protocol.CompileRequest{
		Pattern:       *pattern,
		CompileTimeMs: *timeout,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	httpReq, err := http.NewRequest(http.MethodPost, server+"/regex/compile", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result protocol.CompileResponse
	if err := json.Unmarshal(respBody, &result); err == nil {
		if result.Success {
			fmt.Println("regex_id:", result.RegexID)
			fmt.Println("pattern:", *pattern)
		} else {
			fmt.Fprintln(os.Stderr, "error:", result.Error)
			os.Exit(1)
		}
	} else {
		fmt.Fprintf(os.Stderr, "server response: %s\n", string(respBody))
		os.Exit(1)
	}
}

func runMatch(server string, args []string) {
	fs := flag.NewFlagSet("match", flag.ExitOnError)
	regexID := fs.String("id", "", "regex_id from compile command")
	text := fs.String("text", "", "text to match")
	timeout := fs.Int("timeout", 0, "match timeout in milliseconds (default: 5000)")
	fs.Parse(args)

	if *regexID == "" {
		fmt.Fprintln(os.Stderr, "error: -id is required")
		os.Exit(1)
	}
	if *text == "" {
		fmt.Fprintln(os.Stderr, "error: -text is required")
		os.Exit(1)
	}

	req := protocol.MatchRequest{
		RegexID:     *regexID,
		Text:        *text,
		MatchTimeMs: *timeout,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	httpReq, err := http.NewRequest(http.MethodPost, server+"/regex/match", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result protocol.MatchResponse
	if err := json.Unmarshal(respBody, &result); err == nil {
		if result.Success {
			fmt.Printf("found %d match(es)\n", len(result.Matches))
			for i, m := range result.Matches {
				fmt.Printf("\nmatch #%d\n", i+1)
				fmt.Printf("  text:   %q\n", m.Match)
				fmt.Printf("  start:  %d\n", m.Start)
				fmt.Printf("  end:    %d\n", m.End)
				if len(m.Groups) > 0 {
					fmt.Printf("  groups: %v\n", m.Groups)
				}
			}
		} else {
			if result.Timeout {
				fmt.Fprintln(os.Stderr, "error: match timed out")
			} else {
				fmt.Fprintln(os.Stderr, "error:", result.Error)
			}
			os.Exit(1)
		}
	} else {
		fmt.Fprintf(os.Stderr, "server response: %s\n", string(respBody))
		os.Exit(1)
	}
}

func runPurge(server string, args []string) {
	_ = args

	httpReq, err := http.NewRequest(http.MethodDelete, server+"/regex/cache", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result protocol.PurgeResponse
	if err := json.Unmarshal(respBody, &result); err == nil {
		if result.Success {
			fmt.Printf("purged %d entries from cache\n", result.Purged)
		} else {
			fmt.Fprintln(os.Stderr, "error:", result.Error)
			os.Exit(1)
		}
	} else {
		fmt.Fprintf(os.Stderr, "server response: %s\n", string(respBody))
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Usage: client <command> [options]

Commands:
  compile   Compile a regex pattern
  match     Execute match using a compiled regex
  purge     Clear all cached regex patterns
  help      Show this help

Examples:
  client compile -pattern '\d+'
  client compile -pattern '\d+' -timeout 2000
  client match -id <regex_id> -text 'abc123def456'
  client match -id <regex_id> -text 'abc123def456' -timeout 1000
  client purge

Environment:
  SERVER    Server URL (default: http://localhost:8080)`)
}
