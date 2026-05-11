package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"huffman-codec/common"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const serverURL = "http://localhost:8400"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "encode":
		handleEncode()
	case "decode":
		handleDecode()
	case "frequencies":
		handleFrequencies()
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Huffman Codec Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client encode <text|--file <path>> [--output <path>]")
	fmt.Println("  client decode --input <path> [--output <path>]")
	fmt.Println("  client frequencies")
	fmt.Println("  client help")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  encode       Encode text using Huffman coding")
	fmt.Println("  decode       Decode previously encoded data")
	fmt.Println("  frequencies  Get current frequency table from server")
	fmt.Println("  help         Show this help message")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --file <path>     Read input text from file")
	fmt.Println("  --input <path>    Read encoded data from file (for decode)")
	fmt.Println("  --output <path>   Write output to file")
}

func handleEncode() {
	encodeCmd := flag.NewFlagSet("encode", flag.ExitOnError)
	filePath := encodeCmd.String("file", "", "Read input text from file")
	outputPath := encodeCmd.String("output", "", "Write output to file")

	if err := encodeCmd.Parse(os.Args[2:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	var text string
	if *filePath != "" {
		content, err := os.ReadFile(*filePath)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}
		text = string(content)
	} else if encodeCmd.NArg() > 0 {
		text = encodeCmd.Arg(0)
	} else {
		fmt.Println("Error: No input provided. Use either text argument or --file flag.")
		os.Exit(1)
	}

	reqBody, _ := json.Marshal(common.EncodeRequest{Text: text})
	resp, err := http.Post(serverURL+"/encode", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Server error: %s\n", string(body))
		os.Exit(1)
	}

	var encodeResp common.EncodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&encodeResp); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		os.Exit(1)
	}

	if *outputPath != "" {
		saveEncodeResult(*outputPath, &encodeResp)
		fmt.Printf("Encoded data saved to: %s\n", *outputPath)
	}

	fmt.Println()
	fmt.Println("=== Encoding Complete ===")
	fmt.Printf("Original size:  %d bytes\n", encodeResp.OriginalSize)
	fmt.Printf("Encoded size:   %d bytes\n", encodeResp.EncodedSize)
	fmt.Printf("Compression ratio: %.2f%%\n", float64(encodeResp.EncodedSize)/float64(encodeResp.OriginalSize)*100)
	fmt.Println()
	fmt.Println("=== Frequency Table ===")
	for char, freq := range encodeResp.Frequencies {
		displayChar := char
		if char == "\n" {
			displayChar = "\\n"
		} else if char == "\t" {
			displayChar = "\\t"
		} else if char == " " {
			displayChar = "<space>"
		}
		fmt.Printf("  %q: %d\n", displayChar, freq)
	}
}

func handleDecode() {
	decodeCmd := flag.NewFlagSet("decode", flag.ExitOnError)
	inputPath := decodeCmd.String("input", "", "Read encoded data from file (required)")
	outputPath := decodeCmd.String("output", "", "Write output to file")

	if err := decodeCmd.Parse(os.Args[2:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if *inputPath == "" {
		fmt.Println("Error: --input flag is required for decode command")
		os.Exit(1)
	}

	data, err := os.ReadFile(*inputPath)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	var decodeReq common.DecodeRequest
	if err := json.Unmarshal(data, &decodeReq); err != nil {
		fmt.Printf("Error parsing encoded data: %v\n", err)
		os.Exit(1)
	}

	reqBody, _ := json.Marshal(decodeReq)
	resp, err := http.Post(serverURL+"/decode", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Server error: %s\n", string(body))
		os.Exit(1)
	}

	var decodeResp common.DecodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&decodeResp); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		os.Exit(1)
	}

	if *outputPath != "" {
		if err := os.WriteFile(*outputPath, []byte(decodeResp.Text), 0644); err != nil {
			fmt.Printf("Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Decoded text saved to: %s\n", *outputPath)
	}

	fmt.Println()
	fmt.Println("=== Decoding Complete ===")
	fmt.Printf("Encoded size:   %d bytes\n", decodeResp.EncodedSize)
	fmt.Printf("Original size:  %d bytes\n", decodeResp.OriginalSize)
	if *outputPath == "" {
		fmt.Println()
		fmt.Println("=== Decoded Text ===")
		fmt.Println(decodeResp.Text)
	}
}

func handleFrequencies() {
	resp, err := http.Get(serverURL + "/frequencies")
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Server error: %s\n", string(body))
		os.Exit(1)
	}

	var freqResp common.FrequencyResponse
	if err := json.NewDecoder(resp.Body).Decode(&freqResp); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Current Frequency Table ===")
	if len(freqResp.Frequencies) == 0 {
		fmt.Println("  (No data yet)")
		return
	}

	for char, freq := range freqResp.Frequencies {
		displayChar := char
		if char == "\n" {
			displayChar = "\\n"
		} else if char == "\t" {
			displayChar = "\\t"
		} else if char == " " {
			displayChar = "<space>"
		}
		fmt.Printf("  %q: %d\n", displayChar, freq)
	}
}

func saveEncodeResult(outputPath string, resp *common.EncodeResponse) {
	decodeReq := common.DecodeRequest{
		Data:        resp.Data,
		PaddingBits: resp.PaddingBits,
		TreeData:    resp.TreeData,
	}

	data, err := json.MarshalIndent(decodeReq, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling result: %v\n", err)
		os.Exit(1)
	}

	dir := filepath.Dir(outputPath)
	if dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory: %v\n", err)
			os.Exit(1)
		}
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}
}
