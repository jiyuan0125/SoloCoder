package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"euler-path/common"
)

func main() {
	var serverURL string
	var nodesStr string
	var edgesStr string
	var start string
	var inputFile string

	flag.StringVar(&serverURL, "server", "http://localhost:8080", "server URL")
	flag.StringVar(&nodesStr, "nodes", "", "comma-separated list of nodes")
	flag.StringVar(&edgesStr, "edges", "", "comma-separated list of edges, each edge is from-to format")
	flag.StringVar(&start, "start", "", "start node (optional)")
	flag.StringVar(&inputFile, "file", "", "input JSON file")
	flag.Parse()

	var req common.EulerPathRequest

	if inputFile != "" {
		data, err := ioutil.ReadFile(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
			os.Exit(1)
		}
		if err := json.Unmarshal(data, &req); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing input file: %v\n", err)
			os.Exit(1)
		}
	} else {
		if nodesStr != "" {
			req.Nodes = strings.Split(nodesStr, ",")
		}
		if edgesStr != "" {
			edgeParts := strings.Split(edgesStr, ",")
			for _, part := range edgeParts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				fromTo := strings.Split(part, "-")
				if len(fromTo) != 2 {
					fmt.Fprintf(os.Stderr, "Invalid edge format: %s (expected from-to)\n", part)
					os.Exit(1)
				}
				req.Edges = append(req.Edges, common.Edge{
					From: strings.TrimSpace(fromTo[0]),
					To:   strings.TrimSpace(fromTo[1]),
				})
			}
		}
		if start != "" {
			req.Start = start
		}
	}

	if len(req.Edges) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no edges provided. Use -edges or -file option required.")
		fmt.Fprintln(os.Stderr, "Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding request: %v\n", err)
		os.Exit(1)
	}

	url := serverURL + "/euler-path"
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	var result common.EulerPathResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		fmt.Fprintf(os.Stderr, "Response body: %s\n", string(body))
		os.Exit(1)
	}

	printResult(&result)
}

func printResult(resp *common.EulerPathResponse) {
	if resp.Error != "" {
		fmt.Printf("Error: %s\n", resp.Error)
	}

	fmt.Printf("Has Euler Path: %v\n", resp.HasEulerPath)
	fmt.Printf("Has Euler Circuit: %v\n", resp.HasEulerCircuit)

	if len(resp.Path) > 0 {
		fmt.Printf("\nPath (%d nodes):\n", len(resp.Path))
		for i, node := range resp.Path {
			if i > 0 {
				fmt.Print(" -> ")
			}
			fmt.Print(node)
		}
		fmt.Println()
	}
}
