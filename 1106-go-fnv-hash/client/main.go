package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/example/fnv/pkg/api"
)

const (
	DefaultServer = "http://localhost:8080"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := os.Getenv("FNV_SERVER")
	if serverURL == "" {
		serverURL = DefaultServer
	}

	cmd := os.Args[1]
	switch cmd {
	case "hash":
		if err := runHash(serverURL, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "consistent":
		if err := runConsistent(serverURL, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  client hash [--bits 32|64] [--algorithm fnv1a] < input_data")
	fmt.Println("  client consistent --nodes node1,node2,node3 --key data_key")
	fmt.Println()
	fmt.Println("Hash subcommand reads data from stdin.")
	fmt.Println("Environment variable FNV_SERVER can override server URL.")
}

func runHash(serverURL string, args []string) error {
	bits := 64
	algorithm := "fnv1a"

	i := 0
	for i < len(args) {
		arg := args[i]
		switch arg {
		case "--bits":
			if i+1 >= len(args) {
				return fmt.Errorf("--bits requires a value")
			}
			fmt.Sscanf(args[i+1], "%d", &bits)
			i += 2
		case "--algorithm":
			if i+1 >= len(args) {
				return fmt.Errorf("--algorithm requires a value")
			}
			algorithm = args[i+1]
			i += 2
		default:
			return fmt.Errorf("unknown argument: %s", arg)
		}
	}

	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %v", err)
	}

	req := api.HashRequest{
		Algorithm: algorithm,
		Bits:      bits,
		Input:     string(input),
	}

	var resp api.HashResponse
	if err := postJSON(serverURL+"/hash", req, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("server error: %s", resp.Error)
	}

	fmt.Println(resp.Hash)
	return nil
}

func runConsistent(serverURL string, args []string) error {
	var nodes []string
	var key string

	i := 0
	for i < len(args) {
		arg := args[i]
		switch arg {
		case "--nodes":
			if i+1 >= len(args) {
				return fmt.Errorf("--nodes requires a value")
			}
			nodes = strings.Split(args[i+1], ",")
			i += 2
		case "--key":
			if i+1 >= len(args) {
				return fmt.Errorf("--key requires a value")
			}
			key = args[i+1]
			i += 2
		default:
			return fmt.Errorf("unknown argument: %s", arg)
		}
	}

	if key == "" {
		return fmt.Errorf("--key is required")
	}

	req := api.ConsistentRequest{
		Nodes: nodes,
		Key:   key,
	}

	var resp api.ConsistentResponse
	if err := postJSON(serverURL+"/consistent", req, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("server error: %s", resp.Error)
	}

	fmt.Println(resp.Node)
	return nil
}

func postJSON(url string, req, resp interface{}) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	httpResp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if err := json.Unmarshal(body, resp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %v\nbody: %s", err, string(body))
	}

	return nil
}
