package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"bigcalc/common"
)

const serverURL = "http://localhost:8080/calc"

func main() {
	if len(os.Args) < 4 {
		printUsage()
		os.Exit(1)
	}

	opStr := os.Args[1]
	a := os.Args[2]
	b := os.Args[3]

	op, err := parseOperation(opStr)
	if err != nil {
		fmt.Println("Error:", err)
		printUsage()
		os.Exit(1)
	}

	req := common.Request{
		Operation: op,
		A:         a,
		B:         b,
	}

	resp, err := sendRequest(req)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Println("Error:", resp.Error)
		os.Exit(1)
	}

	printResponse(op, resp)
}

func printUsage() {
	fmt.Println("Usage: bigcalc <operation> <a> <b>")
	fmt.Println("Operations: add, subtract, multiply, divide")
	fmt.Println("Examples:")
	fmt.Println("  bigcalc add \"99999999999999999999\" \"1\"")
	fmt.Println("  bigcalc subtract \"5\" \"10\"")
	fmt.Println("  bigcalc multiply \"12345678901234567890\" \"98765432109876543210\"")
	fmt.Println("  bigcalc divide \"10\" \"3\"")
}

func parseOperation(op string) (common.Operation, error) {
	switch op {
	case "add":
		return common.OpAdd, nil
	case "subtract", "sub":
		return common.OpSubtract, nil
	case "multiply", "mul":
		return common.OpMultiply, nil
	case "divide", "div":
		return common.OpDivide, nil
	default:
		return "", fmt.Errorf("unknown operation: %s", op)
	}
}

func sendRequest(req common.Request) (*common.Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpResp, err := http.Post(serverURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	var resp common.Response
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func printResponse(op common.Operation, resp *common.Response) {
	switch op {
	case common.OpAdd:
		fmt.Println(resp.Result)
	case common.OpSubtract:
		fmt.Println(resp.Result)
	case common.OpMultiply:
		fmt.Println(resp.Result)
	case common.OpDivide:
		fmt.Printf("quotient: %s\n", resp.Quotient)
		fmt.Printf("remainder: %s\n", resp.Remainder)
	}
}
