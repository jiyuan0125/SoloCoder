package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"bktree-app/internal/models"
)

var serverURL string

func main() {
	flag.StringVar(&serverURL, "server", "http://localhost:8201", "BK tree server URL")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	cmd := args[0]

	switch cmd {
	case "add":
		if len(args) < 2 {
			fmt.Println("Usage: add <word>")
			os.Exit(1)
		}
		addWord(args[1])

	case "search":
		if len(args) < 2 {
			fmt.Println("Usage: search <query> [max_distance]")
			os.Exit(1)
		}
		maxDistance := 2
		if len(args) >= 3 {
			fmt.Sscanf(args[2], "%d", &maxDistance)
		}
		searchWords(args[1], maxDistance)

	case "remove":
		if len(args) < 2 {
			fmt.Println("Usage: remove <word>")
			os.Exit(1)
		}
		removeWord(args[1])

	case "import":
		if len(args) < 2 {
			fmt.Println("Usage: import <filepath>")
			os.Exit(1)
		}
		importWords(args[1])

	case "info":
		getInfo()

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("BK Tree Client - Spell Checker using BK Tree")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client add <word>                Add a single word")
	fmt.Println("  client search <query> [k]        Search for similar words (default k=2)")
	fmt.Println("  client remove <word>             Remove a word")
	fmt.Println("  client import <filepath>         Import words from file (one per line)")
	fmt.Println("  client info                      Show tree info")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -server string    Server URL (default http://localhost:8201)")
}

func addWord(word string) {
	req := models.AddRequest{Word: word}
	body, _ := json.Marshal(req)

	resp, err := http.Post(serverURL+"/add", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result models.AddResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Success {
		fmt.Printf("Added: %s\n", word)
	}
}

func searchWords(query string, maxDistance int) {
	req := models.SearchRequest{Query: query, MaxDistance: maxDistance}
	body, _ := json.Marshal(req)

	resp, err := http.Post(serverURL+"/search", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result models.SearchResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if !result.Success {
		fmt.Println("Search failed")
		return
	}

	fmt.Printf("Search results for '%s' (max distance %d):\n", query, maxDistance)
	if len(result.Matches) == 0 {
		fmt.Println("  No matches found")
		return
	}

	for _, match := range result.Matches {
		fmt.Printf("  %s (distance: %d)\n", match.Word, match.Distance)
	}
}

func removeWord(word string) {
	req := models.RemoveRequest{Word: word}
	body, _ := json.Marshal(req)

	resp, err := http.Post(serverURL+"/remove", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result models.RemoveResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Success {
		if result.Removed {
			fmt.Printf("Removed: %s\n", word)
		} else {
			fmt.Printf("Not found: %s\n", word)
		}
	}
}

func importWords(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			words = append(words, line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	req := models.ImportRequest{Words: words}
	body, _ := json.Marshal(req)

	resp, err := http.Post(serverURL+"/import", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result models.ImportResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Success {
		fmt.Printf("Imported %d words\n", result.Imported)
	}
}

func getInfo() {
	resp, err := http.Get(serverURL + "/info")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result models.InfoResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Success {
		fmt.Printf("Tree Size: %d nodes\n", result.Size)
		fmt.Printf("Tree Depth: %d\n", result.Depth)
	}
}
