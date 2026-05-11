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

	"vectorclock/pkg/api"
	"vectorclock/pkg/vectorclock"
)

var serverURL string

func main() {
	flag.StringVar(&serverURL, "server", "http://localhost:8080", "Server URL (default: http://localhost:8080)")
	flag.Parse()

	if envURL := os.Getenv("VC_SERVER"); envURL != "" {
		serverURL = envURL
	}

	if len(flag.Args()) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := flag.Args()[0]
	switch command {
	case "produce":
		handleProduce(flag.Args()[1:])
	case "receive":
		handleReceive(flag.Args()[1:])
	case "clock":
		handleClock(flag.Args()[1:])
	case "compare":
		handleCompare(flag.Args()[1:])
	case "events":
		handleEvents(flag.Args()[1:])
	case "prune":
		handlePrune(flag.Args()[1:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Vector Clock Client

Usage: vc [command] [options]

Commands:
  produce -node=<node_id> [-content=<content>]
    Produce a local event on a node

  receive -node=<node_id> -from=<from_node_id> -clock=<json_clock> [-content=<content>]
    Receive and merge an event from another node

  clock -node=<node_id>
    Query the current clock of a node

  compare -clock_a=<json_clock> -clock_b=<json_clock>
    Compare two vector clocks for causality

  events -node=<node_id>
    Get all events from a node

  prune -node=<node_id> [-retain=<count>]
    Prune the clock history of a node

Global options:
  -server=<url>
    Server URL (default: http://localhost:8080)
`)
}

func handleProduce(args []string) {
	fs := flag.NewFlagSet("produce", flag.ExitOnError)
	nodeID := fs.String("node", "", "Node ID")
	content := fs.String("content", "local event", "Event content")
	fs.Parse(args)

	if *nodeID == "" {
		fmt.Println("Error: -node is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	req := &api.ProduceRequest{
		NodeID:  *nodeID,
		Content: *content,
	}

	var resp api.ProduceResponse
	if err := postJSON("/produce", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Event produced: ID=%d\n", resp.EventID)
	fmt.Printf("Clock: %v\n", resp.Clock)
}

func handleReceive(args []string) {
	fs := flag.NewFlagSet("receive", flag.ExitOnError)
	nodeID := fs.String("node", "", "Target Node ID")
	fromNodeID := fs.String("from", "", "Source Node ID")
	clockStr := fs.String("clock", "", "Remote clock JSON")
	content := fs.String("content", "received event", "Event content")
	fs.Parse(args)

	if *nodeID == "" || *fromNodeID == "" || *clockStr == "" {
		fmt.Println("Error: -node, -from, and -clock are required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	var remoteClock vectorclock.VectorClock
	if err := json.Unmarshal([]byte(*clockStr), &remoteClock); err != nil {
		fmt.Printf("Error parsing clock JSON: %v\n", err)
		os.Exit(1)
	}

	req := &api.ReceiveRequest{
		TargetNodeID: *nodeID,
		FromNodeID:   *fromNodeID,
		RemoteClock:  remoteClock,
		Content:      *content,
	}

	var resp api.ReceiveResponse
	if err := postJSON("/receive", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Event received: ID=%d\n", resp.EventID)
	fmt.Printf("Merged clock: %v\n", resp.Clock)
}

func handleClock(args []string) {
	fs := flag.NewFlagSet("clock", flag.ExitOnError)
	nodeID := fs.String("node", "", "Node ID")
	fs.Parse(args)

	if *nodeID == "" {
		fmt.Println("Error: -node is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	var resp api.ClockResponse
	if err := getJSON("/clock?node_id="+*nodeID, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Node: %s\n", resp.NodeID)
	fmt.Printf("Current clock: %v\n", resp.Clock)
}

func handleCompare(args []string) {
	fs := flag.NewFlagSet("compare", flag.ExitOnError)
	clockAStr := fs.String("clock_a", "", "Clock A JSON")
	clockBStr := fs.String("clock_b", "", "Clock B JSON")
	fs.Parse(args)

	if *clockAStr == "" || *clockBStr == "" {
		fmt.Println("Error: -clock_a and -clock_b are required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	var clockA, clockB vectorclock.VectorClock
	if err := json.Unmarshal([]byte(*clockAStr), &clockA); err != nil {
		fmt.Printf("Error parsing clock_a JSON: %v\n", err)
		os.Exit(1)
	}
	if err := json.Unmarshal([]byte(*clockBStr), &clockB); err != nil {
		fmt.Printf("Error parsing clock_b JSON: %v\n", err)
		os.Exit(1)
	}

	req := &api.CompareRequest{
		ClockA: clockA,
		ClockB: clockB,
	}

	var resp api.CompareResponse
	if err := postJSON("/compare", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Result: %s\n", resp.Result)
}

func handleEvents(args []string) {
	fs := flag.NewFlagSet("events", flag.ExitOnError)
	nodeID := fs.String("node", "", "Node ID")
	fs.Parse(args)

	if *nodeID == "" {
		fmt.Println("Error: -node is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	var resp api.EventsResponse
	if err := getJSON("/events?node_id="+*nodeID, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Node: %s, Total events: %d\n", resp.NodeID, len(resp.Events))
	for i, event := range resp.Events {
		fmt.Printf("\nEvent %d:\n", i+1)
		fmt.Printf("  ID: %d\n", event.EventID)
		fmt.Printf("  Node: %s\n", event.NodeID)
		fmt.Printf("  Clock: %v\n", event.Clock)
		fmt.Printf("  Content: %s\n", event.Content)
		fmt.Printf("  Time: %s\n", event.CreationTime.Format("2006-01-02 15:04:05"))
	}
}

func handlePrune(args []string) {
	fs := flag.NewFlagSet("prune", flag.ExitOnError)
	nodeID := fs.String("node", "", "Node ID")
	retainCount := fs.Int("retain", 10, "Number of active nodes to retain")
	fs.Parse(args)

	if *nodeID == "" {
		fmt.Println("Error: -node is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	req := &api.PruneRequest{
		NodeID:      *nodeID,
		RetainCount: *retainCount,
	}

	var resp api.PruneResponse
	if err := postJSON("/prune", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Node: %s\n", resp.NodeID)
	fmt.Printf("Pruned dimensions: %d\n", resp.PrunedCount)
}

func postJSON(endpoint string, req interface{}, resp interface{}) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpResp, err := http.Post(strings.TrimRight(serverURL, "/")+endpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if httpResp.StatusCode >= 400 {
		var errResp api.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error != "" {
			return fmt.Errorf("server error: %s", errResp.Error)
		}
		return fmt.Errorf("server returned status %d: %s", httpResp.StatusCode, string(respBody))
	}

	if err := json.Unmarshal(respBody, resp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

func getJSON(endpoint string, resp interface{}) error {
	httpResp, err := http.Get(strings.TrimRight(serverURL, "/") + endpoint)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if httpResp.StatusCode >= 400 {
		var errResp api.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error != "" {
			return fmt.Errorf("server error: %s", errResp.Error)
		}
		return fmt.Errorf("server returned status %d: %s", httpResp.StatusCode, string(respBody))
	}

	if err := json.Unmarshal(respBody, resp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
