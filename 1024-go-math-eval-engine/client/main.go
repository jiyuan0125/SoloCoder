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

	"matheval/common"
)

type variableFlag map[string]interface{}

func (v *variableFlag) String() string {
	return fmt.Sprintf("%v", *v)
}

func (v *variableFlag) Set(value string) error {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid variable format, use key=value")
	}
	key := strings.TrimSpace(parts[0])
	valStr := strings.TrimSpace(parts[1])
	if valStr == "true" {
		(*v)[key] = true
	} else if valStr == "false" {
		(*v)[key] = false
	} else if strings.Contains(valStr, ".") {
		var f float64
		if _, err := fmt.Sscanf(valStr, "%f", &f); err == nil {
			(*v)[key] = f
		} else {
			(*v)[key] = valStr
		}
	} else {
		var i int64
		if _, err := fmt.Sscanf(valStr, "%d", &i); err == nil {
			(*v)[key] = i
		} else {
			(*v)[key] = valStr
		}
	}
	return nil
}

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	expression := flag.String("expr", "", "Expression to evaluate")
	variables := variableFlag{}
	flag.Var(&variables, "var", "Variable in format key=value (can be used multiple times)")
	flag.Parse()
	if *expression == "" {
		fmt.Println("Error: expression is required")
		fmt.Println("Usage: client -expr \"1 + 2 * 3\" [-var a=10] [-var b=20]")
		os.Exit(1)
	}
	req := common.EvaluateRequest{
		Expression: *expression,
		Variables:  variables,
	}
	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: failed to encode request: %v\n", err)
		os.Exit(1)
	}
	url := strings.TrimRight(*serverURL, "/") + "/evaluate"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error: failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: failed to read response: %v\n", err)
		os.Exit(1)
	}
	var result common.EvaluateResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Printf("Error: failed to parse response: %v\n", err)
		fmt.Printf("Response body: %s\n", string(respBody))
		os.Exit(1)
	}
	if result.Success {
		fmt.Printf("Result: %v\n", result.Value)
		fmt.Printf("Type: %s\n", result.Type)
	} else {
		fmt.Println("Error(s):")
		for _, e := range result.Errors {
			fmt.Printf("  - %s\n", e)
		}
		os.Exit(1)
	}
}
