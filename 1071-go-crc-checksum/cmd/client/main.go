package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"crc-checksum/pkg/api"
)

type clientConfig struct {
	serverURL    string
	crcType      string
	variant      string
	data         string
	filePath     string
	expectedCRC  string
	mode         string
	poly32       string
	init32       string
	refIn32      bool
	refOut32     bool
	xorOut32     string
	poly64       string
	init64       string
	refIn64      bool
	refOut64     bool
	xorOut64     string
}

func readFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

func sendCalculate(cfg *clientConfig) error {
	var data []byte
	var err error

	if cfg.filePath != "" {
		data, err = readFile(cfg.filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}
	} else {
		data = []byte(cfg.data)
	}

	encoded := base64.StdEncoding.EncodeToString(data)

	req := api.CalculateRequest{
		Data:     encoded,
		IsBase64: true,
	}

	if strings.EqualFold(cfg.crcType, "crc32") {
		req.Type = api.CRCType32
		req.Variant = cfg.variant
		if cfg.poly32 != "" {
			req.CustomCRC32 = &api.CustomCRC32Config{
				Polynomial: cfg.poly32,
				Init:       cfg.init32,
				RefIn:      cfg.refIn32,
				RefOut:     cfg.refOut32,
				XorOut:     cfg.xorOut32,
			}
		}
	} else if strings.EqualFold(cfg.crcType, "crc64") {
		req.Type = api.CRCType64
		req.Variant = cfg.variant
		if cfg.poly64 != "" {
			req.CustomCRC64 = &api.CustomCRC64Config{
				Polynomial: cfg.poly64,
				Init:       cfg.init64,
				RefIn:      cfg.refIn64,
				RefOut:     cfg.refOut64,
				XorOut:     cfg.xorOut64,
			}
		}
	} else {
		return fmt.Errorf("invalid CRC type: %s, must be crc32 or crc64", cfg.crcType)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(cfg.serverURL+"/calculate", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	var result api.CalculateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if !result.Success {
		return fmt.Errorf("server error: %s", result.Error)
	}

	fmt.Printf("CRC Result: %s\n", result.CRC)
	return nil
}

func sendVerify(cfg *clientConfig) error {
	var data []byte
	var err error

	if cfg.filePath != "" {
		data, err = readFile(cfg.filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}
	} else {
		data = []byte(cfg.data)
	}

	encoded := base64.StdEncoding.EncodeToString(data)

	req := api.VerifyRequest{
		Data:        encoded,
		IsBase64:    true,
		ExpectedCRC: cfg.expectedCRC,
	}

	if strings.EqualFold(cfg.crcType, "crc32") {
		req.Type = api.CRCType32
		req.Variant = cfg.variant
		if cfg.poly32 != "" {
			req.CustomCRC32 = &api.CustomCRC32Config{
				Polynomial: cfg.poly32,
				Init:       cfg.init32,
				RefIn:      cfg.refIn32,
				RefOut:     cfg.refOut32,
				XorOut:     cfg.xorOut32,
			}
		}
	} else if strings.EqualFold(cfg.crcType, "crc64") {
		req.Type = api.CRCType64
		req.Variant = cfg.variant
		if cfg.poly64 != "" {
			req.CustomCRC64 = &api.CustomCRC64Config{
				Polynomial: cfg.poly64,
				Init:       cfg.init64,
				RefIn:      cfg.refIn64,
				RefOut:     cfg.refOut64,
				XorOut:     cfg.xorOut64,
			}
		}
	} else {
		return fmt.Errorf("invalid CRC type: %s, must be crc32 or crc64", cfg.crcType)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(cfg.serverURL+"/verify", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	var result api.VerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if !result.Success {
		return fmt.Errorf("server error: %s", result.Error)
	}

	if result.Match {
		fmt.Printf("✓ CRC Matches: %s\n", result.Actual)
	} else {
		fmt.Printf("✗ CRC Mismatch: expected %s, got %s\n", cfg.expectedCRC, result.Actual)
		os.Exit(1)
	}

	return nil
}

func listVariants(cfg *clientConfig) error {
	resp, err := http.Get(cfg.serverURL + "/variants")
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	var result api.VariantsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if !result.Success {
		return fmt.Errorf("server error: %s", result.Error)
	}

	fmt.Println("Available CRC32 Variants:")
	for name, desc := range result.CRC32 {
		fmt.Printf("  %s: %s\n", name, desc)
	}

	fmt.Println("\nAvailable CRC64 Variants:")
	for name, desc := range result.CRC64 {
		fmt.Printf("  %s: %s\n", name, desc)
	}

	return nil
}

func printUsage() {
	fmt.Println("CRC Checksum Client")
	fmt.Println("\nUsage:")
	fmt.Println("  client calculate [options]    Calculate CRC value")
	fmt.Println("  client verify [options]       Verify CRC value")
	fmt.Println("  client variants               List available variants")
	fmt.Println("\nCommon Options:")
	fmt.Println("  --server <url>               Server URL (default: http://localhost:8303)")
	fmt.Println("  --type <crc32|crc64>         CRC type (required for calculate/verify)")
	fmt.Println("  --variant <name>             CRC variant name")
	fmt.Println("  --data <string>              Input data string")
	fmt.Println("  --file <path>                Input file path")
	fmt.Println("\nCalculate Options:")
	fmt.Println("  (uses common options)")
	fmt.Println("\nVerify Options:")
	fmt.Println("  --expected <crc>             Expected CRC value (required)")
	fmt.Println("\nCustom CRC32 Options:")
	fmt.Println("  --poly32 <hex>               32-bit polynomial (hex)")
	fmt.Println("  --init32 <hex>               Initial value (hex)")
	fmt.Println("  --refin32                    Reflect input bytes")
	fmt.Println("  --refout32                   Reflect output")
	fmt.Println("  --xorout32 <hex>             Final XOR value (hex)")
	fmt.Println("\nCustom CRC64 Options:")
	fmt.Println("  --poly64 <hex>               64-bit polynomial (hex)")
	fmt.Println("  --init64 <hex>               Initial value (hex)")
	fmt.Println("  --refin64                    Reflect input bytes")
	fmt.Println("  --refout64                   Reflect output")
	fmt.Println("  --xorout64 <hex>             Final XOR value (hex)")
	fmt.Println("\nExamples:")
	fmt.Println("  client calculate --type crc32 --data \"hello world\"")
	fmt.Println("  client calculate --type crc64 --file data.bin")
	fmt.Println("  client verify --type crc32 --data \"hello world\" --expected 0x4A17B156")
	fmt.Println("  client variants")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	mode := os.Args[1]

	fs := flag.NewFlagSet("client", flag.ExitOnError)
	cfg := &clientConfig{mode: mode}

	fs.StringVar(&cfg.serverURL, "server", "http://localhost:8303", "Server URL")
	fs.StringVar(&cfg.crcType, "type", "crc32", "CRC type: crc32 or crc64")
	fs.StringVar(&cfg.variant, "variant", "", "CRC variant name")
	fs.StringVar(&cfg.data, "data", "", "Input data string")
	fs.StringVar(&cfg.filePath, "file", "", "Input file path")
	fs.StringVar(&cfg.expectedCRC, "expected", "", "Expected CRC value (for verify mode)")

	fs.StringVar(&cfg.poly32, "poly32", "", "Custom CRC32 polynomial (hex)")
	fs.StringVar(&cfg.init32, "init32", "0x00000000", "Custom CRC32 init value (hex)")
	fs.BoolVar(&cfg.refIn32, "refin32", false, "Reflect input bytes for CRC32")
	fs.BoolVar(&cfg.refOut32, "refout32", false, "Reflect output for CRC32")
	fs.StringVar(&cfg.xorOut32, "xorout32", "0x00000000", "Custom CRC32 xor_out value (hex)")

	fs.StringVar(&cfg.poly64, "poly64", "", "Custom CRC64 polynomial (hex)")
	fs.StringVar(&cfg.init64, "init64", "0x0000000000000000", "Custom CRC64 init value (hex)")
	fs.BoolVar(&cfg.refIn64, "refin64", false, "Reflect input bytes for CRC64")
	fs.BoolVar(&cfg.refOut64, "refout64", false, "Reflect output for CRC64")
	fs.StringVar(&cfg.xorOut64, "xorout64", "0x0000000000000000", "Custom CRC64 xor_out value (hex)")

	fs.Parse(os.Args[2:])

	var err error
	switch mode {
	case "calculate":
		err = sendCalculate(cfg)
	case "verify":
		if cfg.expectedCRC == "" {
			fmt.Println("Error: --expected is required for verify mode")
			os.Exit(1)
		}
		err = sendVerify(cfg)
	case "variants":
		err = listVariants(cfg)
	default:
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
