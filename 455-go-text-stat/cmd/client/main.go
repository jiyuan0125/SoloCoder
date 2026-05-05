package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"textstat/pkg/common"
)

var serverURL = "http://localhost:8080"

func printHelp() {
	fmt.Println("Text Statistics Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client stats --text \"your text\" [--top-n 10]")
	fmt.Println("  client similarity --text1 \"first text\" --text2 \"second text\"")
	fmt.Println("  client help")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  stats       - Analyze text statistics")
	fmt.Println("  similarity  - Calculate similarity between two texts")
	fmt.Println("  help        - Show this help message")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --text      - Text to analyze (for stats command)")
	fmt.Println("  --text1     - First text to compare (for similarity command)")
	fmt.Println("  --text2     - Second text to compare (for similarity command)")
	fmt.Println("  --top-n     - Number of top frequent words to show (default: 10)")
	fmt.Println("  --server    - Server URL (default: http://localhost:8080)")
	fmt.Println()
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "help", "--help", "-h":
		printHelp()
	case "stats":
		handleStats()
	case "similarity":
		handleSimilarity()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}

func handleStats() {
	statsFlag := flag.NewFlagSet("stats", flag.ExitOnError)
	textFlag := statsFlag.String("text", "", "Text to analyze")
	topNFlag := statsFlag.Int("top-n", 10, "Number of top frequent words")
	serverFlag := statsFlag.String("server", serverURL, "Server URL")

	statsFlag.Parse(os.Args[2:])

	if *textFlag == "" {
		fmt.Println("Error: --text is required")
		os.Exit(1)
	}

	if *serverFlag != "" {
		serverURL = *serverFlag
	}

	req := common.TextStatsRequest{
		Text: *textFlag,
		TopN: *topNFlag,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: failed to create request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/api/stats", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: failed to read response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.Unmarshal(respBody, &errResp)
		fmt.Printf("Error: %s\n", errResp.Error)
		os.Exit(1)
	}

	var statsResp common.TextStatsResponse
	if err := json.Unmarshal(respBody, &statsResp); err != nil {
		fmt.Printf("Error: failed to parse response: %v\n", err)
		os.Exit(1)
	}

	printStatsResponse(&statsResp)
}

func handleSimilarity() {
	simFlag := flag.NewFlagSet("similarity", flag.ExitOnError)
	text1Flag := simFlag.String("text1", "", "First text")
	text2Flag := simFlag.String("text2", "", "Second text")
	serverFlag := simFlag.String("server", serverURL, "Server URL")

	simFlag.Parse(os.Args[2:])

	if *text1Flag == "" || *text2Flag == "" {
		fmt.Println("Error: --text1 and --text2 are required")
		os.Exit(1)
	}

	if *serverFlag != "" {
		serverURL = *serverFlag
	}

	req := common.SimilarityRequest{
		Text1: *text1Flag,
		Text2: *text2Flag,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: failed to create request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/api/similarity", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: failed to read response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.Unmarshal(respBody, &errResp)
		fmt.Printf("Error: %s\n", errResp.Error)
		os.Exit(1)
	}

	var simResp common.SimilarityResponse
	if err := json.Unmarshal(respBody, &simResp); err != nil {
		fmt.Printf("Error: failed to parse response: %v\n", err)
		os.Exit(1)
	}

	printSimilarityResponse(&simResp)
}

func printStatsResponse(resp *common.TextStatsResponse) {
	fmt.Println("========================================")
	fmt.Println("          Text Statistics Result        ")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Printf("Character Count (with spaces):    %d\n", resp.CharCountWithSpaces)
	fmt.Printf("Character Count (without spaces): %d\n", resp.CharCountWithoutSpaces)
	fmt.Printf("Word Count:                        %d\n", resp.WordCount)
	fmt.Printf("Line Count:                        %d\n", resp.LineCount)
	fmt.Printf("Paragraph Count:                   %d\n", resp.ParagraphCount)
	fmt.Println()
	
	if len(resp.TopWords) > 0 {
		fmt.Println("Top Frequent Words:")
		fmt.Println("----------------------------------------")
		for i, w := range resp.TopWords {
			fmt.Printf("  %2d. %-20s (count: %d)\n", i+1, w.Word, w.Count)
		}
	}
	fmt.Println("========================================")
}

func printSimilarityResponse(resp *common.SimilarityResponse) {
	fmt.Println("========================================")
	fmt.Println("         Text Similarity Result         ")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Printf("Cosine Similarity (TF-IDF):  %.4f\n", resp.Similarity)
	fmt.Println()
	
	level := ""
	switch {
	case resp.Similarity >= 0.9:
		level = "Extremely High - Almost identical"
	case resp.Similarity >= 0.7:
		level = "High - Significant overlap"
	case resp.Similarity >= 0.5:
		level = "Moderate - Some similarity"
	case resp.Similarity >= 0.3:
		level = "Low - Minor similarity"
	default:
		level = "Very Low - Mostly different"
	}
	fmt.Printf("Similarity Level:              %s\n", level)
	fmt.Println()
	fmt.Println("========================================")
}
