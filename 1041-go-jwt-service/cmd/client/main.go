package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

const (
	usage = `JWT Token Service Client

Usage:
  jwt-client <command> [options]

Commands:
  issue     - Issue a new token pair
  refresh   - Refresh token pair using refresh token
  validate  - Validate a token
  revoke    - Revoke a token by token ID

Global Options:
  -u, --url     Server base URL (default: http://localhost:8080)
  -h, --help    Show this help message

Examples:
  jwt-client issue --user-id user123 --claims '{"role":"admin","permissions":["read","write"]}'
  jwt-client refresh --refresh-token <token>
  jwt-client validate --token <token>
  jwt-client revoke --token-id <jti>
`
)

func main() {
	if len(os.Args) < 2 {
		printUsageAndExit("command is required")
	}

	cmd := os.Args[1]

	switch cmd {
	case "issue":
		handleIssue()
	case "refresh":
		handleRefresh()
	case "validate":
		handleValidate()
	case "revoke":
		handleRevoke()
	case "-h", "--help", "help":
		fmt.Print(usage)
		os.Exit(0)
	default:
		printUsageAndExit(fmt.Sprintf("unknown command: %s", cmd))
	}
}

func printUsageAndExit(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, "Error: %s\n\n", msg)
	}
	fmt.Fprint(os.Stderr, usage)
	os.Exit(1)
}

func parseCommonFlags(fs *flag.FlagSet) string {
	var serverURL string
	fs.StringVar(&serverURL, "url", "http://localhost:8080", "Server base URL")
	fs.StringVar(&serverURL, "u", "http://localhost:8080", "Server base URL (short)")
	return serverURL
}

func handleIssue() {
	fs := flag.NewFlagSet("issue", flag.ExitOnError)
	serverURL := parseCommonFlags(fs)

	var userID, claimsJSON string
	fs.StringVar(&userID, "user-id", "", "User ID (required)")
	fs.StringVar(&claimsJSON, "claims", "", "Custom claims as JSON")

	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	if userID == "" {
		printUsageAndExit("--user-id is required")
	}

	var claims map[string]any
	if claimsJSON != "" {
		if err := json.Unmarshal([]byte(claimsJSON), &claims); err != nil {
			printUsageAndExit(fmt.Sprintf("invalid claims JSON: %v", err))
		}
	}

	client := NewClient(serverURL)
	resp, err := client.Issue(userID, claims)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	printJSON(resp)
}

func handleRefresh() {
	fs := flag.NewFlagSet("refresh", flag.ExitOnError)
	serverURL := parseCommonFlags(fs)

	var refreshToken string
	fs.StringVar(&refreshToken, "refresh-token", "", "Refresh token (required)")
	fs.StringVar(&refreshToken, "r", "", "Refresh token (short)")

	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	if refreshToken == "" {
		printUsageAndExit("--refresh-token is required")
	}

	client := NewClient(serverURL)
	resp, err := client.Refresh(refreshToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	printJSON(resp)
}

func handleValidate() {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	serverURL := parseCommonFlags(fs)

	var token string
	fs.StringVar(&token, "token", "", "Token to validate (required)")
	fs.StringVar(&token, "t", "", "Token to validate (short)")

	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	if token == "" {
		printUsageAndExit("--token is required")
	}

	client := NewClient(serverURL)
	resp, err := client.Validate(token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	printJSON(resp)
}

func handleRevoke() {
	fs := flag.NewFlagSet("revoke", flag.ExitOnError)
	serverURL := parseCommonFlags(fs)

	var tokenID string
	fs.StringVar(&tokenID, "token-id", "", "Token ID (jti) to revoke (required)")
	fs.StringVar(&tokenID, "id", "", "Token ID (jti) to revoke (short)")

	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	if tokenID == "" {
		printUsageAndExit("--token-id is required")
	}

	client := NewClient(serverURL)
	resp, err := client.Revoke(tokenID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	printJSON(resp)
}

func printJSON(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(sortJSONKeys(string(data)))
}

func sortJSONKeys(jsonStr string) string {
	var m map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		return jsonStr
	}

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		valJSON, _ := json.MarshalIndent(m[k], "  ", "  ")
		valStr := strings.TrimSpace(string(valJSON))
		parts = append(parts, fmt.Sprintf("  %q: %s", k, valStr))
	}

	return "{\n" + strings.Join(parts, ",\n") + "\n}"
}
