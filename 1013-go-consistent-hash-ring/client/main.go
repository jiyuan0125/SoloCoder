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

	"github.com/solocoder/consistenthash/api"
)

const defaultServer = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	server := defaultServer
	if env := os.Getenv("HASH_SERVER"); env != "" {
		server = env
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "add":
		handleAdd(server, args)
	case "remove":
		handleRemove(server, args)
	case "lookup":
		handleLookup(server, args)
	case "stats":
		handleStats(server, args)
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Consistent Hash Client

Usage:
  client [command] [options]

Commands:
  add     <node-id> [weight]    Add a node to the hash ring
  remove  <node-id>             Remove a node from the hash ring
  lookup  <key>                 Lookup which node is responsible for a key
  stats                         Display hash ring statistics
  help                          Show this help message

Environment:
  HASH_SERVER    Server address (default: http://localhost:8080)

Examples:
  client add node1
  client add node2 3
  client lookup mykey
  client remove node1
  client stats`)
}

func handleAdd(server string, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: client add <node-id> [weight]")
		os.Exit(1)
	}

	id := args[0]
	weight := 1

	if len(args) > 1 {
		w, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Printf("Invalid weight: %s\n", args[1])
			os.Exit(1)
		}
		weight = w
	}

	req := api.AddNodeRequest{
		ID:     id,
		Weight: weight,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Failed to encode request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(server+"/hash/nodes", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		_ = json.Unmarshal(respBody, &errResp)
		if errResp.Error != "" {
			fmt.Printf("Error: %s\n", errResp.Error)
		} else {
			fmt.Printf("Request failed with status: %d\n", resp.StatusCode)
		}
		os.Exit(1)
	}

	var addResp api.AddNodeResponse
	_ = json.Unmarshal(respBody, &addResp)
	fmt.Printf("Node '%s' added (weight=%d)\n", id, weight)
}

func handleRemove(server string, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: client remove <node-id>")
		os.Exit(1)
	}

	id := args[0]

	req, err := http.NewRequest(http.MethodDelete, server+"/hash/nodes/"+id, nil)
	if err != nil {
		fmt.Printf("Failed to create request: %v\n", err)
		os.Exit(1)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		_ = json.Unmarshal(respBody, &errResp)
		if errResp.Error != "" {
			fmt.Printf("Error: %s\n", errResp.Error)
		} else {
			fmt.Printf("Request failed with status: %d\n", resp.StatusCode)
		}
		os.Exit(1)
	}

	fmt.Printf("Node '%s' removed\n", id)
}

func handleLookup(server string, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: client lookup <key>")
		os.Exit(1)
	}

	key := args[0]

	resp, err := http.Get(server + "/hash/lookup?key=" + strings.ReplaceAll(key, " ", "+"))
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var lookupResp api.LookupResponse
		_ = json.Unmarshal(respBody, &lookupResp)
		if lookupResp.Error != "" {
			fmt.Printf("Error: %s\n", lookupResp.Error)
		} else {
			fmt.Printf("Request failed with status: %d\n", resp.StatusCode)
		}
		os.Exit(1)
	}

	var lookupResp api.LookupResponse
	_ = json.Unmarshal(respBody, &lookupResp)
	fmt.Printf("Key '%s' -> Node: %s\n", key, lookupResp.Node)
}

func handleStats(server string, args []string) {
	flagSet := flag.NewFlagSet("stats", flag.ExitOnError)
	flagSet.Parse(args)

	resp, err := http.Get(server + "/hash/stats")
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Request failed with status: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	var stats api.StatsResponse
	_ = json.Unmarshal(respBody, &stats)

	fmt.Println("=== Hash Ring Statistics ===")
	fmt.Printf("Total Virtual Nodes: %d\n", stats.TotalVirtualNodes)
	fmt.Printf("Nodes: %d\n", len(stats.Nodes))
	fmt.Println()

	if len(stats.Nodes) == 0 {
		fmt.Println("No nodes in the ring.")
		return
	}

	nodeNames := make([]string, 0, len(stats.Nodes))
	for name := range stats.Nodes {
		nodeNames = append(nodeNames, name)
	}
	sort.Strings(nodeNames)

	for _, name := range nodeNames {
		info := stats.Nodes[name]
		fmt.Printf("Node: %s\n", name)
		fmt.Printf("  Weight:           %d\n", info.Weight)
		fmt.Printf("  Virtual Nodes:    %d\n", info.VirtualNodeCount)
		fmt.Printf("  Distribution:     %.2f%%\n", info.Percentage)
		fmt.Println()
	}
}
