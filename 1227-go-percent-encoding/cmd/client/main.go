package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"percent-encoding/internal/api"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "encode":
		handleEncode()
	case "decode":
		handleDecode()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Percent-Encoding Client

Usage:
  client encode [options] <input>
  client decode [options] <input>
  client help

Commands:
  encode  Encode input using specified mode
  decode  Decode input using specified mode

Options:
  --mode <url|uri|form>  Encoding/decoding mode (required)
  --server <url>         Server URL (default: http://localhost:8208)
  --component <type>     URL component for encoding (path|query|fragment|all)
`)
}

func handleEncode() {
	fs := flag.NewFlagSet("encode", flag.ExitOnError)
	mode := fs.String("mode", "", "Encoding mode (url|uri|form)")
	server := fs.String("server", "http://localhost:8208", "Server URL")
	component := fs.String("component", "", "URL component (path|query|fragment|all)")
	fs.Parse(os.Args[2:])

	if *mode == "" {
		fmt.Fprintln(os.Stderr, "Error: --mode is required")
		fs.Usage()
		os.Exit(1)
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: input is required")
		fs.Usage()
		os.Exit(1)
	}

	input := fs.Arg(0)

	endpoint := fmt.Sprintf("%s/encode/%s", *server, *mode)
	reqBody := api.EncodeRequest{
		Mode:      *mode,
		Input:     input,
		Component: *component,
	}

	resp, err := sendRequest(endpoint, reqBody)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Println(resp.Output)
}

func handleDecode() {
	fs := flag.NewFlagSet("decode", flag.ExitOnError)
	mode := fs.String("mode", "", "Decoding mode (url|uri|form)")
	server := fs.String("server", "http://localhost:8208", "Server URL")
	fs.Parse(os.Args[2:])

	if *mode == "" {
		fmt.Fprintln(os.Stderr, "Error: --mode is required")
		fs.Usage()
		os.Exit(1)
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: input is required")
		fs.Usage()
		os.Exit(1)
	}

	input := fs.Arg(0)

	endpoint := fmt.Sprintf("%s/decode", *server)
	reqBody := api.DecodeRequest{
		Mode:  *mode,
		Input: input,
	}

	resp, err := sendDecodeRequest(endpoint, reqBody)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Println(resp.Output)
}

func sendRequest(endpoint string, reqBody api.EncodeRequest) (*api.EncodeResponse, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp api.EncodeResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	return &apiResp, nil
}

func sendDecodeRequest(endpoint string, reqBody api.DecodeRequest) (*api.DecodeResponse, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp api.DecodeResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	return &apiResp, nil
}
