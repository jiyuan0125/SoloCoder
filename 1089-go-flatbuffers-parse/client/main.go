package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/example/flatbuffers-parser/common"
)

func main() {
	var serverURL string
	var bufferFile string
	var schemaFile string
	var command string

	flag.StringVar(&serverURL, "server", "http://localhost:8500", "Server URL")
	flag.StringVar(&bufferFile, "buffer", "", "Path to FlatBuffers buffer file")
	flag.StringVar(&schemaFile, "schema", "", "Path to JSON schema file")
	flag.StringVar(&command, "command", "inspect", "Command: inspect | health")

	flag.Parse()

	if command == "health" {
		err := checkHealth(serverURL)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if command != "inspect" {
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}

	if bufferFile == "" || schemaFile == "" {
		fmt.Println("Error: -buffer and -schema are required for inspect command")
		flag.Usage()
		os.Exit(1)
	}

	bufferData, err := ioutil.ReadFile(bufferFile)
	if err != nil {
		fmt.Printf("Error reading buffer file: %v\n", err)
		os.Exit(1)
	}

	schemaData, err := ioutil.ReadFile(schemaFile)
	if err != nil {
		fmt.Printf("Error reading schema file: %v\n", err)
		os.Exit(1)
	}

	var schema common.SchemaDefinition
	if err := json.Unmarshal(schemaData, &schema); err != nil {
		fmt.Printf("Error parsing schema JSON: %v\n", err)
		os.Exit(1)
	}

	req := common.InspectRequest{
		Buffer: common.EncodeBuffer(bufferData),
		Schema: &schema,
	}

	resp, err := sendInspectRequest(serverURL, &req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Server error: %s\n", resp.Message)
		os.Exit(1)
	}

	resultJSON, _ := json.MarshalIndent(resp.Result, "", "  ")
	fmt.Println(string(resultJSON))
}

func checkHealth(serverURL string) error {
	resp, err := http.Get(serverURL + "/health")
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	return nil
}

func sendInspectRequest(serverURL string, req *common.InspectRequest) (*common.InspectResponse, error) {
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	httpResp, err := http.Post(
		serverURL+"/inspect",
		"application/json",
		bytes.NewReader(reqData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d: %s", httpResp.StatusCode, string(body))
	}

	var resp common.InspectResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &resp, nil
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  Check server health:\n")
		fmt.Fprintf(os.Stderr, "    %s -command health\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  Inspect a FlatBuffers buffer:\n")
		fmt.Fprintf(os.Stderr, "    %s -buffer data.bin -schema schema.json\n", os.Args[0])
	}
}
