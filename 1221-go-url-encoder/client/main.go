package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"urlencoder/common"
)

func getServerURL() string {
	if envURL := os.Getenv("SERVER_URL"); envURL != "" {
		return strings.TrimRight(envURL, "/")
	}
	return "http://localhost:8202"
}

func postJSON(url string, payload interface{}, result interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	return nil
}

func cmdEncode(args []string) {
	fs := flag.NewFlagSet("encode", flag.ExitOnError)
	serverURL := fs.String("server", getServerURL(), "server URL")
	fs.Parse(args)

	if fs.NArg() != 1 {
		fmt.Println("usage: urlencoder encode <text>")
		os.Exit(1)
	}

	text := fs.Arg(0)
	server := strings.TrimRight(*serverURL, "/")

	var resp common.EncodeResponse
	if err := postJSON(server+"/encode", common.EncodeRequest{Text: text}, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if resp.Error != "" {
		fmt.Fprintf(os.Stderr, "error: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Println(resp.Result)
}

func cmdDecode(args []string) {
	fs := flag.NewFlagSet("decode", flag.ExitOnError)
	serverURL := fs.String("server", getServerURL(), "server URL")
	fs.Parse(args)

	if fs.NArg() != 1 {
		fmt.Println("usage: urlencoder decode <text>")
		os.Exit(1)
	}

	text := fs.Arg(0)
	server := strings.TrimRight(*serverURL, "/")

	var resp common.DecodeResponse
	if err := postJSON(server+"/decode", common.DecodeRequest{Text: text}, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if resp.Error != "" {
		fmt.Fprintf(os.Stderr, "error: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Println(resp.Result)
}

func cmdBatch(args []string) {
	fs := flag.NewFlagSet("batch", flag.ExitOnError)
	serverURL := fs.String("server", getServerURL(), "server URL")
	operation := fs.String("op", "", "operation: encode or decode (required)")
	fs.Parse(args)

	if *operation == "" {
		fmt.Println("usage: urlencoder batch -op <encode|decode> <text1> <text2> ...")
		os.Exit(1)
	}

	if fs.NArg() < 1 {
		fmt.Println("usage: urlencoder batch -op <encode|decode> <text1> <text2> ...")
		os.Exit(1)
	}

	op := strings.ToLower(strings.TrimSpace(*operation))
	if op != "encode" && op != "decode" {
		fmt.Fprintf(os.Stderr, "error: operation must be 'encode' or 'decode'\n")
		os.Exit(1)
	}

	texts := fs.Args()
	server := strings.TrimRight(*serverURL, "/")

	var resp common.BatchResponse
	if err := postJSON(server+"/batch", common.BatchRequest{Operation: op, Texts: texts}, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	hasError := false
	for i, errStr := range resp.Errors {
		if errStr != "" {
			fmt.Fprintf(os.Stderr, "item %d error: %s\n", i, errStr)
			hasError = true
		}
	}

	for i, result := range resp.Results {
		fmt.Printf("%d: %s\n", i, result)
	}

	if hasError {
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("URL Encoder/Decoder (RFC 3986)")
	fmt.Println()
	fmt.Println("usage: urlencoder <command> [options] <arguments>")
	fmt.Println()
	fmt.Println("commands:")
	fmt.Println("  encode  <text>           encode text")
	fmt.Println("  decode  <text>           decode text")
	fmt.Println("  batch   -op <op> <texts> batch encode or decode")
	fmt.Println()
	fmt.Println("options:")
	fmt.Println("  -server <url>            server URL (default: http://localhost:8202)")
	fmt.Println("  -op     <encode|decode>  batch operation (required for batch)")
	fmt.Println()
	fmt.Println("environment:")
	fmt.Println("  SERVER_URL               default server URL")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := strings.ToLower(os.Args[1])
	args := os.Args[2:]

	switch cmd {
	case "encode":
		cmdEncode(args)
	case "decode":
		cmdDecode(args)
	case "batch":
		cmdBatch(args)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "error: unknown command '%s'\n", cmd)
		printUsage()
		os.Exit(1)
	}
}
