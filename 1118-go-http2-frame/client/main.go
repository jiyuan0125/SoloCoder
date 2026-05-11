package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/solocoder/http2-analyzer/api"
)

func main() {
	var serverURL string
	var hexData string
	var filePath string
	var outputJSON bool

	flag.StringVar(&serverURL, "server", "http://localhost:8304", "Server URL")
	flag.StringVar(&hexData, "hex", "", "Hex string to parse")
	flag.StringVar(&filePath, "file", "", "Binary file to parse")
	flag.BoolVar(&outputJSON, "json", false, "Output JSON instead of human readable")
	flag.Parse()

	if hexData == "" && filePath == "" {
		fmt.Fprintln(os.Stderr, "Error: must provide either -hex or -file")
		flag.Usage()
		os.Exit(1)
	}

	var req api.ParseRequest

	if hexData != "" {
		req.Data = hexData
		req.DataKind = "hex"

		if _, err := hex.DecodeString(hexData); err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid hex string: %v\n", err)
			os.Exit(1)
		}
	} else {
		req.Data = filePath
		req.DataKind = "file"

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", filePath)
			os.Exit(1)
		}
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to marshal request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/parse", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to read response: %v\n", err)
		os.Exit(1)
	}

	var parseResp api.ParseResponse
	if err := json.Unmarshal(respBody, &parseResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse response: %v\n", err)
		fmt.Fprintf(os.Stderr, "Response body: %s\n", string(respBody))
		os.Exit(1)
	}

	if outputJSON {
		fmt.Println(string(respBody))
	} else {
		printHumanReadable(&parseResp)
	}

	if parseResp.Error != "" {
		os.Exit(2)
	}
}

func printHumanReadable(resp *api.ParseResponse) {
	fmt.Printf("=== HTTP/2 Frame Analysis ===\n\n")

	if len(resp.Frames) == 0 {
		fmt.Println("No frames parsed.")
	} else {
		fmt.Printf("Parsed %d frame(s):\n\n", len(resp.Frames))

		for i, frame := range resp.Frames {
			fmt.Printf("--- Frame %d (offset: %d) ---\n", i+1, frame.Offset)
			fmt.Printf("  Type:    %s (0x%02x)\n", frame.Type, frame.TypeCode)
			fmt.Printf("  Stream:  %d\n", frame.StreamID)
			fmt.Printf("  Length:  %d bytes\n", frame.Length)
			fmt.Printf("  Flags:   ")
			for j, f := range frame.Flags {
				if j > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("%s", f)
			}
			fmt.Printf(" (0x%02x)\n", frame.FlagsCode)

			if frame.Payload != nil {
				fmt.Println("  Payload:")
				printPayload(frame.Payload, "    ")
			}
			fmt.Println()
		}
	}

	if resp.Error != "" {
		fmt.Printf("\n=== Error ===\n%s\n", resp.Error)
	}
}

func printPayload(payload interface{}, indent string) {
	switch v := payload.(type) {
	case map[string]interface{}:
		for key, val := range v {
			switch inner := val.(type) {
			case map[string]interface{}:
				fmt.Printf("%s%s:\n", indent, key)
				printPayload(inner, indent+"  ")
			case []interface{}:
				fmt.Printf("%s%s:\n", indent, key)
				for idx, item := range inner {
					if itemMap, ok := item.(map[string]interface{}); ok {
						fmt.Printf("%s  [%d]:\n", indent, idx)
						printPayload(itemMap, indent+"    ")
					} else {
						fmt.Printf("%s  [%d]: %v\n", indent, idx, item)
					}
				}
			default:
				fmt.Printf("%s%s: %v\n", indent, key, val)
			}
		}
	default:
		fmt.Printf("%s%v\n", indent, v)
	}
}
