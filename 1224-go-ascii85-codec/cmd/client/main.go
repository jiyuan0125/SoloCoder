package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"ascii85-codec/api"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := flag.String("server", "http://localhost:8305", "Server URL (default: http://localhost:8305)")
	flag.Usage = printUsage

	cmd := os.Args[1]
	os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
	flag.Parse()

	var err error
	switch cmd {
	case "encode":
		err = handleEncode(*serverURL)
	case "decode":
		err = handleDecode(*serverURL)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleEncode(serverURL string) error {
	mode := flag.String("mode", "adobe", "Encoding mode: adobe or btoa (default: adobe)")
	flag.Parse()

	data, err := readInput()
	if err != nil {
		return err
	}

	req := api.EncodeRequest{
		Data: string(data),
		Mode: *mode,
	}

	resp, err := postJSON(serverURL+"/encode", req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return readError(resp)
	}

	var result api.EncodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Println(result.Encoded)
	return nil
}

func handleDecode(serverURL string) error {
	flag.Parse()

	data, err := readInput()
	if err != nil {
		return err
	}

	req := api.DecodeRequest{
		Encoded: string(data),
	}

	resp, err := postJSON(serverURL+"/decode", req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return readError(resp)
	}

	var result api.DecodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Print(result.Decoded)
	return nil
}

func readInput() ([]byte, error) {
	if flag.NArg() > 0 {
		return []byte(flag.Arg(0)), nil
	}

	return io.ReadAll(os.Stdin)
}

func postJSON(url string, req interface{}) (*http.Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	return http.Post(url, "application/json", bytes.NewReader(body))
}

func readError(resp *http.Response) error {
	var errResp api.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	return fmt.Errorf("%s", errResp.Error)
}

func printUsage() {
	fmt.Println(`Usage: ascii85-client [flags] <command> [arguments]

Commands:
  encode    Encode data to Ascii85
  decode    Decode Ascii85 data

Flags:
  -server string   Server URL (default "http://localhost:8305")

For encode:
  -mode string     Encoding mode: adobe or btoa (default "adobe")

Examples:
  echo "Hello" | ascii85-client encode
  ascii85-client encode "Hello World"
  ascii85-client encode -mode btoa "Hello"
  echo "<~87cURD]j7BES~>" | ascii85-client decode
  ascii85-client decode "<~87cURD]j7BES~>"`)
}
