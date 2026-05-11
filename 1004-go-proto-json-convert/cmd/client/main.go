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

	"protojson.example.com/protojson/internal/common"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}
	return &Client{baseURL: strings.TrimSuffix(baseURL, "/")}
}

func (c *Client) ProtoToJSON(protoName string, protoData []byte) (*common.ProtoToJSONResponse, error) {
	req := common.ProtoToJSONRequest{
		ProtoName: protoName,
		ProtoData: protoData,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.baseURL+"/api/v1/proto-to-json", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var result common.ProtoToJSONResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

func (c *Client) JSONToProto(protoName string, jsonData []byte) (*common.JSONToProtoResponse, error) {
	req := common.JSONToProtoRequest{
		ProtoName: protoName,
		JSONData:  jsonData,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.baseURL+"/api/v1/json-to-proto", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var result common.JSONToProtoResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

func printUsage() {
	fmt.Println("Protobuf JSON Converter Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client proto-to-json <proto-name> <json-input>")
	fmt.Println("  client json-to-proto <proto-name> <json-input>")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -server string")
	fmt.Println("        Server URL (default \"localhost:8101\")")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  client proto-to-json my.package.Message '{\"field\": 123}'")
	fmt.Println("  client json-to-proto my.package.Message '{\"field\": \"value\"}'")
}

func main() {
	serverURL := flag.String("server", "localhost:8101", "Server URL")
	flag.Parse()

	args := flag.Args()
	if len(args) < 3 {
		printUsage()
		os.Exit(1)
	}

	cmd := args[0]
	protoName := args[1]
	inputData := args[2]

	client := NewClient(*serverURL)

	switch cmd {
	case "proto-to-json":
		resp, err := client.ProtoToJSON(protoName, []byte(inputData))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if resp.Error != "" {
			fmt.Fprintf(os.Stderr, "Server Error: %s\n", resp.Error)
			os.Exit(1)
		}
		var pretty bytes.Buffer
		json.Indent(&pretty, resp.JSONData, "", "  ")
		fmt.Println(pretty.String())

	case "json-to-proto":
		resp, err := client.JSONToProto(protoName, []byte(inputData))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if resp.Error != "" {
			fmt.Fprintf(os.Stderr, "Server Error: %s\n", resp.Error)
			os.Exit(1)
		}
		var pretty bytes.Buffer
		json.Indent(&pretty, resp.ProtoData, "", "  ")
		fmt.Println(pretty.String())

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}
