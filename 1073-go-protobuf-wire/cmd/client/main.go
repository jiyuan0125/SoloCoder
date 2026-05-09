package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"protobuf-wire/internal/common"
)

func main() {
	var serverURL string
	var filePath string

	flag.StringVar(&serverURL, "server", "http://localhost:8080", "protobuf wire parser server URL")
	flag.StringVar(&filePath, "file", "", "path to protobuf binary file")
	flag.Parse()

	if filePath == "" {
		fmt.Fprintln(os.Stderr, "error: -file flag is required")
		flag.Usage()
		os.Exit(1)
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
		os.Exit(1)
	}

	reqBody, err := json.Marshal(common.ParseRequest{Data: data})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error encoding request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/parse", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading response: %v\n", err)
		os.Exit(1)
	}

	var parsedResp common.ParseResponse
	if err := json.Unmarshal(respBody, &parsedResp); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing response: %v\n", err)
		fmt.Fprintln(os.Stderr, "raw response:")
		fmt.Fprintln(os.Stderr, string(respBody))
		os.Exit(1)
	}

	formatted, err := json.MarshalIndent(parsedResp, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error formatting response: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(formatted))

	if !parsedResp.Success {
		os.Exit(1)
	}
}
