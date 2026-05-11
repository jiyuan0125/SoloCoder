package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/creditcard-gen/pkg/api"
)

const (
	defaultServerURL = "http://localhost:8502"
	warningMessage   = "⚠️  警告：以下为测试用途的虚拟卡号，仅用于开发和测试环境，不可用于真实交易。"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = defaultServerURL
	}

	command := os.Args[1]

	switch command {
	case "generate":
		handleGenerate(serverURL, os.Args[2:])
	case "validate":
		handleValidate(serverURL, os.Args[2:])
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Credit Card Generator Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client generate --type <card-type> --count <number>")
	fmt.Println("  client validate <card-number>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  generate   Generate test credit card numbers")
	fmt.Println("  validate   Validate a credit card number")
	fmt.Println()
	fmt.Println("Supported card types: visa, mastercard, amex, jcb")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  client generate --type visa --count 5")
	fmt.Println("  client validate 4532015112830366")
}

func handleGenerate(serverURL string, args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	cardType := fs.String("type", "", "Card type (visa, mastercard, amex, jcb)")
	count := fs.Int("count", 1, "Number of cards to generate (max 1000)")

	if err := fs.Parse(args); err != nil {
		fmt.Println("Error parsing arguments:", err)
		os.Exit(1)
	}

	if *cardType == "" {
		fmt.Println("Error: --type is required")
		fs.Usage()
		os.Exit(1)
	}

	req := api.GenerateRequest{
		CardType: *cardType,
		Count:    *count,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Println("Error creating request:", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/generate", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errMsg, ok := errResp["error"].(string); ok {
			fmt.Printf("Error: %s\n", errMsg)
		} else {
			fmt.Printf("Server error: HTTP %d\n", resp.StatusCode)
		}
		os.Exit(1)
	}

	var genResp api.GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		fmt.Println("Error reading response:", err)
		os.Exit(1)
	}

	if !genResp.Success {
		fmt.Printf("Error: %s\n", genResp.Error)
		os.Exit(1)
	}

	fmt.Println(warningMessage)
	fmt.Println()
	fmt.Printf("Generated %d %s test card(s):\n", genResp.Count, genResp.Cards[0].DisplayName)
	fmt.Println()

	for i, card := range genResp.Cards {
		fmt.Printf("  %d. %s\n", i+1, card.Number)
	}

	fmt.Println()
	fmt.Println(genResp.Warning)
}

func handleValidate(serverURL string, args []string) {
	if len(args) < 1 {
		fmt.Println("Error: card number is required")
		fmt.Println("Usage: client validate <card-number>")
		os.Exit(1)
	}

	cardNumber := strings.TrimSpace(args[0])

	req := api.ValidateRequest{
		CardNumber: cardNumber,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Println("Error creating request:", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/validate", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errMsg, ok := errResp["error"].(string); ok {
			fmt.Printf("Error: %s\n", errMsg)
		} else {
			fmt.Printf("Server error: HTTP %d\n", resp.StatusCode)
		}
		os.Exit(1)
	}

	var valResp api.ValidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&valResp); err != nil {
		fmt.Println("Error reading response:", err)
		os.Exit(1)
	}

	if !valResp.Success {
		fmt.Printf("Error: %s\n", valResp.Error)
		os.Exit(1)
	}

	fmt.Println(warningMessage)
	fmt.Println()
	fmt.Printf("Card Number: %s\n", valResp.CardNumber)
	fmt.Printf("Valid:       %s\n", formatBool(valResp.Valid))
	if valResp.Valid {
		fmt.Printf("Card Type:   %s\n", valResp.CardType)
	}
	fmt.Printf("Message:     %s\n", valResp.Message)
	fmt.Println()
	fmt.Println(valResp.Warning)
}

func formatBool(b bool) string {
	if b {
		return "Yes ✅"
	}
	return "No ❌"
}
