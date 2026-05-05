package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"godiff/pkg/common"
)

var serverURL string

func main() {
	flag.StringVar(&serverURL, "server", "http://localhost:8080", "Diff server URL")
	oldFile := flag.String("old", "", "Old file path")
	newFile := flag.String("new", "", "New file path")
	moveThreshold := flag.Int("threshold", 20, "Move detection threshold (chars)")
	showUnified := flag.Bool("unified", false, "Show unified diff format")
	showStats := flag.Bool("stats", true, "Show statistics")
	showJSON := flag.Bool("json", false, "Output raw JSON response")

	flag.Parse()

	if *oldFile == "" || *newFile == "" {
		fmt.Println("Usage: diff-client -old <old-file> -new <new-file> [options]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	oldContent, err := os.ReadFile(*oldFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading old file: %v\n", err)
		os.Exit(1)
	}

	newContent, err := os.ReadFile(*newFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading new file: %v\n", err)
		os.Exit(1)
	}

	req := common.CompareRequest{
		OldText:       string(oldContent),
		NewText:       string(newContent),
		MoveThreshold: *moveThreshold,
	}

	resp, err := sendRequest(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "Server error: %s\n", resp.Error)
		os.Exit(1)
	}

	if *showJSON {
		jsonData, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(jsonData))
		return
	}

	if *showUnified && resp.UnifiedDiff != "" {
		fmt.Println(resp.UnifiedDiff)
	} else if !*showUnified {
		printDetailedDiffs(resp)
	}

	if *showStats {
		fmt.Println("\n--- Statistics ---")
		fmt.Printf("Added lines:    %d\n", resp.Stats.Added)
		fmt.Printf("Deleted lines:  %d\n", resp.Stats.Deleted)
		fmt.Printf("Changed lines:  %d\n", resp.Stats.Changed)
		fmt.Printf("Moved lines:    %d\n", resp.Stats.Moved)
		total := resp.Stats.Added + resp.Stats.Deleted + resp.Stats.Changed + resp.Stats.Moved
		fmt.Printf("Total changes:  %d\n", total)
	}
}

func sendRequest(req common.CompareRequest) (*common.CompareResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(serverURL+"/api/compare", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result common.CompareResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

func printDetailedDiffs(resp *common.CompareResponse) {
	for _, ld := range resp.LineDiffs {
		switch ld.Type {
		case common.DiffEqual:
			continue
		case common.DiffAdd:
			fmt.Printf("+ [%d] %s\n", ld.NewLine, ld.Content)
		case common.DiffDelete:
			fmt.Printf("- [%d] %s\n", ld.OldLine, ld.Content)
		case common.DiffChange:
			fmt.Printf("~ [%d->%d] ", ld.OldLine, ld.NewLine)
			if len(ld.WordDiffs) > 0 {
				for _, wd := range ld.WordDiffs {
					switch wd.Type {
					case common.DiffDelete:
						fmt.Printf("[-%s-]", wd.Content)
					case common.DiffAdd:
						fmt.Printf("[+%s+]", wd.Content)
					default:
						fmt.Print(wd.Content)
					}
				}
				fmt.Println()
			} else {
				fmt.Println(ld.Content)
			}
		case common.DiffMove:
			fmt.Printf("M [%d->%d] %s\n", ld.OldLine, ld.NewLine, ld.Content)
		}
	}

	if len(resp.MovedBlocks) > 0 {
		fmt.Println("\n--- Moved Blocks ---")
		for i, block := range resp.MovedBlocks {
			fmt.Printf("Block %d: lines [%d-%d] -> [%d-%d]\n",
				i+1, block.OldStart, block.OldEnd, block.NewStart, block.NewEnd)
		}
	}
}
