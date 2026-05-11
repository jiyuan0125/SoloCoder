package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"fractional-service/api"
)

const (
	defaultServerURL = "http://localhost:8080"
)

func main() {
	serverURL := flag.String("server", defaultServerURL, "Server URL")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	operation := api.Operation(args[0])

	var req *api.OperationRequest
	var err error

	switch operation {
	case api.OperationDecimal:
		req, err = buildDecimalRequest(args[1:])
	case api.OperationAdd, api.OperationSub, api.OperationMul, api.OperationDiv:
		req, err = buildBinaryRequest(operation, args[1:])
	default:
		fmt.Printf("Unknown operation: %s\n", operation)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		printUsage()
		os.Exit(1)
	}

	resp, err := sendRequest(*serverURL, req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}

	printResponse(resp)
}

func buildBinaryRequest(operation api.Operation, args []string) (*api.OperationRequest, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("%s requires exactly 2 operands", operation)
	}

	return &api.OperationRequest{
		Operation: operation,
		Operand1:  args[0],
		Operand2:  args[1],
	}, nil
}

func buildDecimalRequest(args []string) (*api.OperationRequest, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("decimal requires at least 1 operand")
	}

	precision := 10
	if len(args) >= 2 {
		p, err := strconv.Atoi(args[1])
		if err != nil {
			return nil, fmt.Errorf("invalid precision: %v", err)
		}
		precision = p
	}

	return &api.OperationRequest{
		Operation: api.OperationDecimal,
		Operand1:  args[0],
		Precision: precision,
	}, nil
}

func sendRequest(serverURL string, req *api.OperationRequest) (*api.OperationResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, serverURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer httpResp.Body.Close()

	respBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var resp api.OperationResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return &resp, nil
}

func printResponse(resp *api.OperationResponse) {
	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		return
	}

	result := resp.Result
	if result == nil {
		fmt.Println("No result")
		return
	}

	if result.Integer != "" {
		fmt.Println(result.Integer)
	} else {
		fmt.Println(result.Fraction)
	}

	if result.Mixed != "" && result.Mixed != result.Fraction && result.Mixed != result.Integer {
		fmt.Printf("Mixed: %s\n", result.Mixed)
	}

	if result.Decimal != "" {
		fmt.Printf("Decimal: %s\n", result.Decimal)
	}
}

func printUsage() {
	fmt.Println("Usage: frac <operation> <arguments>")
	fmt.Println()
	fmt.Println("Operations:")
	fmt.Println("  add <operand1> <operand2>   Add two fractions")
	fmt.Println("  sub <operand1> <operand2>   Subtract two fractions")
	fmt.Println("  mul <operand1> <operand2>   Multiply two fractions")
	fmt.Println("  div <operand1> <operand2>   Divide two fractions")
	fmt.Println("  decimal <operand> [precision]  Convert to decimal")
	fmt.Println()
	fmt.Println("Operand formats:")
	fmt.Println("  Integer: 3, -5")
	fmt.Println("  Decimal: 0.5, 0.333")
	fmt.Println("  Fraction: 3/4, -2/7")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  frac add \"1/3\" \"1/6\"")
	fmt.Println("  frac mul \"2/3\" \"3/4\"")
	fmt.Println("  frac decimal \"1/3\" 15")
}
