package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"idgen/api"
)

const (
	defaultServer = "http://localhost:8080"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverAddr := getServerAddr()

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "register":
		handleRegister(serverAddr, args)
	case "unregister":
		handleUnregister(serverAddr, args)
	case "generate":
		handleGenerate(serverAddr, args)
	case "batch-generate":
		handleBatchGenerate(serverAddr, args)
	case "list-nodes":
		handleListNodes(serverAddr, args)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func getServerAddr() string {
	if env := os.Getenv("IDGEN_SERVER"); env != "" {
		return env
	}
	return defaultServer
}

func printUsage() {
	fmt.Println("Usage: client <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  register       Register a new node")
	fmt.Println("  unregister     Unregister a node")
	fmt.Println("  generate       Generate a single ID")
	fmt.Println("  batch-generate Generate multiple IDs")
	fmt.Println("  list-nodes     List all registered nodes")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  IDGEN_SERVER   Server address (default: http://localhost:8080)")
}

func handleRegister(server string, args []string) {
	fs := flag.NewFlagSet("register", flag.ExitOnError)
	name := fs.String("name", "", "Node name (required)")
	remark := fs.String("remark", "", "Optional remark")
	fs.Parse(args)

	if *name == "" {
		fmt.Fprintln(os.Stderr, "error: -name is required")
		os.Exit(1)
	}

	req := api.RegisterNodeRequest{
		Name:   *name,
		Remark: *remark,
	}

	var resp api.RegisterNodeResponse
	if err := postJSON(server+"/register", req, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "error: %s\n", resp.Message)
		os.Exit(1)
	}

	fmt.Println(resp.NodeID)
}

func handleUnregister(server string, args []string) {
	fs := flag.NewFlagSet("unregister", flag.ExitOnError)
	nodeID := fs.Uint("node-id", 0, "Node ID (required)")
	fs.Parse(args)

	if *nodeID == 0 {
		fmt.Fprintln(os.Stderr, "error: -node-id is required")
		os.Exit(1)
	}

	req := api.UnregisterNodeRequest{
		NodeID: uint16(*nodeID),
	}

	var resp api.UnregisterNodeResponse
	if err := postJSON(server+"/unregister", req, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "error: %s\n", resp.Message)
		os.Exit(1)
	}

	fmt.Println("ok")
}

func handleGenerate(server string, args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	nodeID := fs.Uint("node-id", 0, "Node ID (required)")
	fs.Parse(args)

	if *nodeID == 0 {
		fmt.Fprintln(os.Stderr, "error: -node-id is required")
		os.Exit(1)
	}

	req := api.GenerateIDRequest{
		NodeID: uint16(*nodeID),
	}

	var resp api.GenerateIDResponse
	if err := postJSON(server+"/generate", req, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "error: %s\n", resp.Message)
		os.Exit(1)
	}

	fmt.Println(resp.ID)
}

func handleBatchGenerate(server string, args []string) {
	fs := flag.NewFlagSet("batch-generate", flag.ExitOnError)
	nodeID := fs.Uint("node-id", 0, "Node ID (required)")
	count := fs.Int("count", 0, "Number of IDs to generate (1-1000, required)")
	fs.Parse(args)

	if *nodeID == 0 {
		fmt.Fprintln(os.Stderr, "error: -node-id is required")
		os.Exit(1)
	}

	if *count <= 0 || *count > api.MaxBatchCount {
		fmt.Fprintf(os.Stderr, "error: -count must be between 1 and %d\n", api.MaxBatchCount)
		os.Exit(1)
	}

	req := api.BatchGenerateRequest{
		NodeID: uint16(*nodeID),
		Count:  *count,
	}

	var resp api.BatchGenerateResponse
	if err := postJSON(server+"/batch-generate", req, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "error: %s\n", resp.Message)
		os.Exit(1)
	}

	for _, id := range resp.IDs {
		fmt.Println(id)
	}
}

func handleListNodes(server string, args []string) {
	fs := flag.NewFlagSet("list-nodes", flag.ExitOnError)
	fs.Parse(args)

	resp, err := http.Get(server + "/nodes")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading response: %v\n", err)
		os.Exit(1)
	}

	var listResp api.ListNodesResponse
	if err := json.Unmarshal(body, &listResp); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing response: %v\n", err)
		os.Exit(1)
	}

	if !listResp.Success {
		fmt.Fprintf(os.Stderr, "error: %s\n", listResp.Message)
		os.Exit(1)
	}

	if len(listResp.Nodes) == 0 {
		fmt.Println("No nodes registered.")
		return
	}

	fmt.Printf("%-8s %-20s %s\n", "NODE_ID", "NAME", "REMARK")
	for _, node := range listResp.Nodes {
		remark := node.Remark
		if remark == "" {
			remark = "-"
		}
		fmt.Printf("%-8d %-20s %s\n", node.NodeID, node.Name, remark)
	}
}

func postJSON(url string, req interface{}, resp interface{}) error {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpResp, err := http.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if httpResp.StatusCode >= 400 {
		var errResp api.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Message != "" {
			return fmt.Errorf("%s", errResp.Message)
		}
		return fmt.Errorf("server returned status %d: %s", httpResp.StatusCode, string(respBody))
	}

	if err := json.Unmarshal(respBody, resp); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	return nil
}
