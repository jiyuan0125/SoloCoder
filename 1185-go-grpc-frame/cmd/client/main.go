package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"grpc-parser/pkg/api"
)

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func formatJSON(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(data)
}

func callServer(serverURL string, data []byte) (*api.ParseResponse, error) {
	req := api.ParseRequest{
		Bytes: data,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(serverURL+"/parse", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil {
			return nil, fmt.Errorf("server error: %s", errResp.Error)
		}
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result api.ParseResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

func readFromFile(filename string) ([]byte, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func readHexFromStdin() ([]byte, error) {
	fmt.Println("Enter hex data (Ctrl+D to finish):")
	reader := bufio.NewReader(os.Stdin)
	var hexBuilder strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		hexBuilder.WriteString(line)
	}

	hexStr := strings.ReplaceAll(strings.ReplaceAll(hexBuilder.String(), " ", ""), "\n", "")
	return hex.DecodeString(hexStr)
}

func printResult(result *api.ParseResponse) {
	fmt.Println("\n=== Parse Result ===")
	fmt.Printf("Success: %v\n", result.Success)
	fmt.Printf("Total Streams: %d\n", result.TotalStreams)

	if len(result.Errors) > 0 {
		fmt.Println("\nErrors:")
		for _, err := range result.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}

	for i, frame := range result.HTTP2Frames {
		fmt.Printf("\n--- HTTP/2 Frame #%d ---\n", i+1)
		fmt.Printf("  Type: %s\n", frame.Type)
		fmt.Printf("  Stream ID: %d\n", frame.StreamID)
		fmt.Printf("  Length: %d\n", frame.Length)
		fmt.Printf("  Flags: %v\n", frame.Flags)
		fmt.Printf("  END_STREAM: %v\n", frame.HasEndStream)
		fmt.Printf("  END_HEADERS: %v\n", frame.HasEndHeaders)

		if frame.Headers != nil {
			fmt.Printf("\n  Headers:\n")
			fmt.Printf("    Is GRPC: %v\n", frame.Headers.IsGRPC)
			fmt.Printf("    Is Trailer: %v\n", frame.Headers.IsTrailer)
			fmt.Println("    Header Fields:")
			for k, v := range frame.Headers.Headers {
				fmt.Printf("      %s: %s\n", k, v)
			}
		}

		if len(frame.GRPCFrames) > 0 {
			for j, gf := range frame.GRPCFrames {
				fmt.Printf("\n  gRPC Frame #%d:\n", j+1)
				fmt.Printf("    Compressed: %v\n", gf.Compressed)
				fmt.Printf("    Length: %d\n", gf.Length)
				fmt.Printf("    Raw Message (hex): %s\n", gf.RawMessage)

				if gf.ParseError != "" {
					fmt.Printf("    Parse Error: %s\n", gf.ParseError)
				}

				if len(gf.Protobuf) > 0 {
					fmt.Printf("    Protobuf Fields:\n")
					for _, pf := range gf.Protobuf {
						fmt.Printf("      - Field %d (%s): ", pf.FieldNumber, pf.WireType)
						switch v := pf.Value.(type) {
						case map[string]interface{}:
							fmt.Printf("%v\n", v)
						default:
							fmt.Printf("%v\n", v)
						}
					}
				}
			}
		}

		if frame.ParseError != "" {
			fmt.Printf("\n  Parse Error: %s\n", frame.ParseError)
		}
	}
}

func main() {
	var serverURL string
	var filename string
	var hexInput string
	var useStdin bool

	flag.StringVar(&serverURL, "server", "", "Server URL (default: http://localhost:8204)")
	flag.StringVar(&filename, "file", "", "File containing gRPC over HTTP/2 data (raw bytes or hex)")
	flag.StringVar(&hexInput, "hex", "", "Hex string to parse")
	flag.BoolVar(&useStdin, "stdin", false, "Read hex input from stdin")
	flag.Parse()

	if serverURL == "" {
		serverURL = getEnv("GRPC_PARSER_SERVER", "http://localhost:8204")
	}

	var data []byte
	var err error

	if filename != "" {
		fmt.Printf("Reading from file: %s\n", filename)
		fileData, fileErr := readFromFile(filename)
		if fileErr != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", fileErr)
			os.Exit(1)
		}

		hexStr := strings.ReplaceAll(strings.ReplaceAll(string(fileData), " ", ""), "\n", "")
		decoded, hexErr := hex.DecodeString(hexStr)
		if hexErr == nil {
			data = decoded
		} else {
			data = fileData
		}
	} else if hexInput != "" {
		hexStr := strings.ReplaceAll(strings.ReplaceAll(hexInput, " ", ""), "\n", "")
		data, err = hex.DecodeString(hexStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid hex input: %v\n", err)
			os.Exit(1)
		}
	} else if useStdin {
		data, err = readHexFromStdin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Usage:")
		fmt.Println("  grpc-client -server <url> -file <filename>")
		fmt.Println("  grpc-client -server <url> -hex <hex_string>")
		fmt.Println("  grpc-client -server <url> -stdin")
		fmt.Println("\nExamples:")
		fmt.Println("  grpc-client -file capture.bin")
		fmt.Println("  grpc-client -hex \"0000000c000000010000000004080a026869\"")
		fmt.Println("\nEnvironment variables:")
		fmt.Println("  GRPC_PARSER_SERVER - Server URL (default: http://localhost:8204)")
		os.Exit(1)
	}

	if len(data) == 0 {
		fmt.Fprintln(os.Stderr, "Error: No data to parse")
		os.Exit(1)
	}

	fmt.Printf("Parsing %d bytes...\n", len(data))
	fmt.Printf("Connecting to server: %s\n", serverURL)

	result, err := callServer(serverURL, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	printResult(result)
}
