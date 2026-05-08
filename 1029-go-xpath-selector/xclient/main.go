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

	"github.com/example/xpath-selector/xproto"
)

type Client struct {
	serverURL string
}

func NewClient(serverURL string) *Client {
	if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
		serverURL = "http://" + serverURL
	}
	if !strings.HasSuffix(serverURL, "/query") {
		if !strings.HasSuffix(serverURL, "/") {
			serverURL += "/"
		}
		serverURL += "query"
	}
	return &Client{serverURL: serverURL}
}

func (c *Client) Query(req xproto.QueryRequest) (*xproto.QueryResult, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}
	
	httpReq, err := http.NewRequest("POST", c.serverURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}
	
	var result xproto.QueryResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}
	
	return &result, nil
}

func printUsage() {
	fmt.Println("XPath Query Client - Command Line Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  xclient [options] -x <xpath> <xml-file>")
	fmt.Println("  xclient [options] -x <xpath> - < (read XML from stdin)")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  xclient -x //book/title books.xml")
	fmt.Println("  xclient -x \"book[price>35]\" - < books.xml")
	fmt.Println("  xclient -x //@id -s http://server:8080 data.xml")
}

func main() {
	flag.Usage = printUsage
	
	xpath := flag.String("x", "", "XPath expression (required)")
	server := flag.String("s", "http://localhost:8080", "Server URL")
	returnType := flag.String("t", "", "Return type: nodes|string|number|boolean")
	nsFlag := flag.String("n", "", "Namespace mappings: prefix1=uri1,prefix2=uri2")
	pretty := flag.Bool("p", false, "Pretty print output")
	help := flag.Bool("h", false, "Show help")
	
	flag.Parse()
	
	if *help {
		printUsage()
		os.Exit(0)
	}
	
	if *xpath == "" {
		fmt.Fprintln(os.Stderr, "Error: XPath expression is required (-x)")
		printUsage()
		os.Exit(1)
	}
	
	args := flag.Args()
	
	var xmlContent string
	
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: XML file path or '-' for stdin is required")
		printUsage()
		os.Exit(1)
	}
	
	if args[0] == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
		}
		xmlContent = string(data)
	} else {
		data, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
		xmlContent = string(data)
	}
	
	var namespaces []xproto.NamespaceMapping
	if *nsFlag != "" {
		parts := strings.Split(*nsFlag, ",")
		for _, part := range parts {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				namespaces = append(namespaces, xproto.NamespaceMapping{
					Prefix: strings.TrimSpace(kv[0]),
					URI:    strings.TrimSpace(kv[1]),
				})
			}
		}
	}
	
	req := xproto.QueryRequest{
		XML:        xmlContent,
		XPath:      *xpath,
		Namespaces: namespaces,
		ReturnType: *returnType,
	}
	
	client := NewClient(*server)
	result, err := client.Query(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		os.Exit(1)
	}
	
	if !result.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", result.Error)
		os.Exit(1)
	}
	
	var output []byte
	if *pretty {
		output, err = json.MarshalIndent(result, "", "  ")
	} else {
		output, err = json.Marshal(result)
	}
	
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println(string(output))
}
