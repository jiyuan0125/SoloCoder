package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"yaml-anchor-resolver/pkg/anchor"
	"yaml-anchor-resolver/pkg/api"
)

func main() {
	var (
		filePath   string
		yamlText   string
		serverURL  string
		outputFormat string
		listAnchors bool
		useLocal   bool
	)

	flag.StringVar(&filePath, "file", "", "Path to YAML file")
	flag.StringVar(&yamlText, "text", "", "YAML text to parse")
	flag.StringVar(&serverURL, "server", "http://localhost:8080", "Server URL")
	flag.StringVar(&outputFormat, "format", "json", "Output format: json or yaml")
	flag.BoolVar(&listAnchors, "list", false, "List all anchors and references")
	flag.BoolVar(&useLocal, "local", true, "Use local resolver (false to use server)")
	flag.Parse()

	if filePath == "" && yamlText == "" {
		fmt.Println("Error: Please provide either -file or -text")
		flag.Usage()
		os.Exit(1)
	}

	var content string
	var err error
	if filePath != "" {
		content, err = readFile(filePath)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}
	} else {
		content = yamlText
	}

	if listAnchors {
		if useLocal {
			listAnchorsLocal(content)
		} else {
			listAnchorsRemote(serverURL, content)
		}
	} else {
		if useLocal {
			resolveLocal(content, outputFormat)
		} else {
			resolveRemote(serverURL, content, outputFormat)
		}
	}
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func listAnchorsLocal(content string) {
	resolver := anchor.NewResolver()
	result, err := resolver.Resolve(content)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Anchors:")
	for _, a := range result.Anchors {
		fmt.Printf("  - %s at %s\n", a.Name, a.Location)
	}

	fmt.Println("\nReferences:")
	for _, r := range result.References {
		mergeStr := ""
		if r.IsMergeKey {
			mergeStr = " (merge key)"
		}
		fmt.Printf("  - %s at %s%s\n", r.Name, r.Location, mergeStr)
	}

	fmt.Println("\nRelations:")
	for _, rel := range result.Relations {
		mergeStr := ""
		if rel.IsMerge {
			mergeStr = " (merge)"
		}
		fmt.Printf("  - %s -> %s%s\n", rel.Anchor, rel.Reference, mergeStr)
	}

	if len(result.Warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, w := range result.Warnings {
			fmt.Printf("  - %s\n", w)
		}
	}
}

func listAnchorsRemote(serverURL, content string) {
	url := fmt.Sprintf("%s/anchors", serverURL)
	reqBody := api.ListAnchorsRequest{YAMLText: content}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("Error encoding request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error calling server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	var result api.ListAnchorsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		fmt.Printf("Raw response: %s\n", string(body))
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Error)
		os.Exit(1)
	}

	fmt.Println("Anchors:")
	for _, a := range result.Anchors {
		fmt.Printf("  - %s at %s\n", a.Name, a.Location)
	}

	fmt.Println("\nReferences:")
	for _, r := range result.Refs {
		mergeStr := ""
		if r.IsMergeKey {
			mergeStr = " (merge key)"
		}
		fmt.Printf("  - %s at %s%s\n", r.Name, r.Location, mergeStr)
	}

	fmt.Println("\nRelations:")
	for _, rel := range result.Relations {
		mergeStr := ""
		if rel.IsMerge {
			mergeStr = " (merge)"
		}
		fmt.Printf("  - %s -> %s%s\n", rel.Anchor, rel.Reference, mergeStr)
	}
}

func resolveLocal(content, format string) {
	resolver := anchor.NewResolver()
	result, err := resolver.Resolve(content)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(result.Warnings) > 0 {
		fmt.Fprintln(os.Stderr, "Warnings:")
		for _, w := range result.Warnings {
			fmt.Fprintf(os.Stderr, "  - %s\n", w)
		}
	}

	var output string
	if format == "yaml" {
		output, err = resolver.ToYAML(result.ResolvedData)
	} else {
		output, err = resolver.ToJSON(result.ResolvedData)
	}

	if err != nil {
		fmt.Printf("Error formatting output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(output)
}

func resolveRemote(serverURL, content, format string) {
	url := fmt.Sprintf("%s/resolve", serverURL)
	reqBody := api.ResolveRequest{YAMLText: content}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("Error encoding request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error calling server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	var result api.ResolveResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		fmt.Printf("Raw response: %s\n", string(body))
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Error)
		os.Exit(1)
	}

	if len(result.Warnings) > 0 {
		fmt.Fprintln(os.Stderr, "Warnings:")
		for _, w := range result.Warnings {
			fmt.Fprintf(os.Stderr, "  - %s\n", w)
		}
	}

	var output string
	resolver := anchor.NewResolver()
	if format == "yaml" {
		output, err = resolver.ToYAML(result.Data)
	} else {
		output, err = resolver.ToJSON(result.Data)
	}

	if err != nil {
		fmt.Printf("Error formatting output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(output)
}
