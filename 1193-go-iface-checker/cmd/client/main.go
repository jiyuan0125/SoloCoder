package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"iface-checker/pkg/models"
	"iface-checker/pkg/parser"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	var mode string
	var serverURL string
	var input string
	var structName string
	var interfaceName string
	var useValueReceiver bool

	flag.StringVar(&mode, "mode", "file", "operation mode: file or inline")
	flag.StringVar(&serverURL, "server", "http://localhost:8080", "server URL")
	flag.StringVar(&input, "input", "", "input: file path or JSON string")
	flag.StringVar(&structName, "struct", "", "struct name to check")
	flag.StringVar(&interfaceName, "interface", "", "interface name to check")
	flag.BoolVar(&useValueReceiver, "value-receiver", false, "use value receiver instead of pointer receiver")
	flag.Parse()

	if input == "" {
		fmt.Println("error: input is required")
		flag.Usage()
		os.Exit(1)
	}

	var parseResult *parser.ParseResult
	var err error

	if mode == "file" {
		parseResult, err = parser.ParseFile(input)
		if err != nil {
			fmt.Printf("error parsing file: %v\n", err)
			os.Exit(1)
		}
	} else if mode == "inline" {
		parseResult, err = parseInlineJSON(input)
		if err != nil {
			fmt.Printf("error parsing inline JSON: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("error: unknown mode %s\n", mode)
		flag.Usage()
		os.Exit(1)
	}

	if len(parseResult.Structs) == 0 {
		fmt.Println("error: no structs found in input")
		os.Exit(1)
	}

	if len(parseResult.Interfaces) == 0 {
		fmt.Println("error: no interfaces found in input")
		os.Exit(1)
	}

	var targetStructName string
	if structName == "" {
		if len(parseResult.Structs) > 1 {
			fmt.Println("error: multiple structs found, please specify -struct")
			for _, s := range parseResult.Structs {
				fmt.Printf("  - %s\n", s.Name)
			}
			os.Exit(1)
		}
		targetStructName = parseResult.Structs[0].Name
	} else {
		found := false
		for _, s := range parseResult.Structs {
			if s.Name == structName {
				targetStructName = s.Name
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("error: struct %s not found\n", structName)
			os.Exit(1)
		}
	}

	checkAll := interfaceName == ""
	var targetInterfaces []models.InterfaceType

	if checkAll {
		targetInterfaces = parseResult.Interfaces
	} else {
		found := false
		for _, i := range parseResult.Interfaces {
			if i.Name == interfaceName {
				targetInterfaces = append(targetInterfaces, i)
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("error: interface %s not found\n", interfaceName)
			os.Exit(1)
		}
	}

	req := models.CheckAllRequest{
		TargetStruct:    targetStructName,
		Structs:         parseResult.Structs,
		Interfaces:      targetInterfaces,
		UseValueReceiver: useValueReceiver,
	}

	resp, err := sendCheckAllRequest(serverURL, req)
	if err != nil {
		fmt.Printf("error sending request: %v\n", err)
		os.Exit(1)
	}

	printResults(resp)
}

func parseInlineJSON(input string) (*parser.ParseResult, error) {
	var result parser.ParseResult
	err := json.Unmarshal([]byte(input), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func sendCheckAllRequest(serverURL string, req models.CheckAllRequest) (*models.CheckAllResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := strings.TrimRight(serverURL, "/") + "/check-all"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result models.CheckAllResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func printResults(resp *models.CheckAllResponse) {
	for _, result := range resp.Results {
		fmt.Println("==================================================")
		fmt.Printf("Checking against interface: %s\n", result.InterfaceName)
		fmt.Println("==================================================")

		if result.Satisfies {
			fmt.Printf("\u2713 %s\n", result.Message)
		} else {
			fmt.Printf("\u2717 %s\n", result.Message)
		}

		if len(result.MissingMethods) > 0 {
			fmt.Println("\nMissing methods:")
			for _, m := range result.MissingMethods {
				fmt.Printf("  - %s\n", m.MethodName)
			}
		}

		if len(result.MismatchedMethods) > 0 {
			fmt.Println("\nMismatched methods:")
			for _, m := range result.MismatchedMethods {
				fmt.Printf("  - %s:\n", m.MethodName)
				for _, d := range m.Details {
					fmt.Printf("    * %s: expected %s, got %s\n", d.IssueType, d.Expected, d.Actual)
				}
			}
		}

		fmt.Println()
	}
}
