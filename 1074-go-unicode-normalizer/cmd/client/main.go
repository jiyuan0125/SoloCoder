package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"unicode-normalizer/pkg/api"
)

const version = "1.0.0"

func main() {
	var (
		inputFile   string
		outputFile  string
		form        string
		text        string
		serverURL   string
		analyzeMode bool
		showVersion bool
	)

	flag.StringVar(&inputFile, "input", "", "Input file path (or use stdin)")
	flag.StringVar(&outputFile, "output", "", "Output file path (or use stdout)")
	flag.StringVar(&form, "form", "NFC", "Normalization form: NFC, NFD, NFKC, NFKD")
	flag.StringVar(&text, "text", "", "Input text (alternative to file)")
	flag.StringVar(&serverURL, "server", "http://localhost:8080", "Server URL")
	flag.BoolVar(&analyzeMode, "analyze", false, "Analyze changes mode")
	flag.BoolVar(&showVersion, "version", false, "Show version")

	flag.Parse()

	if showVersion {
		fmt.Printf("Unicode Normalizer Client v%s\n", version)
		return
	}

	if err := validateForm(form); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	inputText, err := readInput(inputFile, text)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	var result string
	if analyzeMode {
		result, err = doAnalyze(serverURL, inputText, form)
	} else {
		result, err = doNormalize(serverURL, inputText, form)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := writeOutput(outputFile, result); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}
}

func validateForm(form string) error {
	form = strings.ToUpper(form)
	validForms := []string{"NFC", "NFD", "NFKC", "NFKD"}
	for _, f := range validForms {
		if form == f {
			return nil
		}
	}
	return fmt.Errorf("invalid form: %s (must be one of: %s)", form, strings.Join(validForms, ", "))
}

func readInput(inputFile, text string) (string, error) {
	if text != "" {
		return text, nil
	}

	if inputFile != "" {
		data, err := os.ReadFile(inputFile)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	return "", fmt.Errorf("no input provided. Use -input, -text, or pipe from stdin")
}

func writeOutput(outputFile, result string) error {
	if outputFile != "" {
		return os.WriteFile(outputFile, []byte(result), 0644)
	}
	fmt.Print(result)
	return nil
}

func doNormalize(serverURL, text, form string) (string, error) {
	reqBody := api.NormalizeRequest{
		Text: text,
		Form: form,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(serverURL+"/api/normalize", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	var response api.NormalizeResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if !response.Success {
		return "", fmt.Errorf("server error: %s", response.Error)
	}

	return response.NormalizedText, nil
}

func doAnalyze(serverURL, text, form string) (string, error) {
	reqBody := api.AnalyzeRequest{
		Text: text,
		Form: form,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(serverURL+"/api/analyze", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	var response api.AnalyzeResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if !response.Success {
		return "", fmt.Errorf("server error: %s", response.Error)
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Unicode Normalization Analysis\n"))
	builder.WriteString(fmt.Sprintf("================================\n\n"))
	builder.WriteString(fmt.Sprintf("Form: %s\n", response.Form))
	builder.WriteString(fmt.Sprintf("Unicode Version: %s\n\n", response.UnicodeVersion))
	builder.WriteString(fmt.Sprintf("Changes found: %d\n\n", response.ChangeCount))

	if response.ChangeCount > 0 {
		builder.WriteString("Detailed Changes:\n")
		builder.WriteString("-----------------\n\n")
		for i, change := range response.Changes {
			builder.WriteString(fmt.Sprintf("%d. Original: %s (%s)\n", i+1, change.Original, change.OriginalCode))
			builder.WriteString(fmt.Sprintf("   Normalized: %s (%s)\n\n", change.Normalized, strings.Join(change.NormalizedCodes, " ")))
		}
	}

	builder.WriteString(fmt.Sprintf("\nNormalized Result:\n"))
	builder.WriteString(fmt.Sprintf("------------------\n"))
	builder.WriteString(response.NormalizedText)

	return builder.String(), nil
}
