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

	"httpcacheproxy/pkg/common"
)

var (
	proxyServer = flag.String("server", "http://localhost:8080", "Cache proxy server address")
	method      = flag.String("method", "GET", "HTTP method")
	headers     = flag.String("headers", "", "HTTP headers (format: Key1:Value1,Key2:Value2)")
	body        = flag.String("body", "", "Request body")
	showHeaders = flag.Bool("headers", false, "Show response headers")
	showStatus  = flag.Bool("status", false, "Show status code only")
	verbose     = flag.Bool("v", false, "Verbose output")
)

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: cacheclient [flags] <url>")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	targetURL := flag.Arg(0)

	req := common.ProxyRequest{
		Method: *method,
		URL:    targetURL,
		Header: make(http.Header),
		Body:   []byte(*body),
	}

	if *headers != "" {
		for _, h := range splitHeaders(*headers) {
			parts := strings.SplitN(h, ":", 2)
			if len(parts) == 2 {
				req.Header.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}
	}

	if *verbose {
		fmt.Printf("Sending request to: %s\n", targetURL)
		fmt.Printf("Method: %s\n", req.Method)
		if len(req.Header) > 0 {
			fmt.Println("Headers:")
			for k, v := range req.Header {
				fmt.Printf("  %s: %s\n", k, strings.Join(v, ", "))
			}
		}
	}

	resp, err := callProxy(*proxyServer, &req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *showStatus {
		fmt.Println(resp.StatusCode)
		return
	}

	if resp.FromCache {
		fmt.Printf("[Cache HIT] Age: %d seconds\n", resp.CacheAge)
	} else {
		fmt.Println("[Cache MISS]")
	}

	if *showHeaders || *verbose {
		fmt.Println("Status Code:", resp.StatusCode)
		fmt.Println("Headers:")
		for k, v := range resp.Header {
			fmt.Printf("  %s: %s\n", k, strings.Join(v, ", "))
		}
	}

	fmt.Println(string(resp.Body))
}

func callProxy(serverAddr string, req *common.ProxyRequest) (*common.ProxyResponse, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpResp, err := http.Post(
		serverAddr+"/proxy",
		"application/json",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call proxy: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("proxy returned error: %d - %s", httpResp.StatusCode, string(respBody))
	}

	var resp common.ProxyResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

func splitHeaders(headers string) []string {
	var result []string
	var current strings.Builder
	inQuote := false

	for i := 0; i < len(headers); i++ {
		c := headers[i]
		if c == '"' {
			inQuote = !inQuote
			continue
		}
		if c == ',' && !inQuote {
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
			continue
		}
		current.WriteByte(c)
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}
