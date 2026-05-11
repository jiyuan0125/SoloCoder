package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"isbn-validator/pkg/api"
)

const serverURL = "http://localhost:8501"

const (
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorReset  = "\033[0m"
	colorYellow = "\033[33m"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "check":
		if len(os.Args) < 3 {
			fmt.Println("Error: ISBN is required for check command")
			fmt.Println("Usage: client check <ISBN>")
			os.Exit(1)
		}
		isbn := os.Args[2]
		handleCheck(isbn)

	case "convert":
		if len(os.Args) < 3 {
			fmt.Println("Error: ISBN is required for convert command")
			fmt.Println("Usage: client convert <ISBN>")
			os.Exit(1)
		}
		isbn := os.Args[2]
		handleConvert(isbn)

	case "format":
		if len(os.Args) < 3 {
			fmt.Println("Error: ISBN is required for format command")
			fmt.Println("Usage: client format <ISBN>")
			os.Exit(1)
		}
		isbn := os.Args[2]
		handleFormat(isbn)

	case "help", "-h", "--help":
		printUsage()

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("ISBN Validator Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client check <ISBN>     Validate an ISBN")
	fmt.Println("  client convert <ISBN>   Convert ISBN-10 to ISBN-13 or vice versa")
	fmt.Println("  client format <ISBN>    Format ISBN with hyphens")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  client check 978-7-115-54291-3")
	fmt.Println("  client check 0306406152")
	fmt.Println("  client convert 0306406152")
	fmt.Println("  client convert 9780306406157")
	fmt.Println("  client format 9787115542913")
}

func handleCheck(isbn string) {
	reqBody, err := json.Marshal(api.CheckRequest{ISBN: isbn})
	if err != nil {
		fmt.Printf("Error: Failed to create request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/api/check", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Printf("Error: Failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result api.CheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error: Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("%s❌ Error: %s%s\n", colorRed, result.Message, colorReset)
		os.Exit(1)
	}

	fmt.Printf("ISBN: %s\n", isbn)
	fmt.Printf("Type: %s\n", result.ISBNType)

	if result.IsValid {
		fmt.Printf("%s✓ Valid%s\n", colorGreen, colorReset)
		fmt.Printf("Check digit: %s\n", result.CheckDigit)
	} else {
		fmt.Printf("%s✗ Invalid%s\n", colorRed, colorReset)
		fmt.Printf("Reason: %s\n", result.Message)
		if result.CalculatedCD != "" {
			fmt.Printf("Expected check digit: %s\n", result.CalculatedCD)
		}
	}
}

func handleConvert(isbn string) {
	reqBody, err := json.Marshal(api.ConvertRequest{ISBN: isbn})
	if err != nil {
		fmt.Printf("Error: Failed to create request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/api/convert", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Printf("Error: Failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result api.ConvertResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error: Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("%s❌ Error: %s%s\n", colorRed, result.Message, colorReset)
		os.Exit(1)
	}

	fmt.Printf("Original: %s (%s)\n", result.Original, result.OriginalType)
	fmt.Printf("Converted: %s (%s)\n", result.Converted, result.ConvertedType)
	fmt.Printf("Check digit: %s\n", result.CheckDigit)
	fmt.Printf("%s✓ Conversion successful%s\n", colorGreen, colorReset)
}

func handleFormat(isbn string) {
	reqBody, err := json.Marshal(api.FormatRequest{ISBN: isbn})
	if err != nil {
		fmt.Printf("Error: Failed to create request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/api/format", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Printf("Error: Failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result api.FormatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error: Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("%s❌ Error: %s%s\n", colorRed, result.Message, colorReset)
		os.Exit(1)
	}

	fmt.Printf("Original: %s\n", result.Original)
	fmt.Printf("Formatted: %s\n", result.Formatted)
	fmt.Printf("%s✓ Format successful%s\n", colorGreen, colorReset)
}
