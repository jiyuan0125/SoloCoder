package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"stun-tool/pkg/common"
)

var (
	serverAddr string
	serverHost string
	serverPort int
	username   string
	password   string
	mode       string
)

func main() {
	flag.StringVar(&serverAddr, "server", "http://localhost:8080", "STUN HTTP server address")
	flag.StringVar(&serverHost, "host", "", "STUN server host (default: stun.l.google.com)")
	flag.IntVar(&serverPort, "port", 0, "STUN server port (default: 19302)")
	flag.StringVar(&username, "username", "", "STUN username (for authentication)")
	flag.StringVar(&password, "password", "", "STUN password (for authentication)")
	flag.StringVar(&mode, "mode", "detect", "Operation mode: detect or binding")
	flag.Parse()

	var err error
	switch mode {
	case "detect":
		err = runDetect()
	case "binding":
		err = runBinding()
	default:
		err = fmt.Errorf("invalid mode: %s (use 'detect' or 'binding')", mode)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runDetect() error {
	req := common.DetectRequest{
		ServerHost: serverHost,
		ServerPort: serverPort,
		Username:   username,
		Password:   password,
	}

	resp, err := sendHTTPRequest(serverAddr+"/api/detect", req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	var result common.DetectResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	printDetectResult(&result)
	return nil
}

func runBinding() error {
	if serverHost == "" {
		return fmt.Errorf("binding mode requires -host parameter")
	}

	req := common.BindingRequest{
		ServerHost: serverHost,
		ServerPort: serverPort,
		Username:   username,
		Password:   password,
	}

	resp, err := sendHTTPRequest(serverAddr+"/api/binding", req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	var result common.BindingResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	printBindingResult(&result)
	return nil
}

func sendHTTPRequest(url string, req interface{}) (*http.Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	return client.Do(httpReq)
}

func printDetectResult(result *common.DetectResponse) {
	fmt.Println("=== NAT Detection Result ===")
	if !result.Success {
		fmt.Printf("Status: Failed\n")
		if result.NATType != "" {
			fmt.Printf("NAT Type: %s\n", result.NATType)
		}
		if result.Error != "" {
			fmt.Printf("Error: %s\n", result.Error)
		}
		return
	}

	fmt.Printf("Status: Success\n")
	fmt.Printf("NAT Type: %s\n", result.NATType)
	if result.LocalIP != "" {
		fmt.Printf("Local IP: %s\n", result.LocalIP)
	}
	if result.PublicIP != "" {
		fmt.Printf("Public IP: %s\n", result.PublicIP)
	}
	if result.PublicPort != 0 {
		fmt.Printf("Public Port: %d\n", result.PublicPort)
	}
}

func printBindingResult(result *common.BindingResponse) {
	fmt.Println("=== Binding Result ===")
	if !result.Success {
		fmt.Printf("Status: Failed\n")
		if result.Error != "" {
			fmt.Printf("Error: %s\n", result.Error)
		}
		return
	}

	fmt.Printf("Status: Success\n")
	fmt.Printf("Mapped IP: %s\n", result.IP)
	fmt.Printf("Mapped Port: %d\n", result.Port)
	family := "IPv4"
	if result.Family == 0x02 {
		family = "IPv6"
	}
	fmt.Printf("Family: %s\n", family)
}
