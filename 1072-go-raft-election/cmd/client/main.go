package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"raft-sim/internal/api"
)

const baseURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "create":
		createCluster(args)
	case "add-node":
		addNode(args)
	case "list":
		listNodes()
	case "status":
		getNodeStatus(args)
	case "submit":
		submitLog(args)
	case "elect":
		triggerElection()
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Raft Election Simulator Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  raft-client <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create <node_count>       Create a cluster with N nodes (min 3)")
	fmt.Println("  add-node <node_id>        Add a new node to the cluster")
	fmt.Println("  list                      List all nodes in the cluster")
	fmt.Println("  status <node_id>          Get detailed status of a specific node")
	fmt.Println("  submit \"<command>\"        Submit a log entry to the leader")
	fmt.Println("  elect                     Trigger an election")
	fmt.Println("  help                      Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  raft-client create 3")
	fmt.Println("  raft-client list")
	fmt.Println("  raft-client submit \"set key=value\"")
	fmt.Println("  raft-client status 1")
}

func createCluster(args []string) {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Println("Error: node_count is required")
		fmt.Println("Usage: raft-client create <node_count>")
		os.Exit(1)
	}

	nodeCount, err := strconv.Atoi(fs.Arg(0))
	if err != nil {
		fmt.Printf("Error: invalid node_count: %s\n", fs.Arg(0))
		os.Exit(1)
	}

	if nodeCount < 3 {
		fmt.Println("Error: cluster must have at least 3 nodes")
		os.Exit(1)
	}

	req := api.CreateClusterRequest{NodeCount: nodeCount}
	var resp api.CreateClusterResponse
	if err := postJSON("/api/cluster/create", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("Success: %s\n", resp.Message)
		fmt.Printf("Created nodes: %v\n", resp.Nodes)
	} else {
		fmt.Printf("Error: %s\n", resp.Message)
		os.Exit(1)
	}
}

func addNode(args []string) {
	fs := flag.NewFlagSet("add-node", flag.ExitOnError)
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Println("Error: node_id is required")
		fmt.Println("Usage: raft-client add-node <node_id>")
		os.Exit(1)
	}

	nodeID, err := strconv.Atoi(fs.Arg(0))
	if err != nil {
		fmt.Printf("Error: invalid node_id: %s\n", fs.Arg(0))
		os.Exit(1)
	}

	req := api.AddNodeRequest{NodeID: nodeID}
	var resp api.AddNodeResponse
	if err := postJSON("/api/cluster/add-node", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("Success: %s\n", resp.Message)
	} else {
		fmt.Printf("Error: %s\n", resp.Message)
		os.Exit(1)
	}
}

func listNodes() {
	var resp api.ListNodesResponse
	if err := getJSON("/api/cluster/nodes", &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Message)
		os.Exit(1)
	}

	if len(resp.Nodes) == 0 {
		fmt.Println("No nodes in cluster. Create one first with: raft-client create <count>")
		return
	}

	sort.Slice(resp.Nodes, func(i, j int) bool {
		return resp.Nodes[i].ID < resp.Nodes[j].ID
	})

	fmt.Printf("Cluster nodes (%d):\n\n", len(resp.Nodes))
	fmt.Printf("%-6s %-12s %-8s %-10s %-15s\n", "ID", "STATE", "TERM", "VOTED_FOR", "COMMIT_INDEX")
	fmt.Println(strings.Repeat("-", 55))

	for _, node := range resp.Nodes {
		votedFor := "-"
		if node.VotedFor >= 0 {
			votedFor = strconv.Itoa(node.VotedFor)
		}
		commitIndex := "-"
		if node.CommitIndex >= 0 {
			commitIndex = strconv.Itoa(node.CommitIndex)
		}
		fmt.Printf("%-6d %-12s %-8d %-10s %-15s\n",
			node.ID, node.State, node.CurrentTerm, votedFor, commitIndex)
	}
}

func getNodeStatus(args []string) {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Println("Error: node_id is required")
		fmt.Println("Usage: raft-client status <node_id>")
		os.Exit(1)
	}

	nodeID, err := strconv.Atoi(fs.Arg(0))
	if err != nil {
		fmt.Printf("Error: invalid node_id: %s\n", fs.Arg(0))
		os.Exit(1)
	}

	var resp api.GetNodeResponse
	if err := getJSON(fmt.Sprintf("/api/cluster/nodes/%d", nodeID), &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Message)
		os.Exit(1)
	}

	node := resp.Node
	fmt.Printf("Node #%d Status:\n", node.ID)
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("  State:           %s\n", node.State)
	fmt.Printf("  Current Term:    %d\n", node.CurrentTerm)

	votedFor := "None"
	if node.VotedFor >= 0 {
		votedFor = strconv.Itoa(node.VotedFor)
	}
	fmt.Printf("  Voted For:       %s\n", votedFor)

	leaderID := "Unknown"
	if node.LeaderID > 0 {
		leaderID = strconv.Itoa(node.LeaderID)
	}
	fmt.Printf("  Leader ID:       %s\n", leaderID)

	commitIndex := "-"
	if node.CommitIndex >= 0 {
		commitIndex = strconv.Itoa(node.CommitIndex)
	}
	fmt.Printf("  Commit Index:    %s\n", commitIndex)
	fmt.Printf("  Log Entries:     %d\n", len(node.LogEntries))

	if len(node.LogEntries) > 0 {
		fmt.Println()
		fmt.Println("  Log:")
		for i, entry := range node.LogEntries {
			status := ""
			if i <= node.CommitIndex {
				status = "[C]"
			}
			fmt.Printf("    [%d] Term=%d %s Command=%q\n",
				i, entry.Term, status, entry.Command)
		}
	}
}

func submitLog(args []string) {
	fs := flag.NewFlagSet("submit", flag.ExitOnError)
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Println("Error: command is required")
		fmt.Println("Usage: raft-client submit \"<command>\"")
		os.Exit(1)
	}

	command := strings.Join(fs.Args(), " ")

	req := api.SubmitLogRequest{Command: command}
	var resp api.SubmitLogResponse
	if err := postJSON("/api/log/submit", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("Success: %s\n", resp.Message)
		fmt.Printf("  Term:  %d\n", resp.Term)
		fmt.Printf("  Index: %d\n", resp.Index)
	} else {
		fmt.Printf("Error: %s\n", resp.Message)
		os.Exit(1)
	}
}

func triggerElection() {
	var resp api.TriggerElectionResponse
	if err := postJSON("/api/election/trigger", nil, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("Success: %s\n", resp.Message)
	} else {
		fmt.Printf("Error: %s\n", resp.Message)
		os.Exit(1)
	}
}

func postJSON(path string, req interface{}, resp interface{}) error {
	var body io.Reader
	if req != nil {
		jsonData, err := json.Marshal(req)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	httpResp, err := http.Post(baseURL+path, "application/json", body)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	return json.NewDecoder(httpResp.Body).Decode(resp)
}

func getJSON(path string, resp interface{}) error {
	httpResp, err := http.Get(baseURL + path)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	return json.NewDecoder(httpResp.Body).Decode(resp)
}
