package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/spf-router/common"
)

type Config struct {
	ServerURL string
}

type TopologyFile struct {
	Nodes []common.Node `json:"nodes"`
	Links []common.Link `json:"links"`
}

var config Config

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	topologyFile := flag.String("topology", "", "Initial topology file")
	flag.Parse()

	config = Config{ServerURL: *serverURL}

	if *topologyFile != "" {
		if err := loadTopologyFile(*topologyFile); err != nil {
			fmt.Printf("Error loading topology: %v\n", err)
			os.Exit(1)
		}
	}

	runInteractive()
}

func loadTopologyFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	var tf TopologyFile
	if err := json.Unmarshal(data, &tf); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	if len(tf.Nodes) == 0 && len(tf.Links) == 0 {
		fmt.Println("Topology file is empty.")
		return nil
	}

	req := common.TopologyUpdateRequest{
		AddNodes: tf.Nodes,
		AddLinks: tf.Links,
	}

	resp, err := updateTopology(req)
	if err != nil {
		return fmt.Errorf("failed to update topology: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("server error: %s", resp.Message)
	}

	fmt.Printf("Loaded %d nodes and %d links from file.\n", len(tf.Nodes), len(tf.Links))
	return nil
}

func runInteractive() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("SPF Router Client")
	fmt.Println("Available commands:")
	fmt.Println("  add-node <node-id>")
	fmt.Println("  remove-node <node-id>")
	fmt.Println("  add-link <from> <to> <cost> [seq-num]")
	fmt.Println("  remove-link <from> <to>")
	fmt.Println("  run-spf <source-node>")
	fmt.Println("  query-paths <source-node>")
	fmt.Println("  help")
	fmt.Println("  exit")

	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		cmd := parts[0]
		args := parts[1:]

		switch cmd {
		case "exit", "quit":
			fmt.Println("Bye.")
			return
		case "help":
			printHelp()
		case "add-node":
			handleAddNode(args)
		case "remove-node":
			handleRemoveNode(args)
		case "add-link":
			handleAddLink(args)
		case "remove-link":
			handleRemoveLink(args)
		case "run-spf":
			handleRunSPF(args)
		case "query-paths":
			handleQueryPaths(args)
		default:
			fmt.Printf("Unknown command: %s. Type 'help' for list of commands.\n", cmd)
		}
	}
}

func printHelp() {
	fmt.Println("\nCommands:")
	fmt.Println("  add-node <node-id>          - Add a node to the topology")
	fmt.Println("  remove-node <node-id>       - Remove a node from the topology")
	fmt.Println("  add-link <from> <to> <cost> [seq-num]")
	fmt.Println("                              - Add a link with cost (and optional sequence number)")
	fmt.Println("  remove-link <from> <to>     - Remove a link between two nodes")
	fmt.Println("  run-spf <source-node>       - Run SPF algorithm from source node")
	fmt.Println("  query-paths <source-node>   - Query shortest paths from source node")
	fmt.Println("  help                        - Show this help message")
	fmt.Println("  exit                        - Exit the client")
}

func handleAddNode(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: add-node <node-id>")
		return
	}

	nodeID := args[0]
	req := common.TopologyUpdateRequest{
		AddNodes: []common.Node{{ID: nodeID}},
	}

	resp, err := updateTopology(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Printf("Node '%s' added successfully.\n", nodeID)
	} else {
		fmt.Printf("Failed to add node: %s\n", resp.Message)
	}
}

func handleRemoveNode(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: remove-node <node-id>")
		return
	}

	nodeID := args[0]
	req := common.TopologyUpdateRequest{
		RemoveNodes: []string{nodeID},
	}

	resp, err := updateTopology(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Printf("Node '%s' removed successfully.\n", nodeID)
	} else {
		fmt.Printf("Failed to remove node: %s\n", resp.Message)
	}
}

func handleAddLink(args []string) {
	if len(args) < 3 || len(args) > 4 {
		fmt.Println("Usage: add-link <from> <to> <cost> [seq-num]")
		return
	}

	from := args[0]
	to := args[1]
	cost, err := strconv.ParseUint(args[2], 10, 32)
	if err != nil {
		fmt.Printf("Invalid cost: %s\n", args[2])
		return
	}

	var seqNum uint64 = 0
	if len(args) == 4 {
		seqNum, err = strconv.ParseUint(args[3], 10, 32)
		if err != nil {
			fmt.Printf("Invalid sequence number: %s\n", args[3])
			return
		}
	}

	req := common.TopologyUpdateRequest{
		AddLinks: []common.Link{
			{
				From:   from,
				To:     to,
				Cost:   uint32(cost),
				SeqNum: uint32(seqNum),
			},
		},
	}

	resp, err := updateTopology(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Printf("Link '%s' -> '%s' (cost: %d) added successfully.\n", from, to, cost)
	} else {
		fmt.Printf("Failed to add link: %s\n", resp.Message)
	}
}

func handleRemoveLink(args []string) {
	if len(args) != 2 {
		fmt.Println("Usage: remove-link <from> <to>")
		return
	}

	from := args[0]
	to := args[1]

	req := common.TopologyUpdateRequest{
		RemoveLinks: []common.Link{
			{From: from, To: to},
		},
	}

	resp, err := updateTopology(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Printf("Link '%s' -> '%s' removed successfully.\n", from, to)
	} else {
		fmt.Printf("Failed to remove link: %s\n", resp.Message)
	}
}

func handleRunSPF(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: run-spf <source-node>")
		return
	}

	source := args[0]
	req := common.SPFRunRequest{Source: source}

	resp, err := runSPF(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Printf("SPF computation completed successfully from source '%s'.\n", source)
	} else {
		fmt.Printf("SPF computation failed: %s\n", resp.Message)
	}
}

func handleQueryPaths(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: query-paths <source-node>")
		return
	}

	source := args[0]
	resp, err := queryPaths(source)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if !resp.Success {
		fmt.Printf("Query failed: %s\n", resp.Message)
		return
	}

	if len(resp.Paths) == 0 {
		fmt.Printf("No paths found from source '%s'. Topology may be empty.\n", source)
		return
	}

	fmt.Printf("\nShortest paths from '%s':\n", source)
	for _, entry := range resp.Paths {
		if entry.Destination == source {
			continue
		}

		status := "VALID"
		if !entry.Valid {
			status = "INVALID"
		}

		if len(entry.Paths) == 0 {
			fmt.Printf("  To %s: UNREACHABLE\n", entry.Destination)
		} else {
			fmt.Printf("  To %s (cost: %d, %s):\n", entry.Destination, entry.TotalCost, status)
			for i, path := range entry.Paths {
				fmt.Printf("    Path %d: %s\n", i+1, strings.Join(path, " -> "))
			}
		}
	}
	fmt.Println()
}

func updateTopology(req common.TopologyUpdateRequest) (*common.TopologyUpdateResponse, error) {
	url := fmt.Sprintf("%s/topology/update", config.ServerURL)
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpResp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	respBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	var resp common.TopologyUpdateResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func runSPF(req common.SPFRunRequest) (*common.SPFRunResponse, error) {
	url := fmt.Sprintf("%s/spf/run", config.ServerURL)
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpResp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	respBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	var resp common.SPFRunResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func queryPaths(source string) (*common.ShortestPathQueryResponse, error) {
	url := fmt.Sprintf("%s/spf/paths?source=%s", config.ServerURL, source)

	httpResp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	respBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	var resp common.ShortestPathQueryResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
