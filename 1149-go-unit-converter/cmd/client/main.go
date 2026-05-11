package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"go-unit-converter/pkg/api"
)

const defaultServerURL = "http://localhost:8080"

func main() {
	args := os.Args[1:]

	if len(args) < 3 {
		printUsage()
		os.Exit(1)
	}

	serverURL := os.Getenv("UNITCONV_SERVER")
	if serverURL == "" {
		serverURL = defaultServerURL
	}

	var value float64
	var from, to string
	tempDelta := false

	argIndex := 0
	if args[0] == "--delta" || args[0] == "-d" {
		tempDelta = true
		argIndex = 1
	}

	if len(args)-argIndex < 3 {
		printUsage()
		os.Exit(1)
	}

	valueStr := args[argIndex]
	from = args[argIndex+1]
	to = args[argIndex+2]

	var err error
	value, err = strconv.ParseFloat(valueStr, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid value: %s\n", valueStr)
		os.Exit(1)
	}

	req := api.ConvertRequest{
		Value:     value,
		From:      from,
		To:        to,
		TempDelta: tempDelta,
	}

	result, err := doConvert(serverURL, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s %s = %s %s\n",
		formatNumber(result.Original),
		strings.ToUpper(result.From),
		formatNumber(result.Value),
		strings.ToUpper(result.To),
	)
}

func doConvert(serverURL string, req api.ConvertRequest) (*api.ConvertResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(serverURL+"/convert", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("%s", errResp.Error)
	}

	var result api.ConvertResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func formatNumber(f float64) string {
	s := strconv.FormatFloat(f, 'f', -1, 64)
	if !strings.Contains(s, ".") {
		return s
	}
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

func printUsage() {
	fmt.Println("Usage: unitconv [--delta|-d] <value> <from_unit> <to_unit>")
	fmt.Println("  --delta, -d  Use temperature delta mode instead of point conversion")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  unitconv 100 km m")
	fmt.Println("  unitconv 98.6 F C")
	fmt.Println("  unitconv 100 km/h m/s")
	fmt.Println("  unitconv 1000 kg/m³ g/cm³")
	fmt.Println("  unitconv --delta 1 C F")
	fmt.Println("")
	fmt.Println("Environment:")
	fmt.Println("  UNITCONV_SERVER  Server URL (default: http://localhost:8080)")
}
