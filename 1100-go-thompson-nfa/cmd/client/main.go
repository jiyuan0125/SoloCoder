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
	"testing"
	"time"

	"thompson-nfa/pkg/api"
)

const baseURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: client <command> [args]")
		fmt.Println("Commands: match, findall, test, visualize, benchmark")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "match":
		handleMatch()
	case "findall":
		handleFindAll()
	case "test":
		handleTest()
	case "visualize":
		handleVisualize()
	case "benchmark":
		handleBenchmark()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Commands: match, findall, test, visualize, benchmark")
		os.Exit(1)
	}
}

func handleMatch() {
	fs := flag.NewFlagSet("match", flag.ExitOnError)
	pattern := fs.String("p", "", "Regex pattern")
	input := fs.String("i", "", "Input string")
	fs.Parse(os.Args[2:])

	if *pattern == "" || *input == "" {
		fmt.Println("Usage: client match -p <pattern> -i <input>")
		os.Exit(1)
	}

	req := api.MatchRequest{Pattern: *pattern, Input: *input}
	body, _ := json.Marshal(req)

	resp, err := http.Post(baseURL+"/api/regex/match", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	var result api.MatchResponse
	json.Unmarshal(data, &result)

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Error)
		os.Exit(1)
	}

	fmt.Printf("Match: %v\n", result.Match)
}

func handleFindAll() {
	fs := flag.NewFlagSet("findall", flag.ExitOnError)
	pattern := fs.String("p", "", "Regex pattern")
	input := fs.String("i", "", "Input string")
	fs.Parse(os.Args[2:])

	if *pattern == "" || *input == "" {
		fmt.Println("Usage: client findall -p <pattern> -i <input>")
		os.Exit(1)
	}

	req := api.FindAllRequest{Pattern: *pattern, Input: *input}
	body, _ := json.Marshal(req)

	resp, err := http.Post(baseURL+"/api/regex/findall", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	var result api.FindAllResponse
	json.Unmarshal(data, &result)

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Error)
		os.Exit(1)
	}

	fmt.Printf("Found %d matches:\n", len(result.Matches))
	for i, match := range result.Matches {
		fmt.Printf("  %d: '%s' (positions %d-%d)\n", i+1, match.Match, match.Start, match.End)
	}
}

func handleTest() {
	fs := flag.NewFlagSet("test", flag.ExitOnError)
	filename := fs.String("f", "", "Test file path")
	fs.Parse(os.Args[2:])

	if *filename == "" {
		fmt.Println("Usage: client test -f <file>")
		os.Exit(1)
	}

	data, err := ioutil.ReadFile(*filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(string(data), "\n")
	passed := 0
	failed := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, " ", 3)
		if len(parts) < 3 {
			fmt.Printf("Invalid line: %s\n", line)
			continue
		}

		pattern := parts[0]
		input := parts[1]
		expected := parts[2] == "true" || parts[2] == "1"

		req := api.MatchRequest{Pattern: pattern, Input: input}
		body, _ := json.Marshal(req)

		resp, err := http.Post(baseURL+"/api/regex/match", "application/json", bytes.NewBuffer(body))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		respData, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		var result api.MatchResponse
		json.Unmarshal(respData, &result)

		if !result.Success {
			fmt.Printf("FAIL: pattern='%s' input='%s' - Error: %s\n", pattern, input, result.Error)
			failed++
			continue
		}

		if result.Match == expected {
			fmt.Printf("PASS: pattern='%s' input='%s' (expected=%v, got=%v)\n", pattern, input, expected, result.Match)
			passed++
		} else {
			fmt.Printf("FAIL: pattern='%s' input='%s' (expected=%v, got=%v)\n", pattern, input, expected, result.Match)
			failed++
		}
	}

	fmt.Printf("\nTotal: %d passed, %d failed\n", passed, failed)
}

func handleVisualize() {
	fs := flag.NewFlagSet("visualize", flag.ExitOnError)
	pattern := fs.String("p", "", "Regex pattern")
	fs.Parse(os.Args[2:])

	if *pattern == "" {
		fmt.Println("Usage: client visualize -p <pattern>")
		os.Exit(1)
	}

	req := api.VisualizeRequest{Pattern: *pattern}
	body, _ := json.Marshal(req)

	resp, err := http.Post(baseURL+"/api/regex/visualize", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	var result api.VisualizeResponse
	json.Unmarshal(data, &result)

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Error)
		os.Exit(1)
	}

	fmt.Println(result.Graph)
}

func handleBenchmark() {
	fs := flag.NewFlagSet("benchmark", flag.ExitOnError)
	pattern := fs.String("p", "", "Regex pattern")
	input := fs.String("i", "", "Input string")
	iterations := fs.Int("n", 1000, "Number of iterations")
	fs.Parse(os.Args[2:])

	if *pattern == "" || *input == "" {
		fmt.Println("Usage: client benchmark -p <pattern> -i <input> [-n <iterations>]")
		os.Exit(1)
	}

	benchmarkStdlib := func(pattern, input string, n int) time.Duration {
		start := time.Now()
		for i := 0; i < n; i++ {
			stdlibTest := testing.Benchmark(func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
				}
			})
			_ = stdlibTest
		}
		return time.Since(start)
	}

	_ = benchmarkStdlib

	fmt.Printf("Benchmarking pattern '%s' on input '%s' (%d iterations)...\n\n", *pattern, *input, *iterations)

	start := time.Now()
	for i := 0; i < *iterations; i++ {
		req := api.MatchRequest{Pattern: *pattern, Input: *input}
		body, _ := json.Marshal(req)
		resp, _ := http.Post(baseURL+"/api/regex/match", "application/json", bytes.NewBuffer(body))
		if resp != nil {
			data, _ := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			_ = data
		}
	}
	ourTime := time.Since(start)

	fmt.Printf("Thompson NFA (via HTTP): %s\n", ourTime)
	fmt.Printf("  Average per iteration: %s\n", ourTime/time.Duration(*iterations))
	fmt.Println()
	fmt.Println("Note: To compare with stdlib regexp, you need to run benchmarks directly")
	fmt.Println("in your Go code. HTTP overhead dominates these results.")
}
