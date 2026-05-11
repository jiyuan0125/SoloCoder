package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"network-planner/common"
)

func main() {
	var filePath string
	var serverURL string
	var port string

	flag.StringVar(&filePath, "file", "", "Path to JSON file with topology data")
	flag.StringVar(&serverURL, "server", "", "Server URL (e.g., http://localhost:8080)")
	flag.StringVar(&port, "port", "", "Server port (used if server URL not specified)")
	flag.Parse()

	if filePath == "" {
		flag.Usage()
		os.Exit(1)
	}

	if serverURL == "" {
		if port == "" {
			port = os.Getenv("PORT")
		}
		if port == "" {
			port = "8080"
		}
		serverURL = "http://localhost:" + port
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var topology common.TopologyRequest
	if err := json.Unmarshal(data, &topology); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	fmt.Printf("Loaded topology with %d nodes and %d edges\n", len(topology.Nodes), len(topology.Edges))

	reqBody, err := json.Marshal(topology)
	if err != nil {
		log.Fatalf("Failed to marshal request: %v", err)
	}

	resp, err := http.Post(serverURL+"/api/mst", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	var result common.MSTResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalf("Failed to parse response: %v", err)
	}

	fmt.Println()
	if result.Success {
		fmt.Println("=== Minimum Spanning Tree ===")
		fmt.Printf("Total Cost: %d 元\n", result.TotalCost)
		fmt.Println("\nSelected Edges:")
		for i, edge := range result.Edges {
			fmt.Printf("  %d. %s -- %s (成本: %d 元)\n", i+1, edge.From, edge.To, edge.Cost)
		}
	} else {
		fmt.Println("Error:", result.Message)
		if len(result.Isolated) > 0 {
			fmt.Println("\nIsolated Devices:")
			for _, device := range result.Isolated {
				fmt.Println("  -", device)
			}
		}
	}
}
