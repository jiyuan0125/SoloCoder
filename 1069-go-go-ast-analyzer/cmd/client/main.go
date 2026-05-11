package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/example/go-ast-analyzer/internal/api"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8303", "Analysis server URL")
	flag.Parse()
	
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "Usage: analyzer-client <go-file>")
		os.Exit(1)
	}
	
	filename := flag.Arg(0)
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}
	
	req := api.AnalysisRequest{
		Filename: filepath.Base(filename),
		Source:   string(source),
	}
	
	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling request: %v\n", err)
		os.Exit(1)
	}
	
	resp, err := http.Post(*serverURL+"/analyze", "application/json", bytes.NewReader(reqBody))
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
	
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Server error: %s\n", string(body))
		os.Exit(1)
	}
	
	var analysis api.AnalysisResponse
	if err := json.Unmarshal(body, &analysis); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}
	
	printResponse(&analysis)
}

func printResponse(resp *api.AnalysisResponse) {
	if len(resp.SyntaxErrors) > 0 {
		fmt.Println("Syntax Errors:")
		for _, e := range resp.SyntaxErrors {
			fmt.Printf("  %s:%d: %s\n", e.Filename, e.Line, e.Message)
		}
		return
	}
	
	fmt.Println("Function Complexity:")
	for _, c := range resp.Complexities {
		fmt.Printf("  %s (line %d): %d\n", c.Name, c.Line, c.Complexity)
	}
	
	fmt.Println("\nCall Graph:")
	for caller, callees := range resp.CallGraph.Adjacency {
		fmt.Printf("  %s:\n", caller)
		for _, callee := range callees {
			fmt.Printf("    -> %s\n", callee)
		}
	}
	
	fmt.Println("\nUnused Variables:")
	if len(resp.UnusedVariables) == 0 {
		fmt.Println("  None")
	} else {
		for _, v := range resp.UnusedVariables {
			fmt.Printf("  %s:%d: %s\n", v.Filename, v.Line, v.Name)
		}
	}
	
	fmt.Println("\nUnused Imports:")
	if len(resp.UnusedImports) == 0 {
		fmt.Println("  None")
	} else {
		for _, imp := range resp.UnusedImports {
			fmt.Printf("  %s:%d: %s\n", imp.Filename, imp.Line, imp.Path)
		}
	}
}
