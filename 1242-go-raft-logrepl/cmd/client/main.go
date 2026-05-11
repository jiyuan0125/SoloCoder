package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"raftlog/common"
	"strings"
)

func getServerURL() string {
	url := os.Getenv("SERVER_URL")
	if url == "" {
		return "http://localhost:8080"
	}
	return url
}

func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func httpPostJSON(url string, body interface{}) ([]byte, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func usage() {
	fmt.Println("Raft Log Replication Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client submit <node_id> <command>")
	fmt.Println("  client log <node_id>")
	fmt.Println("  client check <node_id1> <node_id2>")
	fmt.Println("  client reset <node_id>")
	fmt.Println("  client status")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  SERVER_URL  Service endpoint (default: http://localhost:8080)")
	os.Exit(2)
}

func cmdSubmit(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: client submit <node_id> <command>")
		os.Exit(2)
	}
	nodeID := args[0]
	command := strings.Join(args[1:], " ")
	req := common.SubmitRequest{
		NodeID:  nodeID,
		Command: command,
	}
	body, err := httpPostJSON(getServerURL()+"/submit", req)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	var resp common.SubmitResponse
	_ = json.Unmarshal(body, &resp)
	if resp.Success {
		fmt.Printf("OK: %s\n", resp.Message)
	} else {
		fmt.Printf("FAILED: %s\n", resp.Message)
		os.Exit(1)
	}
}

func cmdLog(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: client log <node_id>")
		os.Exit(2)
	}
	nodeID := args[0]
	body, err := httpGet(fmt.Sprintf("%s/log?node_id=%s", getServerURL(), nodeID))
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	var resp common.LogResponse
	_ = json.Unmarshal(body, &resp)
	if !resp.Success {
		fmt.Printf("FAILED: %s\n", resp.Message)
		os.Exit(1)
	}
	fmt.Printf("Node %s (%d entries):\n", resp.NodeID, len(resp.Entries))
	for i, e := range resp.Entries {
		fmt.Printf("  [%d] index=%d term=%d command=%q\n", i+1, e.Index, e.Term, e.Command)
	}
}

func cmdCheck(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: client check <node_id1> <node_id2>")
		os.Exit(2)
	}
	n1, n2 := args[0], args[1]
	body, err := httpGet(fmt.Sprintf("%s/check?node_id1=%s&node_id2=%s", getServerURL(), n1, n2))
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	var resp common.ConsistencyCheckResponse
	_ = json.Unmarshal(body, &resp)
	if !resp.Success {
		fmt.Printf("FAILED: %s\n", resp.Message)
		os.Exit(1)
	}
	if resp.Consistent {
		fmt.Printf("CONSISTENT: matched %d entries\n", resp.MatchedCount)
		fmt.Printf("  %s: %d entries\n", resp.Node1ID, resp.Node1Total)
		fmt.Printf("  %s: %d entries\n", resp.Node2ID, resp.Node2Total)
		return
	}
	fmt.Println("INCONSISTENT")
	fmt.Printf("  %s: %d entries\n", resp.Node1ID, resp.Node1Total)
	fmt.Printf("  %s: %d entries\n", resp.Node2ID, resp.Node2Total)
	fmt.Printf("  First mismatch at index %d\n", resp.MismatchIndex)
	if resp.Node1Entry != nil {
		e := resp.Node1Entry
		fmt.Printf("    %s: index=%d term=%d command=%q\n", resp.Node1ID, e.Index, e.Term, e.Command)
	} else {
		fmt.Printf("    %s: (no entry)\n", resp.Node1ID)
	}
	if resp.Node2Entry != nil {
		e := resp.Node2Entry
		fmt.Printf("    %s: index=%d term=%d command=%q\n", resp.Node2ID, e.Index, e.Term, e.Command)
	} else {
		fmt.Printf("    %s: (no entry)\n", resp.Node2ID)
	}
	os.Exit(1)
}

func cmdReset(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: client reset <node_id>")
		os.Exit(2)
	}
	nodeID := args[0]
	req := common.ResetRequest{NodeID: nodeID}
	body, err := httpPostJSON(getServerURL()+"/reset", req)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	var resp common.ResetResponse
	_ = json.Unmarshal(body, &resp)
	if resp.Success {
		fmt.Printf("OK: %s\n", resp.Message)
	} else {
		fmt.Printf("FAILED: %s\n", resp.Message)
		os.Exit(1)
	}
}

func cmdStatus(args []string) {
	body, err := httpGet(getServerURL() + "/status")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	var resp common.StatusResponse
	_ = json.Unmarshal(body, &resp)
	if !resp.Success {
		fmt.Printf("FAILED: %s\n", resp.Message)
		os.Exit(1)
	}
	fmt.Printf("Cluster Status (%d nodes):\n", len(resp.Nodes))
	for _, n := range resp.Nodes {
		fmt.Printf("  Node %s (%s)\n", n.ID, n.Address)
		fmt.Printf("    Entries: %d, LastIndex: %d, LastTerm: %d\n", n.EntryCount, n.LastIndex, n.LastTerm)
		for i, e := range n.Entries {
			fmt.Printf("    [%d] index=%d term=%d command=%q\n", i+1, e.Index, e.Term, e.Command)
		}
	}
}

func main() {
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		usage()
	}
	cmd := strings.ToLower(args[0])
	switch cmd {
	case "submit":
		cmdSubmit(args[1:])
	case "log":
		cmdLog(args[1:])
	case "check":
		cmdCheck(args[1:])
	case "reset":
		cmdReset(args[1:])
	case "status":
		cmdStatus(args[1:])
	default:
		usage()
	}
}
