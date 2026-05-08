package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"fulltext-search/api"
)

const serverURL = "http://localhost:8080"

func addDoc(docID, content string) error {
	req := api.AddDocRequest{
		DocID:   docID,
		Content: content,
	}
	return sendRequest("/add", req, nil)
}

func deleteDoc(docID string) error {
	req := api.DeleteDocRequest{
		DocID: docID,
	}
	return sendRequest("/delete", req, nil)
}

func search(query string) (*api.SearchResponse, error) {
	req := api.SearchRequest{
		Query: query,
	}
	var resp api.SearchResponse
	if err := sendRequest("/search", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func sendRequest(path string, body interface{}, result interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(serverURL+path, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  add <doc_id> <content>  - Add a document")
	fmt.Println("  delete <doc_id>         - Delete a document")
	fmt.Println("  search <query>          - Search documents")
	fmt.Println("  help                    - Show this help")
	fmt.Println("  quit                    - Exit the program")
	fmt.Println()
	fmt.Println("Query syntax:")
	fmt.Println("  AND, OR, NOT - Boolean operators (UPPERCASE)")
	fmt.Println("  ()           - Parentheses for grouping")
	fmt.Println("  Example: (中 OR 英) AND NOT 文")
}

func handleCommand(line string) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return
	}

	cmd := strings.ToLower(parts[0])
	switch cmd {
	case "help":
		printHelp()
	case "quit", "exit":
		fmt.Println("Goodbye!")
		os.Exit(0)
	case "add":
		if len(parts) < 3 {
			fmt.Println("Usage: add <doc_id> <content>")
			return
		}
		docID := parts[1]
		content := strings.Join(parts[2:], " ")
		if err := addDoc(docID, content); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Document '%s' added successfully.\n", docID)
	case "delete":
		if len(parts) < 2 {
			fmt.Println("Usage: delete <doc_id>")
			return
		}
		docID := parts[1]
		if err := deleteDoc(docID); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Document '%s' deleted successfully.\n", docID)
	case "search":
		if len(parts) < 2 {
			fmt.Println("Usage: search <query>")
			return
		}
		query := strings.Join(parts[1:], " ")
		result, err := search(query)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if !result.Success {
			fmt.Printf("Search failed: %s\n", result.Message)
			return
		}
		if len(result.Results) == 0 {
			fmt.Println("No results found.")
			return
		}
		fmt.Printf("Found %d result(s):\n", len(result.Results))
		for i, item := range result.Results {
			fmt.Printf("  %d. %s (score: %.6f)\n", i+1, item.DocID, item.Score)
		}
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		fmt.Println("Type 'help' for available commands.")
	}
}

func main() {
	fmt.Println("Full-text Search Client")
	fmt.Println("Type 'help' for available commands.")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		handleCommand(line)
	}
}
