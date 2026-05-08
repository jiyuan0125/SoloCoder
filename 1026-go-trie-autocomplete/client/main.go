package main

import (
	"autocomplete/common"
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type Client struct {
	serverURL string
}

func NewClient(serverURL string) *Client {
	return &Client{
		serverURL: strings.TrimRight(serverURL, "/"),
	}
}

func (c *Client) Search(prefix string, limit int) (*common.SearchResponse, error) {
	req := common.SearchRequest{
		Prefix: prefix,
		Limit:  limit,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.serverURL+"/search", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) Add(word string) (*common.Response, error) {
	req := common.AddRequest{
		Word: word,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.serverURL+"/add", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) Delete(word string) (*common.Response, error) {
	req := common.DeleteRequest{
		Word: word,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.serverURL+"/delete", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	flag.Parse()

	client := NewClient(*serverURL)

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Autocomplete Client")
	fmt.Println("Commands:")
	fmt.Println("  search <prefix> [limit] - Search for words matching prefix")
	fmt.Println("  add <word>              - Add a new word")
	fmt.Println("  delete <word>           - Delete a word")
	fmt.Println("  exit                    - Exit the client")
	fmt.Println()

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cmd := parts[0]

		switch cmd {
		case "exit", "quit":
			fmt.Println("Goodbye!")
			return

		case "search":
			if len(parts) < 2 {
				fmt.Println("Usage: search <prefix> [limit]")
				continue
			}

			prefix := parts[1]
			limit := 10
			if len(parts) >= 3 {
				fmt.Sscanf(parts[2], "%d", &limit)
			}

			result, err := client.Search(prefix, limit)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}

			if len(result.Words) == 0 {
				fmt.Println("No results found.")
			} else {
				fmt.Printf("Found %d result(s):\n", len(result.Words))
				for i, word := range result.Words {
					fmt.Printf("  %d. %s\n", i+1, word)
				}
			}

		case "add":
			if len(parts) < 2 {
				fmt.Println("Usage: add <word>")
				continue
			}

			word := strings.Join(parts[1:], "")
			result, err := client.Add(word)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}

			if result.Success {
				fmt.Printf("Success: %s\n", result.Message)
			} else {
				fmt.Printf("Failed: %s\n", result.Message)
			}

		case "delete":
			if len(parts) < 2 {
				fmt.Println("Usage: delete <word>")
				continue
			}

			word := strings.Join(parts[1:], "")
			result, err := client.Delete(word)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}

			if result.Success {
				fmt.Printf("Success: %s\n", result.Message)
			} else {
				fmt.Printf("Failed: %s\n", result.Message)
			}

		default:
			fmt.Printf("Unknown command: %s\n", cmd)
			fmt.Println("Available commands: search, add, delete, exit")
		}

		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
	}
}
