package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"iban/internal/api"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const (
	defaultServerURL = "http://localhost:8604"
)

func main() {
	var iban string
	var inputFile string
	var serverURL string
	var formatMode bool

	flag.StringVar(&iban, "iban", "", "IBAN to validate")
	flag.StringVar(&inputFile, "file", "", "File containing IBANs to validate (one per line)")
	flag.StringVar(&serverURL, "server", defaultServerURL, "Server URL")
	flag.BoolVar(&formatMode, "format", false, "Format the IBAN instead of validating")
	flag.Parse()

	if iban == "" && inputFile == "" {
		fmt.Println("Usage:")
		fmt.Println("  iban-client -iban <iban>                    Validate a single IBAN")
		fmt.Println("  iban-client -iban <iban> -format            Format a single IBAN")
		fmt.Println("  iban-client -file <file>                    Validate IBANs from a file")
		fmt.Println("  iban-client -server <url>                   Specify custom server URL")
		os.Exit(1)
	}

	if inputFile != "" {
		validateFromFile(inputFile, serverURL)
		return
	}

	if formatMode {
		formatIBAN(iban, serverURL)
	} else {
		validateIBAN(iban, serverURL)
	}
}

func validateIBAN(ibanStr, serverURL string) {
	req := api.ValidateRequest{IBAN: ibanStr}
	resp, err := sendRequest(serverURL+"/validate", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var result api.ValidateResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	printSingleResult(ibanStr, result)
}

func formatIBAN(ibanStr, serverURL string) {
	req := api.FormatRequest{IBAN: ibanStr}
	resp, err := sendRequest(serverURL+"/format", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var result api.FormatResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Original:  %s\n", ibanStr)
	fmt.Printf("Formatted: %s\n", result.Formatted)
}

func validateFromFile(filePath, serverURL string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	var ibans []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			ibans = append(ibans, line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	if len(ibans) == 0 {
		fmt.Println("No valid IBANs found in file")
		return
	}

	req := api.BatchValidateRequest{IBANs: ibans}
	resp, err := sendRequest(serverURL+"/batch-validate", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var result api.BatchValidateResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	printBatchResults(ibans, result.Results)
}

func sendRequest(url string, req interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server error: %s", string(body))
	}

	return body, nil
}

func printSingleResult(ibanStr string, result api.ValidateResponse) {
	fmt.Printf("IBAN:        %s\n", ibanStr)
	fmt.Printf("Formatted:   %s\n", result.Formatted)
	fmt.Printf("Valid:       %s\n", boolToYesNo(result.Valid))
	if result.CountryName != "" {
		fmt.Printf("Country:     %s (%s)\n", result.CountryName, result.CountryCode)
	}
	fmt.Printf("BBAN:        %s\n", result.BBAN)
	fmt.Printf("Format:      %s\n", boolToValidInvalid(result.FormatValid))
	fmt.Printf("Checksum:    %s\n", boolToValidInvalid(result.ChecksumValid))
}

func printBatchResults(ibans []string, results []api.ValidateResponse) {
	fmt.Printf("Batch Validation Results (%d IBANs):\n\n", len(results))
	fmt.Printf("%-4s %-35s %-8s %-20s\n", "Num", "IBAN", "Valid", "Country")
	fmt.Println(strings.Repeat("-", 75))

	validCount := 0
	invalidCount := 0

	for i, result := range results {
		iban := ibans[i]
		if len(iban) > 32 {
			iban = iban[:32] + "..."
		}

		country := result.CountryName
		if country == "" {
			country = "-"
		}

		fmt.Printf("%-4d %-35s %-8s %-20s\n", i+1, iban, boolToYesNo(result.Valid), country)

		if result.Valid {
			validCount++
		} else {
			invalidCount++
		}
	}

	fmt.Println(strings.Repeat("-", 75))
	fmt.Printf("\nSummary: Valid=%d, Invalid=%d\n", validCount, invalidCount)
}

func boolToYesNo(b bool) string {
	if b {
		return "YES"
	}
	return "NO"
}

func boolToValidInvalid(b bool) string {
	if b {
		return "VALID"
	}
	return "INVALID"
}
