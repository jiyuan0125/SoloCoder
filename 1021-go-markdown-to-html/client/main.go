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
	"md2html/common"
)

func main() {
	var serverURL string
	var inputFile string
	var outputFile string
	var markdown string
	
	flag.StringVar(&serverURL, "server", "http://localhost:8301", "Server URL (default: http://localhost:8301)")
	flag.StringVar(&inputFile, "input", "", "Input Markdown file (use '-' for stdin)")
	flag.StringVar(&outputFile, "output", "", "Output HTML file (default: stdout)")
	flag.StringVar(&markdown, "text", "", "Markdown text to convert")
	flag.Parse()
	
	var mdContent string
	var err error
	
	if markdown != "" {
		mdContent = markdown
	} else if inputFile != "" {
		if inputFile == "-" {
			content, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
				os.Exit(1)
			}
			mdContent = string(content)
		} else {
			content, err := ioutil.ReadFile(inputFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
				os.Exit(1)
			}
			mdContent = string(content)
		}
	} else {
		flag.Usage()
		os.Exit(1)
	}
	
	req := common.ConvertRequest{
		Markdown: mdContent,
	}
	
	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding request: %v\n", err)
		os.Exit(1)
	}
	
	resp, err := http.Post(
		strings.TrimRight(serverURL, "/")+"/convert",
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}
	
	var response common.ConvertResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding response: %v\n", err)
		fmt.Fprintf(os.Stderr, "Response body: %s\n", string(body))
		os.Exit(1)
	}
	
	if !response.Success {
		fmt.Fprintf(os.Stderr, "Server error: %s\n", response.Error)
		os.Exit(1)
	}
	
	if outputFile != "" {
		if err := ioutil.WriteFile(outputFile, []byte(response.HTML), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Print(response.HTML)
	}
}
