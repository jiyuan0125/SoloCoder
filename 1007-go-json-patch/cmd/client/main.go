package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/example/jsonpatch/pkg/common"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8102", "JSON Patch server URL")
	documentFile := flag.String("doc", "", "Path to the JSON document file (required for apply)")
	patchFile := flag.String("patch", "", "Path to the JSON patch file (required)")
	outputFile := flag.String("output", "", "Path to the output file (default: overwrite input document)")
	dryRun := flag.Bool("dry-run", false, "Preview changes without writing to file")
	validateOnly := flag.Bool("validate", false, "Only validate the patch, don't apply")

	flag.Parse()

	if *patchFile == "" {
		fmt.Fprintln(os.Stderr, "Error: -patch is required")
		flag.Usage()
		os.Exit(1)
	}

	if *validateOnly {
		if err := validatePatch(*serverURL, *patchFile); err != nil {
			fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Patch is valid")
		return
	}

	if *documentFile == "" {
		fmt.Fprintln(os.Stderr, "Error: -doc is required for apply operation")
		flag.Usage()
		os.Exit(1)
	}

	result, err := applyPatch(*serverURL, *documentFile, *patchFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Apply failed: %v\n", err)
		os.Exit(1)
	}

	output := *outputFile
	if output == "" {
		output = *documentFile
	}

	if *dryRun {
		fmt.Println("Dry run mode - changes preview:")
		fmt.Println("=====================================")
		fmt.Println(string(result))
		fmt.Println("=====================================")
		fmt.Println("(File not modified)")
		return
	}

	if err := ioutil.WriteFile(output, result, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully applied patch to %s\n", output)
}

func readJSONFile(path string) ([]byte, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	
	var jsonCheck interface{}
	if err := json.Unmarshal(data, &jsonCheck); err != nil {
		return nil, fmt.Errorf("invalid JSON in %s: %w", path, err)
	}
	
	return data, nil
}

func validatePatch(serverURL, patchFile string) error {
	patchData, err := readJSONFile(patchFile)
	if err != nil {
		return err
	}

	req := common.ValidateRequest{
		Patch: patchData,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(serverURL+"/validate", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	var result common.ValidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("server error: %s", result.Error)
	}

	if !result.Valid {
		return fmt.Errorf("%s", result.Error)
	}

	return nil
}

func applyPatch(serverURL, documentFile, patchFile string) ([]byte, error) {
	docData, err := readJSONFile(documentFile)
	if err != nil {
		return nil, err
	}

	patchData, err := readJSONFile(patchFile)
	if err != nil {
		return nil, err
	}

	req := common.ApplyRequest{
		Document: docData,
		Patch:    patchData,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(serverURL+"/apply", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	var result common.ApplyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("%s", result.Error)
	}

	return result.Document, nil
}
