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

	"licenseplate/pkg/api"
)

type clientConfig struct {
	serverURL string
	plate     string
	filePath  string
}

func parseFlags() clientConfig {
	var config clientConfig

	flag.StringVar(&config.serverURL, "server", "http://localhost:8503", "Server URL")
	flag.StringVar(&config.plate, "plate", "", "Single license plate to validate")
	flag.StringVar(&config.filePath, "file", "", "File path for batch validation (one plate per line)")

	flag.Usage = func() {
		fmt.Println("License Plate Validator Client")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  license-client -plate 京A12345")
		fmt.Println("  license-client -file plates.txt")
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
	}

	flag.Parse()
	return config
}

func validateSingle(serverURL, plate string) (*api.ValidateResponse, error) {
	req := api.ValidateRequest{
		Plate: plate,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(serverURL+"/validate", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result api.ValidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func validateBatch(serverURL string, plates []string) (*api.BatchValidateResponse, error) {
	req := api.BatchValidateRequest{
		Plates: plates,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(serverURL+"/batch", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result api.BatchValidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func readPlatesFromFile(filePath string) ([]string, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var plates []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			plates = append(plates, line)
		}
	}

	return plates, nil
}

func printSingleResult(result *api.ValidateResponse) {
	fmt.Printf("车牌号: %s\n", result.Plate)
	fmt.Printf("是否有效: %v\n", result.Valid)
	fmt.Printf("车牌类型: %s\n", result.Type)
	if result.Reason != "" {
		fmt.Printf("失败原因: %s\n", result.Reason)
	}
}

func printBatchResults(results []api.ValidateResponse) {
	fmt.Printf("%-20s %-10s %-15s %s\n", "车牌号", "是否有效", "类型", "原因")
	fmt.Println(strings.Repeat("-", 80))

	validCount := 0
	invalidCount := 0

	for _, r := range results {
		validStr := "无效"
		if r.Valid {
			validStr = "有效"
			validCount++
		} else {
			invalidCount++
		}
		fmt.Printf("%-20s %-10s %-15s %s\n", r.Plate, validStr, r.Type, r.Reason)
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("总计: %d 个, 有效: %d 个, 无效: %d 个\n", len(results), validCount, invalidCount)
}

func main() {
	config := parseFlags()

	if config.plate == "" && config.filePath == "" {
		flag.Usage()
		os.Exit(1)
	}

	if config.plate != "" && config.filePath != "" {
		fmt.Println("Error: Please specify either -plate or -file, not both")
		flag.Usage()
		os.Exit(1)
	}

	if config.plate != "" {
		result, err := validateSingle(config.serverURL, config.plate)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printSingleResult(result)
	}

	if config.filePath != "" {
		plates, err := readPlatesFromFile(config.filePath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if len(plates) == 0 {
			fmt.Println("Error: No plates found in file")
			os.Exit(1)
		}

		results, err := validateBatch(config.serverURL, plates)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printBatchResults(results.Results)
	}
}
