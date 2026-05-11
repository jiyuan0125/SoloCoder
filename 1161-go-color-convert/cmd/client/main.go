package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"colorconv/pkg/common"
)

var serverURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	if envURL := os.Getenv("COLORCONV_SERVER"); envURL != "" {
		serverURL = envURL
	}

	switch {
	case cmd == "brightness":
		if len(args) != 1 {
			fmt.Println("Usage: colorconv brightness <hex>")
			os.Exit(1)
		}
		if err := handleBrightness(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case strings.Contains(cmd, "-to-"):
		parts := strings.SplitN(cmd, "-to-", 2)
		if len(parts) != 2 {
			printUsage()
			os.Exit(1)
		}
		from := parts[0]
		to := parts[1]
		if err := handleConvert(from, to, args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		printUsage()
		os.Exit(1)
	}
}

func handleConvert(from, to string, args []string) error {
	expectedCount := map[string]int{
		"rgb":  3,
		"hsl":  3,
		"hsv":  3,
		"cmyk": 4,
		"hex":  1,
	}

	count, ok := expectedCount[strings.ToLower(from)]
	if !ok {
		return fmt.Errorf("unsupported source format: %s", from)
	}

	if len(args) != count {
		return fmt.Errorf("%s requires %d values", strings.ToUpper(from), count)
	}

	req := common.ConvertRequest{
		From:   strings.ToLower(from),
		Values: args,
		To:     strings.ToLower(to),
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(serverURL+"/convert", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	var result common.ConvertResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf(result.Error)
	}

	printResult(to, result.Result)
	return nil
}

func handleBrightness(hex string) error {
	req := common.BrightnessRequest{Hex: hex}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(serverURL+"/brightness", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	var result common.BrightnessResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf(result.Error)
	}

	fmt.Printf("Relative luminance: %.4f\n", result.Luminance)
	return nil
}

func printResult(to string, result interface{}) {
	switch strings.ToLower(to) {
	case "rgb":
		r := result.(map[string]interface{})
		fmt.Printf("RGB(%d, %d, %d)\n", int(r["r"].(float64)), int(r["g"].(float64)), int(r["b"].(float64)))
	case "hex":
		h := result.(map[string]interface{})
		fmt.Println(h["hex"].(string))
	case "hsl":
		h := result.(map[string]interface{})
		fmt.Printf("HSL(%.2f°, %.2f, %.2f)\n", h["h"].(float64), h["s"].(float64), h["l"].(float64))
	case "hsv":
		h := result.(map[string]interface{})
		fmt.Printf("HSV(%.2f°, %.2f, %.2f)\n", h["h"].(float64), h["s"].(float64), h["v"].(float64))
	case "cmyk":
		c := result.(map[string]interface{})
		fmt.Printf("CMYK(%.2f, %.2f, %.2f, %.2f)\n", c["c"].(float64), c["m"].(float64), c["y"].(float64), c["k"].(float64))
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  colorconv <from>-to-<to> <values...>")
	fmt.Println("  colorconv brightness <hex>")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  colorconv rgb-to-hsl 255 128 0")
	fmt.Println("  colorconv rgb-to-cmyk 255 0 0")
	fmt.Println("  colorconv hsl-to-rgb 30 1 0.5")
	fmt.Println("  colorconv hex-to-rgb #FF8000")
	fmt.Println("  colorconv brightness #FF8000")
	fmt.Println("")
	fmt.Println("Supported formats: rgb, hsl, hsv, cmyk, hex")
	fmt.Println("")
	fmt.Println("Environment variables:")
	fmt.Println("  COLORCONV_SERVER  Server URL (default: http://localhost:8080)")
}
