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

	"bencode-parser/api"
)

var serverURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "decode":
		handleDecode(args)
	case "encode":
		handleEncode(args)
	case "info":
		handleInfo(args)
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("BEncoding Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client decode <bencode-data>        Decode BEncoding data")
	fmt.Println("  client decode -f <file>             Decode BEncoding data from file")
	fmt.Println("  client encode <json-data>           Encode JSON data to BEncoding")
	fmt.Println("  client encode -f <file>             Encode JSON data from file to BEncoding")
	fmt.Println("  client info <bencode-data>          Get info about BEncoding data")
	fmt.Println("  client info -f <file>               Get info about BEncoding data from file")
	fmt.Println("  client help                         Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  client decode 'd3:foo3:bar4:spam4:eggse'")
	fmt.Println("  client encode '{\"foo\":\"bar\",\"spam\":\"eggs\"}'")
	fmt.Println("  client info 'li1ei2ei3ee'")
}

func handleDecode(args []string) {
	fs := flag.NewFlagSet("decode", flag.ExitOnError)
	filePath := fs.String("f", "", "Read from file")
	fs.Parse(args)

	var data string
	if *filePath != "" {
		content, err := ioutil.ReadFile(*filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
		data = string(content)
	} else if len(fs.Args()) > 0 {
		data = strings.Join(fs.Args(), " ")
	} else {
		fmt.Fprintf(os.Stderr, "Error: no input provided\n")
		os.Exit(1)
	}

	req := api.DecodeRequest{Data: data}
	reqBody, _ := json.Marshal(req)

	resp, err := http.Post(serverURL+"/decode", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result api.DecodeResponse
	json.Unmarshal(body, &result)

	if !result.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", result.Error)
		os.Exit(1)
	}

	output, _ := json.MarshalIndent(result.Result, "", "  ")
	fmt.Println(string(output))
}

func handleEncode(args []string) {
	fs := flag.NewFlagSet("encode", flag.ExitOnError)
	filePath := fs.String("f", "", "Read from file")
	fs.Parse(args)

	var data string
	if *filePath != "" {
		content, err := ioutil.ReadFile(*filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
		data = string(content)
	} else if len(fs.Args()) > 0 {
		data = strings.Join(fs.Args(), " ")
	} else {
		fmt.Fprintf(os.Stderr, "Error: no input provided\n")
		os.Exit(1)
	}

	var jsonData interface{}
	if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	req := api.EncodeRequest{Data: jsonData}
	reqBody, _ := json.Marshal(req)

	resp, err := http.Post(serverURL+"/encode", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result api.EncodeResponse
	json.Unmarshal(body, &result)

	if !result.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", result.Error)
		os.Exit(1)
	}

	fmt.Println(result.Result)
}

func handleInfo(args []string) {
	fs := flag.NewFlagSet("info", flag.ExitOnError)
	filePath := fs.String("f", "", "Read from file")
	fs.Parse(args)

	var data string
	if *filePath != "" {
		content, err := ioutil.ReadFile(*filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
		data = string(content)
	} else if len(fs.Args()) > 0 {
		data = strings.Join(fs.Args(), " ")
	} else {
		fmt.Fprintf(os.Stderr, "Error: no input provided\n")
		os.Exit(1)
	}

	req := api.InfoRequest{Data: data}
	reqBody, _ := json.Marshal(req)

	resp, err := http.Post(serverURL+"/info", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result api.InfoResponse
	json.Unmarshal(body, &result)

	if !result.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", result.Error)
		os.Exit(1)
	}

	fmt.Printf("Outer Type: %s\n", result.OuterType)
	if result.OuterType == "dictionary" {
		fmt.Printf("Dictionary Key Count: %d\n", result.DictKeyCount)
	}
	if result.OuterType == "list" {
		fmt.Printf("List Element Count: %d\n", result.ListElemCount)
	}
	fmt.Printf("Max Nesting Depth: %d\n", result.MaxDepth)
}
