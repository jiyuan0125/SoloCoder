package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"double-array-trie/pkg/api"
)

var (
	serverURL string
)

type Command interface {
	Name() string
	Usage() string
	Run(args []string) error
}

type AddCommand struct{}

func (c *AddCommand) Name() string  { return "add" }
func (c *AddCommand) Usage() string { return "add <word> - Add a word to the trie" }
func (c *AddCommand) Run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: %s", c.Usage())
	}
	word := args[0]

	reqBody, _ := json.Marshal(&api.AddRequest{Word: word})
	resp, err := http.Post(serverURL+"/add", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result api.AddResponse
	json.Unmarshal(body, &result)

	if resp.StatusCode == http.StatusOK {
		if result.Success {
			fmt.Printf("Word '%s' added successfully\n", word)
		} else {
			fmt.Printf("Failed to add word: %s\n", result.Message)
		}
		return nil
	}

	var errResp api.ErrorResponse
	json.Unmarshal(body, &errResp)
	return fmt.Errorf(errResp.Error)
}

type SearchCommand struct{}

func (c *SearchCommand) Name() string  { return "search" }
func (c *SearchCommand) Usage() string { return "search <word> - Search if a word exists" }
func (c *SearchCommand) Run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: %s", c.Usage())
	}
	word := args[0]

	resp, err := http.Get(fmt.Sprintf("%s/search?word=%s", serverURL, word))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result api.SearchResponse
	json.Unmarshal(body, &result)

	if result.Exists {
		fmt.Printf("Word '%s' exists\n", word)
	} else {
		fmt.Printf("Word '%s' does not exist\n", word)
	}
	return nil
}

type PrefixCommand struct{}

func (c *PrefixCommand) Name() string  { return "prefix" }
func (c *PrefixCommand) Usage() string { return "prefix <prefix> - Find all words with given prefix" }
func (c *PrefixCommand) Run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: %s", c.Usage())
	}
	prefix := args[0]

	resp, err := http.Get(fmt.Sprintf("%s/prefix?prefix=%s", serverURL, prefix))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result api.PrefixResponse
	json.Unmarshal(body, &result)

	if len(result.Words) == 0 {
		fmt.Printf("No words found with prefix '%s'\n", prefix)
		return nil
	}

	fmt.Printf("Found %d word(s) with prefix '%s':\n", len(result.Words), prefix)
	for _, word := range result.Words {
		fmt.Printf("  %s\n", word)
	}
	return nil
}

type RemoveCommand struct{}

func (c *RemoveCommand) Name() string  { return "remove" }
func (c *RemoveCommand) Usage() string { return "remove <word> - Remove a word from the trie" }
func (c *RemoveCommand) Run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: %s", c.Usage())
	}
	word := args[0]

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/delete?word=%s", serverURL, word), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result api.DeleteResponse
	json.Unmarshal(body, &result)

	if result.Success {
		fmt.Printf("Word '%s' removed successfully\n", word)
	} else {
		fmt.Printf("Failed to remove word: %s\n", result.Message)
	}
	return nil
}

type ImportCommand struct{}

func (c *ImportCommand) Name() string  { return "import" }
func (c *ImportCommand) Usage() string { return "import <file> - Import words from a file (one word per line)" }
func (c *ImportCommand) Run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: %s", c.Usage())
	}
	filename := args[0]

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" {
			words = append(words, word)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	if len(words) == 0 {
		return fmt.Errorf("no words found in file")
	}

	reqBody, _ := json.Marshal(&api.ImportRequest{Words: words})
	resp, err := http.Post(serverURL+"/import", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result api.ImportResponse
	json.Unmarshal(body, &result)

	if result.Success {
		fmt.Printf("Successfully imported %d word(s)\n", result.Added)
	} else {
		return fmt.Errorf("import failed: %s", result.Message)
	}
	return nil
}

type StatCommand struct{}

func (c *StatCommand) Name() string  { return "stat" }
func (c *StatCommand) Usage() string { return "stat - Show trie statistics" }
func (c *StatCommand) Run(args []string) error {
	resp, err := http.Get(serverURL + "/stat")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result api.StatResponse
	json.Unmarshal(body, &result)

	fmt.Println("Trie Statistics:")
	fmt.Printf("  Base array size:   %d\n", result.BaseSize)
	fmt.Printf("  Check array size:  %d\n", result.CheckSize)
	fmt.Printf("  Used slots:        %d\n", result.UsedSlots)
	fmt.Printf("  Word count:        %d\n", result.WordCount)
	fmt.Printf("  Utilization:       %.2f%%\n", result.Utilization)

	if result.Utilization < 30 {
		fmt.Println("\nWarning: Space utilization is below 30%, consider rebuilding the trie")
	}
	return nil
}

type HelpCommand struct {
	commands []Command
}

func (c *HelpCommand) Name() string { return "help" }
func (c *HelpCommand) Usage() string {
	var buf strings.Builder
	buf.WriteString("Available commands:\n")
	for _, cmd := range c.commands {
		buf.WriteString(fmt.Sprintf("  %s\n", cmd.Usage()))
	}
	return buf.String()
}
func (c *HelpCommand) Run(args []string) error {
	fmt.Println("Double-Array Trie Client")
	fmt.Println()
	fmt.Println(c.Usage())
	return nil
}

func getServerURL(cmdArgs []string) string {
	url := "http://localhost:8403"
	if envURL := os.Getenv("TRIE_SERVER_URL"); envURL != "" {
		url = envURL
	}
	flagSet := flag.NewFlagSet("client", flag.ContinueOnError)
	flagURL := flagSet.String("server", "", "server URL")
	flagSet.Parse(cmdArgs)
	if *flagURL != "" {
		url = *flagURL
	}
	return url
}

func findCommandIndex(args []string) int {
	commands := map[string]bool{
		"add":    true,
		"search": true,
		"prefix": true,
		"remove": true,
		"import": true,
		"stat":   true,
	}
	for i, arg := range args {
		if commands[arg] {
			return i
		}
	}
	return -1
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: trie-client [options] <command> [arguments]")
		fmt.Println()
		commands := []Command{
			&AddCommand{},
			&SearchCommand{},
			&PrefixCommand{},
			&RemoveCommand{},
			&ImportCommand{},
			&StatCommand{},
		}
		help := &HelpCommand{commands: commands}
		help.Run(nil)
		os.Exit(1)
	}

	cmdIndex := findCommandIndex(os.Args[1:])
	if cmdIndex == -1 {
		fmt.Fprintf(os.Stderr, "Unknown or missing command\n")
		os.Exit(1)
	}

	flagArgs := os.Args[1 : cmdIndex+1]
	cmdName := os.Args[cmdIndex+1]
	args := os.Args[cmdIndex+2:]

	serverURL = getServerURL(flagArgs)

	commands := []Command{
		&AddCommand{},
		&SearchCommand{},
		&PrefixCommand{},
		&RemoveCommand{},
		&ImportCommand{},
		&StatCommand{},
	}

	for _, cmd := range commands {
		if cmd.Name() == cmdName {
			if err := cmd.Run(args); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmdName)
	os.Exit(1)
}
