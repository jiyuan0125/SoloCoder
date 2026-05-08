package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"http-replay/api"
	"http-replay/client"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := os.Getenv("REPLAY_SERVER")
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	c := client.NewClient(serverURL)

	switch cmd {
	case "record":
		runRecord(c, args)
	case "play":
		runPlay(c, args)
	case "list":
		runList(c, args)
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func runRecord(c *client.Client, args []string) {
	fs := flag.NewFlagSet("record", flag.ExitOnError)
	targetURL := fs.String("target", "", "Target URL to proxy to (required)")

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *targetURL == "" {
		fmt.Println("Error: -target is required")
		fmt.Println("\nUsage: http-replay-client record -target <url>")
		os.Exit(1)
	}

	if err := c.RecordCommand(*targetURL); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func runPlay(c *client.Client, args []string) {
	fs := flag.NewFlagSet("play", flag.ExitOnError)
	sessionID := fs.String("session", "", "Session ID to replay (required)")
	targetURL := fs.String("target", "", "Target URL to replay to")
	mode := fs.String("mode", "strict", "Replay mode: strict or fast")
	concurrency := fs.Int("concurrency", 1, "Concurrency level for fast mode")
	replaceStr := fs.String("replace", "", "URL replacement rules, format: old=new,old2=new2")
	varsStr := fs.String("vars", "", "Variable values, format: name=value,name2=value2")

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *sessionID == "" {
		fmt.Println("Error: -session is required")
		fmt.Println("\nUsage: http-replay-client play -session <id> [options]")
		os.Exit(1)
	}

	var replaceRules []api.ReplaceRule
	if *replaceStr != "" {
		replaceRules = parseReplaceRules(*replaceStr)
	}

	var variableValues []api.VariableValue
	if *varsStr != "" {
		variableValues = parseVariableValues(*varsStr)
	}

	if err := c.PlayCommand(*sessionID, *targetURL, *mode, *concurrency, replaceRules, variableValues); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func runList(c *client.Client, args []string) {
	if err := c.ListCommand(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func parseReplaceRules(s string) []api.ReplaceRule {
	var rules []api.ReplaceRule
	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			rules = append(rules, api.ReplaceRule{
				OldPrefix: strings.TrimSpace(parts[0]),
				NewPrefix: strings.TrimSpace(parts[1]),
			})
		}
	}
	return rules
}

func parseVariableValues(s string) []api.VariableValue {
	var values []api.VariableValue
	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			values = append(values, api.VariableValue{
				Name:  strings.TrimSpace(parts[0]),
				Value: strings.TrimSpace(parts[1]),
			})
		}
	}
	return values
}

func printUsage() {
	fmt.Println("HTTP Replay Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  http-replay-client record -target <url>")
	fmt.Println("  http-replay-client play -session <id> [options]")
	fmt.Println("  http-replay-client list")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  record   Start a new recording session")
	fmt.Println("  play     Replay a recorded session")
	fmt.Println("  list     List all recorded sessions")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  REPLAY_SERVER   Server URL (default: http://localhost:8080)")
}
