package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/solocoder/email-parser/pkg/api"
)

var serverURL string

func main() {
	flag.StringVar(&serverURL, "server", "http://localhost:8080", "Server URL")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]
	email := ""
	if len(args) > 1 {
		email = args[1]
	}

	switch command {
	case "parse":
		if email == "" {
			fmt.Println("Error: email address is required for parse command")
			os.Exit(1)
		}
		handleParse(email)
	case "validate":
		if email == "" {
			fmt.Println("Error: email address is required for validate command")
			os.Exit(1)
		}
		handleValidate(email)
	case "batch":
		if email == "" {
			fmt.Println("Error: email list is required for batch command")
			os.Exit(1)
		}
		handleBatch(email)
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Email Parser Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  email-client [options] <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  parse <email>       Parse and validate a single email address")
	fmt.Println("  validate <email>    Check if an email address is valid")
	fmt.Println("  batch <emails>      Parse multiple email addresses (comma or semicolon separated)")
	fmt.Println("  help                Show this help message")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -server <url>       Server URL (default: http://localhost:8080)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  email-client parse user+tag@example.com")
	fmt.Println("  email-client validate invalid-email")
	fmt.Println("  email-client batch \"a@example.com, b@test.com\"")
}

func handleParse(email string) {
	req := api.ParseRequest{Email: email}
	body, _ := json.Marshal(req)

	resp, err := http.Post(serverURL+"/api/parse", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result api.ParseResponse
	json.Unmarshal(respBody, &result)

	if !result.Success {
		fmt.Printf("Request failed: %s\n", result.Error)
		os.Exit(1)
	}

	if !result.Valid {
		fmt.Printf("Invalid email: %s\n", result.Error)
		os.Exit(1)
	}

	addr := result.Address
	fmt.Println("Email Parsed Successfully!")
	fmt.Println("-------------------------")
	fmt.Printf("Original: %s\n", addr.Original)
	fmt.Printf("Local: %s\n", addr.Local)
	fmt.Printf("Domain: %s\n", addr.Domain)
	if addr.HasAlias {
		fmt.Printf("Has alias: yes\n")
		fmt.Printf("Base local: %s\n", addr.BaseLocal)
		fmt.Printf("Base address: %s\n", addr.BaseAddress)
	} else {
		fmt.Println("Has alias: no")
	}
	if addr.IsQuotedLocal {
		fmt.Println("Local is quoted: yes")
	}
}

func handleValidate(email string) {
	req := api.ValidateRequest{Email: email}
	body, _ := json.Marshal(req)

	resp, err := http.Post(serverURL+"/api/validate", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result api.ValidateResponse
	json.Unmarshal(respBody, &result)

	if !result.Success {
		fmt.Println("Request failed")
		os.Exit(1)
	}

	if result.Valid {
		fmt.Printf("Valid: %s\n", email)
		os.Exit(0)
	} else {
		fmt.Printf("Invalid: %s\n", email)
		os.Exit(1)
	}
}

func handleBatch(emails string) {
	req := api.BatchParseRequest{Emails: emails}
	body, _ := json.Marshal(req)

	resp, err := http.Post(serverURL+"/api/batch", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result api.BatchParseResponse
	json.Unmarshal(respBody, &result)

	if !result.Success {
		fmt.Printf("Request failed: %s\n", result.Error)
		os.Exit(1)
	}

	fmt.Printf("Batch Parse Results\n")
	fmt.Printf("===================\n")
	fmt.Printf("Total: %d, Valid: %d, Invalid: %d\n\n", result.Total, result.Valid, result.Invalid)

	for i, r := range result.Results {
		fmt.Printf("[%d] ", i+1)
		if r.Valid {
			fmt.Printf("✓ Valid: %s\n", r.Address.Original)
			fmt.Printf("    Local: %s, Domain: %s\n", r.Address.Local, r.Address.Domain)
			if r.Address.HasAlias {
				fmt.Printf("    Base: %s\n", r.Address.BaseAddress)
			}
		} else {
			fmt.Printf("✗ Invalid")
			if r.Email != "" {
				fmt.Printf(": %s", r.Email)
			}
			if r.Error != "" {
				fmt.Printf(" (%s)", r.Error)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	if result.Invalid > 0 {
		os.Exit(1)
	}
}
