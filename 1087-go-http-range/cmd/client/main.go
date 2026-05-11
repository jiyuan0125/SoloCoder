package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	url := flag.String("url", "http://localhost:8102", "Server URL")
	rangeHeader := flag.String("range", "", "Range header value (e.g., bytes=0-100)")
	ifRange := flag.String("if-range", "", "If-Range header value (ETag or last modified time)")
	flag.Parse()

	req, err := http.NewRequest("GET", *url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		os.Exit(1)
	}

	if *rangeHeader != "" {
		req.Header.Set("Range", *rangeHeader)
	}

	if *ifRange != "" {
		req.Header.Set("If-Range", *ifRange)
	}

	fmt.Printf("Sending request to: %s\n", *url)
	fmt.Printf("Headers:\n")
	for key, values := range req.Header {
		for _, value := range values {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("\nResponse Status: %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))
	fmt.Printf("Response Headers:\n")
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response body: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nResponse Body (%d bytes):\n", len(body))
	if strings.Contains(resp.Header.Get("Content-Type"), "text/") || 
	   strings.Contains(resp.Header.Get("Content-Type"), "application/json") ||
	   len(body) < 1000 {
		fmt.Printf("%s\n", string(body))
	} else {
		fmt.Printf("<binary data, length: %d bytes>\n", len(body))
	}
}
