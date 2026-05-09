package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/kmpmatcher/api"
)

var serverURL = "http://localhost:8080"

func main() {
	mode := flag.String("mode", "single", "match mode: single or multi")
	pattern := flag.String("pattern", "", "pattern string for single mode")
	patterns := flag.String("patterns", "", "comma-separated patterns for multi mode")
	text := flag.String("text", "", "text string to search")
	textFile := flag.String("text-file", "", "read text from file")
	caseInsensitive := flag.Bool("i", false, "case insensitive match")
	flag.Parse()

	var content string
	if *textFile != "" {
		data, err := os.ReadFile(*textFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
			os.Exit(1)
		}
		content = string(data)
	} else {
		content = *text
	}

	if *mode == "single" {
		if *pattern == "" {
			fmt.Fprintln(os.Stderr, "pattern is required for single mode")
			os.Exit(1)
		}
		doSingle(content, *pattern, *caseInsensitive)
	} else if *mode == "multi" {
		if *patterns == "" {
			fmt.Fprintln(os.Stderr, "patterns is required for multi mode")
			os.Exit(1)
		}
		patternList := strings.Split(*patterns, ",")
		doMulti(content, patternList, *caseInsensitive)
	} else {
		fmt.Fprintln(os.Stderr, "invalid mode, use single or multi")
		os.Exit(1)
	}
}

func doSingle(text, pattern string, caseInsensitive bool) {
	req := api.SingleMatchRequest{
		Text:            text,
		Pattern:         pattern,
		CaseInsensitive: caseInsensitive,
	}
	body, _ := json.Marshal(req)

	resp, err := http.Post(serverURL+"/single", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result api.SingleMatchResponse
	json.Unmarshal(data, &result)

	fmt.Printf("Pattern: %s\n", pattern)
	fmt.Printf("Positions: %v\n", result.Positions)
}

func doMulti(text string, patterns []string, caseInsensitive bool) {
	req := api.MultiMatchRequest{
		Text:            text,
		Patterns:        patterns,
		CaseInsensitive: caseInsensitive,
	}
	body, _ := json.Marshal(req)

	resp, err := http.Post(serverURL+"/multi", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result api.MultiMatchResponse
	json.Unmarshal(data, &result)

	for _, p := range patterns {
		fmt.Printf("Pattern: %s\n", p)
		fmt.Printf("Positions: %v\n\n", result.Results[p])
	}
}
