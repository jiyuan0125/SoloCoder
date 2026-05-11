package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/fuzzysearch/client"
)

const (
	defaultServerURL = "http://localhost:8303"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := os.Getenv("FUZZY_SERVER_URL")
	if serverURL == "" {
		serverURL = defaultServerURL
	}

	fs := flag.NewFlagSet("global", flag.ExitOnError)
	fs.StringVar(&serverURL, "server", serverURL, "Server URL")
	fs.StringVar(&serverURL, "s", serverURL, "Short alias for -server")

	command := os.Args[1]
	args := os.Args[2:]

	c := client.NewFuzzyClient(serverURL)

	var err error
	switch command {
	case "add":
		err = handleAdd(c, args)
	case "search":
		err = handleSearch(c, args)
	case "wildcard":
		err = handleWildcard(c, args)
	case "dict":
		err = handleDict(c, args)
	case "import":
		err = handleImport(c, args)
	case "help", "-h", "--help":
		printUsage()
		return
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleAdd(c *client.FuzzyClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: fuzzy add <word>")
	}
	return c.AddWord(args[0])
}

func handleSearch(c *client.FuzzyClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: fuzzy search <query> [threshold]")
	}

	query := args[0]
	threshold := 1

	if len(args) >= 2 {
		var err error
		threshold, err = strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid threshold: %s", args[1])
		}
		if threshold < 0 {
			return fmt.Errorf("threshold must be non-negative")
		}
	}

	return c.Search(query, threshold)
}

func handleWildcard(c *client.FuzzyClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: fuzzy wildcard <pattern> [threshold]")
	}

	pattern := args[0]
	threshold := 0

	if len(args) >= 2 {
		var err error
		threshold, err = strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid threshold: %s", args[1])
		}
		if threshold < 0 {
			return fmt.Errorf("threshold must be non-negative")
		}
	}

	return c.WildcardSearch(pattern, threshold)
}

func handleDict(c *client.FuzzyClient, args []string) error {
	return c.GetDictionary()
}

func handleImport(c *client.FuzzyClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: fuzzy import <file_path>")
	}
	return c.ImportWords(args[0])
}

func printUsage() {
	fmt.Println("Fuzzy Search Client")
	fmt.Println("Usage:")
	fmt.Println("  fuzzy add <word>              - Add a word to the dictionary")
	fmt.Println("  fuzzy search <query> [k]      - Search for words within edit distance k (default: 1)")
	fmt.Println("  fuzzy wildcard <pattern> [k]  - Search with wildcards (?, *) and threshold k (default: 0)")
	fmt.Println("  fuzzy dict                    - List all words in the dictionary")
	fmt.Println("  fuzzy import <file>           - Import words from a file (one per line)")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -s, --server <url>            - Server URL (default: http://localhost:8303 or FUZZY_SERVER_URL env)")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  fuzzy add hello")
	fmt.Println("  fuzzy search helo 2")
	fmt.Println("  fuzzy wildcard h?llo 0")
	fmt.Println("  fuzzy wildcard h*llo 0")
	fmt.Println("  fuzzy dict")
}
