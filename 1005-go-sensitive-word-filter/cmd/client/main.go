package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sensitive-word-filter/pkg/common"
	"strings"
)

const serverURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "filter":
		if len(os.Args) < 3 {
			fmt.Println("Usage: client filter <text>")
			return
		}
		handleFilter(strings.Join(os.Args[2:], " "))
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Usage: client add <word>")
			return
		}
		handleAdd(strings.Join(os.Args[2:], " "))
	case "remove":
		if len(os.Args) < 3 {
			fmt.Println("Usage: client remove <word>")
			return
		}
		handleRemove(strings.Join(os.Args[2:], " "))
	case "update":
		if len(os.Args) < 4 {
			fmt.Println("Usage: client update <old_word> <new_word>")
			return
		}
		handleUpdate(os.Args[2], os.Args[3])
	case "list":
		handleList()
	case "add-whitelist":
		if len(os.Args) < 3 {
			fmt.Println("Usage: client add-whitelist <word>")
			return
		}
		handleAddWhitelist(strings.Join(os.Args[2:], " "))
	case "remove-whitelist":
		if len(os.Args) < 3 {
			fmt.Println("Usage: client remove-whitelist <word>")
			return
		}
		handleRemoveWhitelist(strings.Join(os.Args[2:], " "))
	case "list-whitelist":
		handleListWhitelist()
	case "stats":
		handleStats()
	case "interactive":
		handleInteractive()
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage: client <command> [args]")
	fmt.Println("Commands:")
	fmt.Println("  filter <text>          - Filter text and show results")
	fmt.Println("  add <word>             - Add a sensitive word")
	fmt.Println("  remove <word>          - Remove a sensitive word")
	fmt.Println("  update <old> <new>     - Update a sensitive word")
	fmt.Println("  list                   - List all sensitive words")
	fmt.Println("  add-whitelist <word>   - Add word to whitelist")
	fmt.Println("  remove-whitelist <word>- Remove word from whitelist")
	fmt.Println("  list-whitelist         - List all whitelist words")
	fmt.Println("  stats                  - Show statistics")
	fmt.Println("  interactive            - Interactive mode")
}

func handleFilter(text string) {
	reqBody, _ := json.Marshal(common.FilterRequest{Text: text})
	resp, err := http.Post(serverURL+"/api/filter", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result common.FilterResponse
	json.Unmarshal(body, &result)

	printFilterResult(&result)
}

func handleAdd(word string) {
	reqBody, _ := json.Marshal(common.AddSensitiveWordRequest{Word: word})
	resp, err := http.Post(serverURL+"/api/sensitive-words/add", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result common.GenericResponse
	json.Unmarshal(body, &result)

	printGenericResponse(&result)
}

func handleRemove(word string) {
	reqBody, _ := json.Marshal(common.DeleteSensitiveWordRequest{Word: word})
	resp, err := http.Post(serverURL+"/api/sensitive-words/remove", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result common.GenericResponse
	json.Unmarshal(body, &result)

	printGenericResponse(&result)
}

func handleUpdate(oldWord, newWord string) {
	reqBody, _ := json.Marshal(common.UpdateSensitiveWordRequest{OldWord: oldWord, NewWord: newWord})
	resp, err := http.Post(serverURL+"/api/sensitive-words/update", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result common.GenericResponse
	json.Unmarshal(body, &result)

	printGenericResponse(&result)
}

func handleList() {
	resp, err := http.Get(serverURL + "/api/sensitive-words")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result common.ListSensitiveWordsResponse
	json.Unmarshal(body, &result)

	fmt.Printf("Total sensitive words: %d\n", result.Total)
	for i, word := range result.Words {
		fmt.Printf("  %d. %s\n", i+1, word)
	}
}

func handleAddWhitelist(word string) {
	reqBody, _ := json.Marshal(common.AddWhitelistRequest{Word: word})
	resp, err := http.Post(serverURL+"/api/whitelist/add", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result common.GenericResponse
	json.Unmarshal(body, &result)

	printGenericResponse(&result)
}

func handleRemoveWhitelist(word string) {
	reqBody, _ := json.Marshal(common.DeleteWhitelistRequest{Word: word})
	resp, err := http.Post(serverURL+"/api/whitelist/remove", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result common.GenericResponse
	json.Unmarshal(body, &result)

	printGenericResponse(&result)
}

func handleListWhitelist() {
	resp, err := http.Get(serverURL + "/api/whitelist")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result common.ListWhitelistResponse
	json.Unmarshal(body, &result)

	fmt.Printf("Total whitelist words: %d\n", result.Total)
	for i, word := range result.Words {
		fmt.Printf("  %d. %s\n", i+1, word)
	}
}

func handleStats() {
	resp, err := http.Get(serverURL + "/api/stats")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result common.StatsResponse
	json.Unmarshal(body, &result)

	fmt.Printf("Total checks: %d\n", result.TotalTotal)
	fmt.Printf("Total hits: %d\n", result.TotalHits)
	if result.TotalTotal > 0 {
		fmt.Printf("Hit rate: %.2f%%\n", float64(result.TotalHits)/float64(result.TotalTotal)*100)
	}

	fmt.Println("\nDaily Statistics:")
	for _, ds := range result.DailyStats {
		fmt.Printf("  %s: total=%d, hits=%d\n", ds.Date, ds.Total, ds.HitCount)
	}

	fmt.Println("\nTop Sensitive Words:")
	for i, tw := range result.TopHits {
		fmt.Printf("  %d. %s: %d hits\n", i+1, tw.Word, tw.Count)
	}
}

func handleInteractive() {
	fmt.Println("Interactive Mode - Type 'quit' or 'exit' to leave")
	fmt.Println("Enter text to filter:")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		text := scanner.Text()
		if text == "quit" || text == "exit" {
			break
		}
		if strings.TrimSpace(text) == "" {
			continue
		}
		handleFilter(text)
		fmt.Println()
	}
}

func printFilterResult(result *common.FilterResponse) {
	fmt.Printf("Text: %s\n", result.Text)
	if result.Hit {
		fmt.Printf("Status: HIT (%d sensitive words found)\n", len(result.Hits))
		fmt.Println("\nDetails:")
		for i, hit := range result.Hits {
			fmt.Printf("  %d. Word: %s (Type: %s)\n", i+1, hit.SensitiveWord, hit.MatchType)
			fmt.Printf("     Positions:\n")
			for j, pos := range hit.Positions {
				fmt.Printf("       %d. Start=%d, Length=%d\n", j+1, pos.Start, pos.Length)
			}
		}
	} else {
		fmt.Println("Status: OK (no sensitive words found)")
	}
}

func printGenericResponse(result *common.GenericResponse) {
	if result.Success {
		fmt.Printf("Success: %s\n", result.Message)
	} else {
		fmt.Printf("Error: %s\n", result.Message)
	}
}
