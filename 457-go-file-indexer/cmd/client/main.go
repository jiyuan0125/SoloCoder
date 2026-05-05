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

	"go-file-indexer/common"
)

var serverURL string

func main() {
	flag.StringVar(&serverURL, "server", "http://localhost:8080", "Server URL")
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]

	switch command {
	case "index":
		if len(args) < 2 {
			fmt.Println("Usage: fileindexer index <directory>")
			os.Exit(1)
		}
		indexDirectory(args[1])

	case "search":
		if len(args) < 2 {
			fmt.Println("Usage: fileindexer search <query>")
			os.Exit(1)
		}
		query := strings.Join(args[1:], " ")
		searchFiles(query)

	case "stats":
		getStats()

	case "update":
		updateIndex()

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Go File Indexer - Client")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  fileindexer [flags] <command> [args]")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  -server string   Server URL (default \"http://localhost:8080\")")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  index <directory>   Index a directory")
	fmt.Println("  search <query>      Search for keywords")
	fmt.Println("  stats               Show index statistics")
	fmt.Println("  update              Perform incremental update")
	fmt.Println("")
	fmt.Println("Notes:")
	fmt.Println("  - Use double quotes for phrase search: search \"hello world\"")
	fmt.Println("  - Search is case-insensitive")
}

func indexDirectory(dirPath string) {
	req := common.IndexRequest{
		Directory: dirPath,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/index", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	var result common.IndexResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Printf("Server response: %s\n", string(respBody))
		os.Exit(1)
	}

	if result.Success {
		fmt.Printf("✓ %s\n", result.Message)
	} else {
		fmt.Printf("✗ %s\n", result.Message)
		os.Exit(1)
	}
}

func searchFiles(query string) {
	searchURL := fmt.Sprintf("%s/search?q=%s", serverURL, query)
	resp, err := http.Get(searchURL)
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	var result common.SearchResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Printf("Server response: %s\n", string(respBody))
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("✗ %s\n", result.Message)
		os.Exit(1)
	}

	if len(result.Hits) == 0 {
		fmt.Println("No results found.")
		return
	}

	fmt.Printf("Found %d result(s)\n", len(result.Hits))
	fmt.Println("")

	for i, hit := range result.Hits {
		fmt.Printf("Result %d (score: %d)\n", i+1, hit.Score)
		fmt.Printf("  File: %s\n", hit.FileName)
		fmt.Printf("  Line: %d\n", hit.LineNumber)
		fmt.Printf("  Content: %s\n", hit.LineContent)
		fmt.Println("")
	}
}

func getStats() {
	resp, err := http.Get(serverURL + "/stats")
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	var result common.StatsResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Printf("Server response: %s\n", string(respBody))
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("✗ Error getting stats\n")
		os.Exit(1)
	}

	fmt.Println("Index Statistics:")
	fmt.Printf("  Indexed files: %d\n", result.IndexedFiles)
	fmt.Printf("  Indexed words: %d\n", result.IndexedWords)
	fmt.Printf("  Last indexed:  %s\n", result.LastIndexedTime)
}

func updateIndex() {
	resp, err := http.Post(serverURL+"/update", "application/json", bytes.NewReader([]byte("{}")))
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	var result common.IndexResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Printf("Server response: %s\n", string(respBody))
		os.Exit(1)
	}

	if result.Success {
		fmt.Printf("✓ %s\n", result.Message)
	} else {
		fmt.Printf("✗ %s\n", result.Message)
		os.Exit(1)
	}
}
