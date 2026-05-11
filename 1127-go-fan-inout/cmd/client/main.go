package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"1127-go-fan-inout/pkg/api"
)

const serverURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "add-source":
		handleAddSource(args)
	case "remove-source":
		handleRemoveSource(args)
	case "set-policy":
		handleSetPolicy(args)
	case "flow":
		handleFlow()
	case "result":
		handleResult()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Fan-in Fan-out Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client add-source --name=<name> [--interval=<ms>] [--payload=<text>]")
	fmt.Println("  client remove-source --name=<name>")
	fmt.Println("  client set-policy --policy=<block|drop>")
	fmt.Println("  client flow")
	fmt.Println("  client result")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  add-source     Add a new data source")
	fmt.Println("  remove-source  Remove an existing data source")
	fmt.Println("  set-policy     Set backpressure policy")
	fmt.Println("  flow           View real-time traffic stats")
	fmt.Println("  result         View aggregated output")
}

func handleAddSource(args []string) {
	fs := flag.NewFlagSet("add-source", flag.ExitOnError)
	name := fs.String("name", "", "Source name (required)")
	interval := fs.Int("interval", 1000, "Interval in milliseconds")
	payload := fs.String("payload", "", "Payload prefix")

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *name == "" {
		fmt.Println("Error: --name is required")
		fs.Usage()
		os.Exit(1)
	}

	req := api.AddSourceRequest{
		Name:       *name,
		IntervalMs: *interval,
		Payload:    *payload,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/api/source/add", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result api.AddSourceResponse
	json.Unmarshal(data, &result)

	if resp.StatusCode != http.StatusOK || !result.Success {
		fmt.Printf("Error: %s\n", result.Message)
		os.Exit(1)
	}

	fmt.Printf("Success: %s\n", result.Message)
}

func handleRemoveSource(args []string) {
	fs := flag.NewFlagSet("remove-source", flag.ExitOnError)
	name := fs.String("name", "", "Source name (required)")

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *name == "" {
		fmt.Println("Error: --name is required")
		fs.Usage()
		os.Exit(1)
	}

	req := api.RemoveSourceRequest{Name: *name}
	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/api/source/remove", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result api.RemoveSourceResponse
	json.Unmarshal(data, &result)

	if resp.StatusCode != http.StatusOK || !result.Success {
		fmt.Printf("Error: %s\n", result.Message)
		os.Exit(1)
	}

	fmt.Printf("Success: %s\n", result.Message)
}

func handleSetPolicy(args []string) {
	fs := flag.NewFlagSet("set-policy", flag.ExitOnError)
	policy := fs.String("policy", "", "Policy: block or drop (required)")

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *policy == "" {
		fmt.Println("Error: --policy is required")
		fs.Usage()
		os.Exit(1)
	}

	p := api.BackpressurePolicy(strings.ToLower(*policy))
	if p != api.PolicyBlock && p != api.PolicyDrop {
		fmt.Println("Error: policy must be 'block' or 'drop'")
		os.Exit(1)
	}

	req := api.SetPolicyRequest{Policy: p}
	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/api/policy", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result api.SetPolicyResponse
	json.Unmarshal(data, &result)

	if resp.StatusCode != http.StatusOK || !result.Success {
		fmt.Printf("Error: %s\n", result.Message)
		os.Exit(1)
	}

	fmt.Printf("Policy set to: %s\n", result.Policy)
}

func handleFlow() {
	resp, err := http.Get(serverURL + "/api/flow")
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result api.FlowResponse
	json.Unmarshal(data, &result)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: server returned status %d\n", resp.StatusCode)
		os.Exit(1)
	}

	if len(result.Sources) == 0 {
		fmt.Println("No active sources")
		return
	}

	fmt.Printf("%-20s %12s %12s %10s %10s %8s\n",
		"SOURCE", "RECEIVED", "SENT", "DROPPED", "BACKLOG", "ACTIVE")
	fmt.Println(strings.Repeat("-", 75))

	for _, s := range result.Sources {
		active := "NO"
		if s.IsActive {
			active = "YES"
		}
		fmt.Printf("%-20s %12d %12d %10d %10d %8s\n",
			s.SourceName, s.TotalReceived, s.TotalSent, s.Dropped, s.CurrentBacklog, active)
	}
}

func handleResult() {
	resp, err := http.Get(serverURL + "/api/result")
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result api.ResultResponse
	json.Unmarshal(data, &result)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: server returned status %d\n", resp.StatusCode)
		os.Exit(1)
	}

	if len(result.Items) == 0 {
		fmt.Println("No aggregated results yet")
		return
	}

	fmt.Printf("%-20s %-30s %s\n", "SOURCE", "PAYLOAD", "TIMESTAMP")
	fmt.Println(strings.Repeat("-", 80))

	for _, item := range result.Items {
		ts := time.Unix(0, item.Timestamp).Format("15:04:05.000")
		fmt.Printf("%-20s %-30s %s\n", item.Source, truncate(item.Payload, 30), ts)
	}

	fmt.Printf("\nTotal items: %d\n", len(result.Items))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return strings.Repeat(".", max)
	}
	return s[:max-3] + "..."
}

func strconvInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}
