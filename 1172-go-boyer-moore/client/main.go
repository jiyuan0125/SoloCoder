package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/solocoder/boyermoore/shared"
)

const DefaultServerURL = "http://localhost:8301"

func getServerURL() string {
	if url := os.Getenv("BM_SERVER"); url != "" {
		return url
	}
	return DefaultServerURL
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "search":
		searchCmd(args)
	case "profile":
		profileCmd(args)
	case "bench":
		benchCmd(args)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("bmclient - Boyer-Moore string search client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  bmclient search -text <text> -pattern <pattern>")
	fmt.Println("  bmclient profile -pattern <pattern>")
	fmt.Println("  bmclient bench -text <text> -patterns <p1,p2,...>")
	fmt.Println()
	fmt.Println("Environment variable:")
	fmt.Println("  BM_SERVER    Server URL (default: http://localhost:8301)")
}

func searchCmd(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	textPtr := fs.String("text", "", "text to search in")
	patternPtr := fs.String("pattern", "", "pattern to search for")
	fs.Parse(args)

	if *textPtr == "" || *patternPtr == "" {
		fmt.Fprintln(os.Stderr, "error: -text and -pattern are required")
		os.Exit(1)
	}

	reqBody, err := json.Marshal(shared.SearchRequest{
		Text:    *textPtr,
		Pattern: *patternPtr,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(getServerURL()+"/search", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "error: server returned %d: %s\n", resp.StatusCode, body)
		os.Exit(1)
	}

	var result shared.SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Pattern: %q\n", result.Pattern)
	fmt.Printf("Text length: %d\n", result.TextLen)
	fmt.Printf("Matches found: %d\n", result.Count)
	if result.Count > 0 {
		fmt.Printf("Positions: %v\n", result.Matches)
	}
}

func profileCmd(args []string) {
	fs := flag.NewFlagSet("profile", flag.ExitOnError)
	patternPtr := fs.String("pattern", "", "pattern to preprocess")
	fs.Parse(args)

	if *patternPtr == "" {
		fmt.Fprintln(os.Stderr, "error: -pattern is required")
		os.Exit(1)
	}

	reqBody, err := json.Marshal(shared.PreprocessRequest{
		Pattern: *patternPtr,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(getServerURL()+"/preprocess", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "error: server returned %d: %s\n", resp.StatusCode, body)
		os.Exit(1)
	}

	var result shared.PreprocessResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Pattern: %q\n", result.Pattern)
	fmt.Printf("Length: %d\n", result.PatternLen)
	fmt.Println()
	fmt.Println("Bad Character Table (bmBc):")
	fmt.Println("  Only characters present in pattern are shown.")
	fmt.Println("  Others default to pattern length.")
	w := tabwriter.NewWriter(os.Stdout, 4, 4, 2, ' ', 0)
	if len(result.BmBc) > 0 {
		keys := make([]rune, 0, len(result.BmBc))
		for k := range result.BmBc {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		for _, k := range keys {
			fmt.Fprintf(w, "  %q\t->\t%d\n", k, result.BmBc[k])
		}
	} else {
		fmt.Println("  (empty)")
	}
	w.Flush()
	fmt.Println()
	fmt.Println("Good Suffix Table (bmGs):")
	fmt.Printf("  Index\tShift\n")
	for i, v := range result.BmGs {
		fmt.Printf("  %d\t%d\n", i, v)
	}
}

func benchCmd(args []string) {
	fs := flag.NewFlagSet("bench", flag.ExitOnError)
	textPtr := fs.String("text", "", "text to search in")
	patternsPtr := fs.String("patterns", "", "comma-separated list of patterns")
	runsPtr := fs.Int("runs", 10, "number of runs per pattern")
	fs.Parse(args)

	if *textPtr == "" || *patternsPtr == "" {
		fmt.Fprintln(os.Stderr, "error: -text and -patterns are required")
		os.Exit(1)
	}

	patterns := splitPatterns(*patternsPtr)
	if len(patterns) == 0 {
		fmt.Fprintln(os.Stderr, "error: no patterns specified")
		os.Exit(1)
	}

	serverURL := getServerURL()
	fmt.Printf("Benchmarking against server: %s\n", serverURL)
	fmt.Printf("Text length: %d, runs per pattern: %d\n", len(*textPtr), *runsPtr)
	fmt.Println()

	results := make([]shared.BenchmarkResult, len(patterns))

	for i, p := range patterns {
		count, nsOp := runBenchmark(serverURL, *textPtr, p, *runsPtr)
		results[i] = shared.BenchmarkResult{
			Pattern: p,
			Count:   count,
			NsOp:    nsOp,
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 4, 4, 2, ' ', 0)
	fmt.Fprintln(w, "Pattern\tMatches\tTime/op")
	fmt.Fprintln(w, "-------\t-------\t-------")
	for _, r := range results {
		timeStr := formatDuration(r.NsOp)
		fmt.Fprintf(w, "%q\t%d\t%s\n", r.Pattern, r.Count, timeStr)
	}
	w.Flush()
}

func runBenchmark(serverURL, text, pattern string, runs int) (int, int64) {
	reqBody, err := json.Marshal(shared.SearchRequest{
		Text:    text,
		Pattern: pattern,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var count int
	start := time.Now()
	for i := 0; i < runs; i++ {
		resp, err := http.Post(serverURL+"/search", "application/json", bytes.NewReader(reqBody))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		var result shared.SearchResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		resp.Body.Close()
		count = result.Count
	}
	elapsed := time.Since(start)
	perOp := elapsed.Nanoseconds() / int64(runs)
	return count, perOp
}

func splitPatterns(s string) []string {
	if s == "" {
		return nil
	}
	var result []string
	var buf bytes.Buffer
	for _, c := range s {
		if c == ',' {
			if buf.Len() > 0 {
				result = append(result, buf.String())
				buf.Reset()
			}
		} else {
			buf.WriteRune(c)
		}
	}
	if buf.Len() > 0 {
		result = append(result, buf.String())
	}
	return result
}

func formatDuration(ns int64) string {
	switch {
	case ns >= 1000000000:
		return fmt.Sprintf("%.2fs", float64(ns)/1e9)
	case ns >= 1000000:
		return fmt.Sprintf("%.2fms", float64(ns)/1e6)
	case ns >= 1000:
		return fmt.Sprintf("%.2fµs", float64(ns)/1e3)
	default:
		return fmt.Sprintf("%dns", ns)
	}
}
