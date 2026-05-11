package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gomod-mvs/pkg/api"
)

var serverURL = "http://localhost:8440"

func main() {
	if envURL := os.Getenv("SERVER_URL"); envURL != "" {
		serverURL = envURL
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "graph":
		handleGraph()
	case "verify":
		handleVerify()
	case "why":
		handleWhy()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Go Module MVS Analyzer")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  gomod-mvs graph     Output dependency graph")
	fmt.Println("  gomod-mvs verify    Verify go.sum integrity")
	fmt.Println("  gomod-mvs why <mod> Query why a module is included")
	fmt.Println("")
	fmt.Println("Flags:")
	flag.PrintDefaults()
}

func readGoModAndSum() (string, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}

	goModPath := filepath.Join(cwd, "go.mod")
	goSumPath := filepath.Join(cwd, "go.sum")

	goModContent, err := os.ReadFile(goModPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read go.mod: %v", err)
	}

	goSumContent, err := os.ReadFile(goSumPath)
	if err != nil {
		goSumContent = []byte{}
	}

	return string(goModContent), string(goSumContent), nil
}

func handleGraph() {
	goModContent, goSumContent, err := readGoModAndSum()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	req := api.GraphRequest{
		GoModContent: goModContent,
		GoSumContent: goSumContent,
	}

	var resp api.GraphResponse
	if err := sendRequest(serverURL+"/graph", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if resp.Error != "" {
		fmt.Printf("Server error: %s\n", resp.Error)
		os.Exit(1)
	}

	printGraph(&resp.Graph)
}

func handleVerify() {
	goModContent, goSumContent, err := readGoModAndSum()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	req := api.VerifyRequest{
		GoModContent: goModContent,
		GoSumContent: goSumContent,
	}

	var resp api.VerifyResponse
	if err := sendRequest(serverURL+"/verify", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if resp.Error != "" {
		fmt.Printf("Server error: %s\n", resp.Error)
		os.Exit(1)
	}

	printVerifyResult(&resp.Result)
}

func handleWhy() {
	if len(os.Args) < 3 {
		fmt.Println("Error: 'why' command requires a module path argument")
		fmt.Println("Usage: gomod-mvs why <module-path>")
		os.Exit(1)
	}

	targetPath := os.Args[2]

	goModContent, goSumContent, err := readGoModAndSum()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	req := api.WhyRequest{
		GoModContent: goModContent,
		GoSumContent: goSumContent,
		TargetPath:   targetPath,
	}

	var resp api.WhyResponse
	if err := sendRequest(serverURL+"/why", req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if resp.Error != "" {
		fmt.Printf("Server error: %s\n", resp.Error)
		os.Exit(1)
	}

	printWhyResult(&resp)
}

func sendRequest(url string, req, resp interface{}) error {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	httpResp, err := httpClient("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 {
		body, _ := io.ReadAll(httpResp.Body)
		return fmt.Errorf("server returned status %d: %s", httpResp.StatusCode, string(body))
	}

	if err := json.NewDecoder(httpResp.Body).Decode(resp); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	return nil
}

func httpClient(method, url string, body io.Reader) (*httpResponse, error) {
	client := &bytes.Buffer{}
	encoder := json.NewEncoder(client)

	req := map[string]interface{}{
		"method": method,
		"url":    url,
		"body":   body,
	}

	_ = encoder.Encode(req)

	return makeHTTPRequest(method, url, body)
}

type httpResponse struct {
	StatusCode int
	Body       io.ReadCloser
}

func makeHTTPRequest(method, url string, body io.Reader) (*httpResponse, error) {
	httpReq, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	return &httpResponse{
		StatusCode: resp.StatusCode,
		Body:       resp.Body,
	}, nil
}

func printGraph(graph *api.GraphInfo) {
	fmt.Printf("Dependency Graph for: %s\n\n", graph.Root)
	
	nodeMap := make(map[string]api.NodeInfo)
	for _, n := range graph.Nodes {
		nodeMap[n.ID] = n
	}
	
	rootNode, exists := nodeMap["root"]
	if exists {
		fmt.Printf("Root: %s\n", rootNode.Path)
	}
	
	fmt.Println("\nNodes:")
	for _, node := range graph.Nodes {
		if node.ID == "root" {
			continue
		}
		
		directMark := ""
		if node.IsDirect {
			directMark = " [direct]"
		}
		
		version := ""
		if node.Version != "" {
			version = " " + node.Version
		}
		
		fmt.Printf("  %s%s%s\n", node.Path, version, directMark)
	}
	
	fmt.Println("\nEdges:")
	edgeMap := make(map[string][]string)
	for _, edge := range graph.Edges {
		edgeMap[edge.From] = append(edgeMap[edge.From], edge.To)
	}
	
	for from, toList := range edgeMap {
		fromNode := nodeMap[from]
		fromLabel := fromNode.Path
		if fromNode.Version != "" && from != "root" {
			fromLabel += "@" + fromNode.Version
		}
		
		for _, to := range toList {
			toNode := nodeMap[to]
			toLabel := toNode.Path
			if toNode.Version != "" {
				toLabel += "@" + toNode.Version
			}
			
			fmt.Printf("  %s -> %s\n", fromLabel, toLabel)
		}
	}
}

func printVerifyResult(result *api.VerifyResultInfo) {
	if result.AllChecked {
		fmt.Println("✅ go.sum verification passed")
		return
	}
	
	fmt.Println("❌ go.sum verification failed")
	fmt.Println("")
	
	if len(result.Orphans) > 0 {
		fmt.Printf("Found %d orphan entries in go.sum (not required by any dependency):\n", len(result.Orphans))
		for _, o := range result.Orphans {
			fmt.Printf("  - %s@%s\n", o.Path, o.Version)
		}
		fmt.Println("")
	}
	
	if len(result.Mismatches) > 0 {
		fmt.Printf("Found %d mismatches:\n", len(result.Mismatches))
		for _, m := range result.Mismatches {
			fmt.Printf("  - %s@%s: %s\n", m.Path, m.Version, m.Error)
		}
	}
}

func printWhyResult(result *api.WhyResponse) {
	fmt.Printf("Query: %s\n\n", result.Target)
	
	if !result.Found {
		fmt.Printf("❌ Module not found in dependency chain\n")
		return
	}
	
	fmt.Printf("✅ Found %d dependency chain(s):\n\n", len(result.Chains))
	
	for i, chain := range result.Chains {
		fmt.Printf("Chain %d:\n", i+1)
		fmt.Printf("  %s\n", strings.Join(chain, " -> "))
	}
}
