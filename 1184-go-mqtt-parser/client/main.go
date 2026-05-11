package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/mqtt/parser/api"
)

var serverURL string

func main() {
	flag.StringVar(&serverURL, "server", "http://localhost:8080", "Server URL")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	cmd := args[0]
	switch cmd {
	case "parse":
		handleParse(args[1:])
	case "build":
		handleBuild(args[1:])
	case "will":
		handleWill(args[1:])
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`MQTT Parser CLI

Usage:
  mqtt-cli [flags] <command> [arguments]

Commands:
  parse <hex|base64> <data>  Parse MQTT packet from hex or base64 encoded data
  build <type> [json-file]   Build MQTT packet (type: CONNECT, CONNACK, PUBLISH, etc.)
  will get                   Get current will configuration
  will put <json-file>       Update will configuration

Flags:
  -server string   Server URL (default "http://localhost:8080")

Examples:
  mqtt-cli parse hex 100c00044d5154540402003c0000
  mqtt-cli parse base64 "EAwABE1RVFQEAgA8AAA="
  mqtt-cli build CONNACK connect.json
  mqtt-cli will get
  mqtt-cli will put will-config.json
`)
}

func handleParse(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: mqtt-cli parse <hex|base64> <data>")
		os.Exit(1)
	}

	format := strings.ToLower(args[0])
	data := args[1]

	req := api.ParseRequest{
		Format: format,
		Data:   data,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/parse", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	var result api.ParseResponse
	json.Unmarshal(respBody, &result)

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Error)
		os.Exit(1)
	}

	pretty, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(pretty))
}

func handleBuild(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: mqtt-cli build <type> [json-file]")
		os.Exit(1)
	}

	packetType := strings.ToUpper(args[0])
	var fields map[string]interface{}

	if len(args) >= 2 {
		fileContent, err := ioutil.ReadFile(args[1])
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}
		if err := json.Unmarshal(fileContent, &fields); err != nil {
			fmt.Printf("Error parsing JSON: %v\n", err)
			os.Exit(1)
		}
	} else {
		fields = map[string]interface{}{}
	}

	req := api.BuildRequest{
		PacketType: packetType,
		Fields:     fields,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/build", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	var result api.BuildResponse
	json.Unmarshal(respBody, &result)

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Error)
		os.Exit(1)
	}

	pretty, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(pretty))
}

func handleWill(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: mqtt-cli will <get|put> [json-file]")
		os.Exit(1)
	}

	switch args[0] {
	case "get":
		handleWillGet()
	case "put":
		if len(args) < 2 {
			fmt.Println("Usage: mqtt-cli will put <json-file>")
			os.Exit(1)
		}
		handleWillPut(args[1])
	default:
		fmt.Println("Usage: mqtt-cli will <get|put> [json-file]")
		os.Exit(1)
	}
}

func handleWillGet() {
	resp, err := http.Get(serverURL + "/will")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	var result api.WillConfigResponse
	json.Unmarshal(respBody, &result)

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Error)
		os.Exit(1)
	}

	pretty, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(pretty))
}

func handleWillPut(filename string) {
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	var cfg api.WillConfig
	if err := json.Unmarshal(fileContent, &cfg); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	body, _ := json.Marshal(cfg)
	req, _ := http.NewRequest("PUT", serverURL+"/will", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	var result api.WillConfigResponse
	json.Unmarshal(respBody, &result)

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Error)
		os.Exit(1)
	}

	pretty, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(pretty))
}
