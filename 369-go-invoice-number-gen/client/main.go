package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"invoice-number-gen/protocol"
)

func main() {
	serverAddr := flag.String("server", "http://localhost:8080", "Invoice server address")
	prefix := flag.String("prefix", "", "Invoice prefix (default: INV)")
	flag.Parse()

	req := protocol.GenerateRequest{
		Prefix: *prefix,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to marshal request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(*serverAddr+"/generate", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to read response: %v\n", err)
		os.Exit(1)
	}

	var genResp protocol.GenerateResponse
	if err := json.Unmarshal(respBody, &genResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse response: %v\nResponse body: %s\n", err, string(respBody))
		os.Exit(1)
	}

	if genResp.Error != "" {
		fmt.Fprintf(os.Stderr, "Error: %s\n", genResp.Error)
		os.Exit(1)
	}

	fmt.Println(genResp.InvoiceNumber)
}
