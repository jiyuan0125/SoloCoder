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

	"saml-parser/pkg/model"
)

const (
	defaultServerURL = "http://localhost:8080"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := getServerURL()

	switch os.Args[1] {
	case "parse":
		handleParse(serverURL)
	case "validate":
		handleValidate(serverURL)
	case "extract":
		handleExtract(serverURL)
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("SAML Parser CLI - A tool for parsing and validating SAML assertions")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  saml parse <file>")
	fmt.Println("  saml validate <file> --audience <sp-id>")
	fmt.Println("  saml extract <file> --attribute <name>")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  SAML_SERVER_URL  Server URL (default: http://localhost:8080)")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  parse     Parse a SAML assertion XML file")
	fmt.Println("  validate  Validate a SAML assertion")
	fmt.Println("  extract   Extract specific attribute values from assertion")
}

func getServerURL() string {
	if envURL := os.Getenv("SAML_SERVER_URL"); envURL != "" {
		return envURL
	}
	return defaultServerURL
}

func handleParse(serverURL string) {
	if len(os.Args) < 3 {
		fmt.Println("Error: Missing file argument")
		fmt.Println("Usage: saml parse <file>")
		os.Exit(1)
	}

	filePath := os.Args[2]
	xmlContent, err := readXMLFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	req := model.ParseRequest{
		XML: xmlContent,
	}

	resp, err := httpPost(serverURL+"/saml/parse", req)
	if err != nil {
		fmt.Printf("Error calling server: %v\n", err)
		os.Exit(1)
	}

	var parseResp model.ParseResponse
	if err := json.Unmarshal(resp, &parseResp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	printParseResult(&parseResp)
}

func handleValidate(serverURL string) {
	validateCmd := flag.NewFlagSet("validate", flag.ExitOnError)
	audience := validateCmd.String("audience", "", "Service Provider Entity ID (required)")
	
	if len(os.Args) < 3 {
		fmt.Println("Error: Missing file argument")
		fmt.Println("Usage: saml validate <file> --audience <sp-id>")
		os.Exit(1)
	}
	
	filePath := os.Args[2]
	
	if err := validateCmd.Parse(os.Args[3:]); err != nil {
		fmt.Println("Error parsing flags")
		os.Exit(1)
	}

	xmlContent, err := readXMLFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	req := model.ValidateRequest{
		XML:      xmlContent,
		Audience: *audience,
	}

	resp, err := httpPost(serverURL+"/saml/validate", req)
	if err != nil {
		fmt.Printf("Error calling server: %v\n", err)
		os.Exit(1)
	}

	var validateResp model.ValidateResponse
	if err := json.Unmarshal(resp, &validateResp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	printValidateResult(&validateResp)
}

func handleExtract(serverURL string) {
	extractCmd := flag.NewFlagSet("extract", flag.ExitOnError)
	attrName := extractCmd.String("attribute", "", "Attribute name to extract (required)")
	
	if len(os.Args) < 3 {
		fmt.Println("Error: Missing file argument")
		fmt.Println("Usage: saml extract <file> --attribute <name>")
		os.Exit(1)
	}
	
	filePath := os.Args[2]
	
	if err := extractCmd.Parse(os.Args[3:]); err != nil {
		fmt.Println("Error parsing flags")
		os.Exit(1)
	}

	if *attrName == "" {
		fmt.Println("Error: --attribute flag is required")
		fmt.Println("Usage: saml extract <file> --attribute <name>")
		os.Exit(1)
	}

	xmlContent, err := readXMLFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	parseReq := model.ParseRequest{
		XML: xmlContent,
	}

	resp, err := httpPost(serverURL+"/saml/parse", parseReq)
	if err != nil {
		fmt.Printf("Error calling server: %v\n", err)
		os.Exit(1)
	}

	var parseResp model.ParseResponse
	if err := json.Unmarshal(resp, &parseResp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if !parseResp.Success {
		fmt.Printf("Parse failed: %s\n", parseResp.Error)
		os.Exit(1)
	}

	found := false
	for _, attr := range parseResp.Attributes {
		if strings.EqualFold(attr.Name, *attrName) {
			for _, val := range attr.Values {
				fmt.Println(val)
			}
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("Attribute '%s' not found or has no values\n", *attrName)
		os.Exit(0)
	}
}

func readXMLFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func httpPost(url string, req interface{}) ([]byte, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func printParseResult(resp *model.ParseResponse) {
	if !resp.Success {
		fmt.Printf("Parse failed: %s\n", resp.Error)
		return
	}

	fmt.Println("=== SAML Assertion Parse Result ===")
	fmt.Println()

	fmt.Println("Subject:")
	if resp.Subject.NameID != "" {
		fmt.Printf("  NameID: %s\n", resp.Subject.NameID)
	}
	if resp.Subject.SubjectConfirmation.Method != "" {
		fmt.Printf("  SubjectConfirmation Method: %s\n", resp.Subject.SubjectConfirmation.Method)
	}
	if resp.Subject.SubjectConfirmation.Data.Recipient != "" {
		fmt.Printf("  SubjectConfirmationData Recipient: %s\n", resp.Subject.SubjectConfirmation.Data.Recipient)
	}
	fmt.Println()

	fmt.Println("Conditions:")
	if resp.Conditions.NotBefore != "" {
		fmt.Printf("  NotBefore: %s\n", resp.Conditions.NotBefore)
	}
	if resp.Conditions.NotOnOrAfter != "" {
		fmt.Printf("  NotOnOrAfter: %s\n", resp.Conditions.NotOnOrAfter)
	}
	if len(resp.Conditions.AudienceRestriction.Audiences) > 0 {
		fmt.Println("  Audiences:")
		for _, aud := range resp.Conditions.AudienceRestriction.Audiences {
			fmt.Printf("    - %s\n", aud)
		}
	}
	fmt.Println()

	fmt.Println("Attributes:")
	if len(resp.Attributes) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, attr := range resp.Attributes {
			fmt.Printf("  %s:\n", attr.Name)
			for _, val := range attr.Values {
				fmt.Printf("    - %s\n", val)
			}
		}
	}
	fmt.Println()

	fmt.Println("Signature:")
	if resp.Signature.HasSignature {
		fmt.Printf("  Has Signature: Yes\n")
		if resp.Signature.Algorithm != "" {
			fmt.Printf("  Algorithm: %s\n", resp.Signature.Algorithm)
		}
	} else {
		fmt.Println("  Has Signature: No")
	}
}

func printValidateResult(resp *model.ValidateResponse) {
	if !resp.Success {
		fmt.Printf("Validation request failed: %v\n", resp.Errors)
		return
	}

	fmt.Println("=== SAML Assertion Validation Result ===")
	fmt.Println()

	if resp.Valid {
		fmt.Println("Status: VALID")
	} else {
		fmt.Println("Status: INVALID")
		if len(resp.Errors) > 0 {
			fmt.Println()
			fmt.Println("Errors:")
			for _, err := range resp.Errors {
				fmt.Printf("  - %s\n", err)
			}
		}
	}
}
