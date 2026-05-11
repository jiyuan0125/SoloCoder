package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"der-ber-parser/api"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "encode":
		handleEncode(args)
	case "decode":
		handleDecode(args)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`DER/BER ASN.1 Parser Client

Usage:
  client encode [options] <json_data>
  client decode [options] <hex_string>

Commands:
  encode  Encode JSON data to DER or BER
  decode  Decode hex string to JSON

Encode Options:
  --mode string   Encoding mode: der or ber (default: der)
  --server string Server address (default: http://localhost:8320)

Decode Options:
  --server string Server address (default: http://localhost:8320)

Examples:
  client encode --mode der '{"type":"SEQUENCE","children":[{"type":"INTEGER","value":123}]}'
  client encode --mode ber '{"type":"SEQUENCE","children":[{"type":"UTF8String","value":"hello"}]}'
  client decode '300302017B'
`)
}

func handleEncode(args []string) {
	fs := flag.NewFlagSet("encode", flag.ExitOnError)
	mode := fs.String("mode", "der", "encoding mode: der or ber")
	server := fs.String("server", "http://localhost:8320", "server address")
	fs.Parse(args)

	remaining := fs.Args()
	if len(remaining) == 0 {
		fmt.Fprintln(os.Stderr, "Error: missing JSON data argument")
		os.Exit(1)
	}

	modeLower := strings.ToLower(*mode)
	if modeLower != "der" && modeLower != "ber" {
		fmt.Fprintf(os.Stderr, "Error: invalid mode '%s', must be 'der' or 'ber'\n", *mode)
		os.Exit(1)
	}

	var data map[string]interface{}
	jsonData := strings.Join(remaining, " ")
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid JSON: %v\n", err)
		os.Exit(1)
	}

	var jt api.EncodeRequest
	if err := json.Unmarshal([]byte(jsonData), &jt.Data); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid JSON data: %v\n", err)
		os.Exit(1)
	}
	jt.Mode = modeLower

	var endpoint string
	if modeLower == "der" {
		endpoint = *server + "/encode-der"
	} else {
		endpoint = *server + "/encode-ber"
	}

	reqBody, err := json.Marshal(jt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to marshal request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to read response: %v\n", err)
		os.Exit(1)
	}

	var result api.EncodeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse response: %v\n", err)
		fmt.Fprintf(os.Stderr, "Response body: %s\n", string(body))
		os.Exit(1)
	}

	if !result.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", result.Error)
		os.Exit(1)
	}

	fmt.Println(result.Hex)
}

func handleDecode(args []string) {
	fs := flag.NewFlagSet("decode", flag.ExitOnError)
	server := fs.String("server", "http://localhost:8320", "server address")
	fs.Parse(args)

	remaining := fs.Args()
	if len(remaining) == 0 {
		fmt.Fprintln(os.Stderr, "Error: missing hex string argument")
		os.Exit(1)
	}

	hexStr := strings.Join(remaining, "")
	hexStr = strings.ReplaceAll(hexStr, " ", "")

	req := api.DecodeRequest{Hex: hexStr}
	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to marshal request: %v\n", err)
		os.Exit(1)
	}

	endpoint := *server + "/decode"
	resp, err := http.Post(endpoint, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to read response: %v\n", err)
		os.Exit(1)
	}

	var result api.DecodeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse response: %v\n", err)
		fmt.Fprintf(os.Stderr, "Response body: %s\n", string(body))
		os.Exit(1)
	}

	if !result.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", result.Error)
		os.Exit(1)
	}

	output, err := json.MarshalIndent(result.Data, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to format output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}
