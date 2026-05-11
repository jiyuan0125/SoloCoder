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

	"hungarian-match/pkg/api"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	mode := flag.String("mode", "match", "Mode: match or query")
	interns := flag.String("interns", "", "Comma-separated intern IDs")
	positions := flag.String("positions", "", "Comma-separated position IDs")
	edges := flag.String("edges", "", "Edges in format I1:P1,I2:P2,I3:P1")
	queryIntern := flag.String("intern", "", "Intern ID to query")
	flag.Parse()

	switch strings.ToLower(*mode) {
	case "match":
		handleMatch(*serverURL, *interns, *positions, *edges)
	case "query":
		handleQuery(*serverURL, *queryIntern)
	default:
		fmt.Fprintf(os.Stderr, "Unknown mode: %s (use 'match' or 'query')\n", *mode)
		os.Exit(1)
	}
}

func handleMatch(serverURL, internsStr, positionsStr, edgesStr string) {
	interns := splitTrim(internsStr)
	positions := splitTrim(positionsStr)
	edges := parseEdges(edgesStr)

	req := api.MatchRequest{
		Interns:   interns,
		Positions: positions,
		Edges:     edges,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/match", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error calling server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Server error: %s\n", string(respBody))
		os.Exit(1)
	}

	var result api.MatchResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Maximum matching: %d\n", result.MaxMatch)
	if len(result.Assignments) > 0 {
		fmt.Println("Assignments:")
		for _, a := range result.Assignments {
			fmt.Printf("  %s -> %s\n", a.Intern, a.Position)
		}
	}
}

func handleQuery(serverURL, internID string) {
	if internID == "" {
		fmt.Fprintln(os.Stderr, "Please provide -intern flag for query mode")
		os.Exit(1)
	}

	req := api.QueryRequest{Intern: internID}
	body, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/query", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error calling server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Server error: %s\n", string(respBody))
		os.Exit(1)
	}

	var result api.QueryResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if result.Matched {
		fmt.Printf("Intern %s is matched to position %s\n", internID, result.Position)
	} else {
		fmt.Printf("Intern %s is not matched\n", internID)
	}
}

func splitTrim(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func parseEdges(s string) []api.Edge {
	if s == "" {
		return []api.Edge{}
	}
	pairs := strings.Split(s, ",")
	edges := make([]api.Edge, 0, len(pairs))
	for _, pair := range pairs {
		trimmed := strings.TrimSpace(pair)
		if trimmed == "" {
			continue
		}
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) == 2 {
			intern := strings.TrimSpace(parts[0])
			pos := strings.TrimSpace(parts[1])
			if intern != "" && pos != "" {
				edges = append(edges, api.Edge{
					Intern:   intern,
					Position: pos,
				})
			}
		}
	}
	return edges
}
