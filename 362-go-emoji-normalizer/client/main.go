package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"emoji-normalizer/protocol"
)

const (
	defaultServerURL = "http://localhost:8080"
	apiEndpoint      = "/api/emoji"
)

func main() {
	var (
		operation   string
		text        string
		replacement string
		serverURL   string
	)

	flag.StringVar(&operation, "op", "", "Operation: extract, count, remove, replace, check_only, contains")
	flag.StringVar(&text, "text", "", "Input text to process")
	flag.StringVar(&replacement, "replace", "[表情]", "Replacement text for replace operation")
	flag.StringVar(&serverURL, "server", defaultServerURL, "Server URL (default: http://localhost:8080)")
	flag.Parse()

	if operation == "" {
		fmt.Println("Error: operation is required")
		printUsage()
		os.Exit(1)
	}

	if text == "" {
		text = strings.Join(flag.Args(), " ")
	}

	if text == "" {
		fmt.Println("Error: text is required")
		printUsage()
		os.Exit(1)
	}

	op, err := validateOperation(operation)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		printUsage()
		os.Exit(1)
	}

	client := NewClient(serverURL + apiEndpoint)

	req := protocol.EmojiRequest{
		Text:        text,
		Operation:   op,
		Replacement: replacement,
	}

	resp, err := client.SendRequest(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		os.Exit(1)
	}

	printResponse(*resp, op)
}

func validateOperation(op string) (string, error) {
	validOps := map[string]string{
		"extract":    protocol.OpExtract,
		"count":      protocol.OpCount,
		"remove":     protocol.OpRemove,
		"replace":    protocol.OpReplace,
		"check_only": protocol.OpCheckOnly,
		"contains":   protocol.OpContains,
	}

	if normalized, ok := validOps[op]; ok {
		return normalized, nil
	}

	return "", fmt.Errorf("invalid operation: %s", op)
}

func printResponse(resp protocol.EmojiResponse, op string) {
	switch op {
	case protocol.OpExtract:
		fmt.Printf("Found %d emoji(s):\n", resp.EmojiCount)
		for i, e := range resp.Emojis {
			fmt.Printf("  [%d] %q (bytes %d-%d)\n", i+1, e.Emoji, e.StartByte, e.EndByte)
		}
	case protocol.OpCount:
		fmt.Printf("Emoji count: %d\n", resp.EmojiCount)
	case protocol.OpRemove:
		fmt.Printf("Result: %q\n", resp.Result)
	case protocol.OpReplace:
		fmt.Printf("Result: %q\n", resp.Result)
	case protocol.OpCheckOnly:
		fmt.Printf("Is only emojis: %v\n", resp.IsOnlyEmojis)
	case protocol.OpContains:
		fmt.Printf("Contains emoji: %v\n", resp.Result == "true")
	}
}

func printUsage() {
	fmt.Println("\nUsage: emoji-client -op <operation> -text <text> [options]")
	fmt.Println("\nOperations:")
	fmt.Println("  extract    - Extract all emojis with positions")
	fmt.Println("  count      - Count number of emojis")
	fmt.Println("  remove     - Remove all emojis from text")
	fmt.Println("  replace    - Replace emojis with specified text")
	fmt.Println("  check_only - Check if text contains only emojis")
	fmt.Println("  contains   - Check if text contains any emoji")
	fmt.Println("\nOptions:")
	fmt.Println("  -replace   Replacement text for 'replace' operation (default: '[表情]')")
	fmt.Println("  -server    Server URL (default: http://localhost:8080)")
	fmt.Println("\nExamples:")
	fmt.Println("  emoji-client -op count -text \"Hello 👋 World 🌍\"")
	fmt.Println("  emoji-client -op replace -text \"Hello 👋\" -replace \"[EMOJI]\"")
	fmt.Println("  emoji-client -op extract \"Hello 👨‍👩‍👧‍👦 World\"")
}
