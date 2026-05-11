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
	"strconv"
	"strings"

	"topk-frequent/pkg/api"
)

const defaultServer = "http://localhost:8216"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "create":
		cmdCreate(args)
	case "add":
		cmdAdd(args)
	case "topk":
		cmdTopK(args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  client create [--width W] [--depth D] [--k K] [--server URL]")
	fmt.Println("  client add [--file FILE] [--server URL]")
	fmt.Println("  client topk [--k K] [--server URL]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create  Create a new Top-K analyzer")
	fmt.Println("  add     Add elements from stdin or file (one per line)")
	fmt.Println("  topk    Query current Top-K results")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --width   Count-Min Sketch width (default: 1000)")
	fmt.Println("  --depth   Count-Min Sketch depth (default: 5)")
	fmt.Println("  --k       Number of top elements (default: 10)")
	fmt.Println("  --file    Read elements from file instead of stdin")
	fmt.Println("  --server  Server URL (default: http://localhost:8216)")
}

func getServerURL(args []string) string {
	for i, arg := range args {
		if arg == "--server" && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(arg, "--server=") {
			return strings.TrimPrefix(arg, "--server=")
		}
	}
	if envServer := os.Getenv("TOPK_SERVER"); envServer != "" {
		return envServer
	}
	return defaultServer
}

func cmdCreate(args []string) {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	width := fs.Int("width", 1000, "Sketch width")
	depth := fs.Int("depth", 5, "Sketch depth")
	k := fs.Int("k", 10, "Top-K value")
	server := fs.String("server", defaultServer, "Server URL")
	fs.Parse(args)

	if *server == defaultServer {
		*server = getServerURL(args)
	}

	req := api.CreateRequest{
		Width: *width,
		Depth: *depth,
		K:     *k,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(*server+"/create", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result api.CreateResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Success {
		fmt.Printf("Success: %s\n", result.Message)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", result.Message)
		os.Exit(1)
	}
}

func cmdAdd(args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	file := fs.String("file", "", "Input file")
	server := fs.String("server", defaultServer, "Server URL")
	fs.Parse(args)

	if *server == defaultServer {
		*server = getServerURL(args)
	}

	var reader io.Reader
	if *file != "" {
		f, err := os.Open(*file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		reader = f
	} else {
		reader = os.Stdin
	}

	items := make([]string, 0)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			items = append(items, line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	if len(items) == 0 {
		fmt.Println("No items to add")
		return
	}

	req := api.AddRequest{Items: items}
	body, _ := json.Marshal(req)
	resp, err := http.Post(*server+"/add", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result api.AddResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Success {
		fmt.Printf("Success: %s\n", result.Message)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", result.Message)
		os.Exit(1)
	}
}

func cmdTopK(args []string) {
	fs := flag.NewFlagSet("topk", flag.ExitOnError)
	k := fs.Int("k", 0, "Override K value")
	server := fs.String("server", defaultServer, "Server URL")
	fs.Parse(args)

	if *server == defaultServer {
		*server = getServerURL(args)
	}

	url := *server + "/topk"
	if *k > 0 {
		url += "?k=" + strconv.Itoa(*k)
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result api.TopKResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if !result.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", result.Message)
		os.Exit(1)
	}

	if len(result.Items) == 0 {
		fmt.Println("No items in Top-K")
		return
	}

	fmt.Printf("Top-%d items:\n", len(result.Items))
	fmt.Println(strings.Repeat("-", 50))
	for i, item := range result.Items {
		uncertain := ""
		if item.Uncertain {
			uncertain = " [uncertain]"
		}
		fmt.Printf("%3d. %-20s freq: %d%s\n", i+1, item.Element, item.Freq, uncertain)
	}
}
