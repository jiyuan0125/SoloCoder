package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"rleencode/pkg/api"
)

var serverURL = "http://localhost:8210"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "encode":
		runEncode(args)
	case "decode":
		runDecode(args)
	case "-h", "--help", "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: rleclient <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  encode -width <int> -height <int> -pixels <2d-array>  Encode pixel data")
	fmt.Println("  decode -data <hex-string> -width <int>               Decode RLE8 data")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -server <url>   Server URL (default: http://localhost:8210)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println(`  rleclient encode -width 4 -height 2 -pixels "[[1,1,2,2],[3,3,3,3]]"`)
	fmt.Println(`  rleclient decode -width 4 -data "020102020000040300000001"`)
}

func runEncode(args []string) {
	fs := flag.NewFlagSet("encode", flag.ExitOnError)
	width := fs.Int("width", 0, "image width")
	height := fs.Int("height", 0, "image height")
	pixelsStr := fs.String("pixels", "", "2D pixel array as JSON string, e.g. [[1,1,2],[3,4,5]]")
	server := fs.String("server", serverURL, "server URL")
	fs.Parse(args)

	if *width <= 0 || *height <= 0 {
		fmt.Fprintln(os.Stderr, "error: width and height must be positive")
		os.Exit(1)
	}
	if *pixelsStr == "" {
		fmt.Fprintln(os.Stderr, "error: pixels is required")
		os.Exit(1)
	}

	pixels, err := parsePixels(*pixelsStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing pixels: %v\n", err)
		os.Exit(1)
	}

	req := api.EncodeRequest{
		Width:  *width,
		Height: *height,
		Pixels: pixels,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(*server+"/encode", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error calling server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "server error (%d): %s\n", resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	var result api.EncodeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing response: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result.Data)
}

func runDecode(args []string) {
	fs := flag.NewFlagSet("decode", flag.ExitOnError)
	width := fs.Int("width", 0, "image width")
	dataStr := fs.String("data", "", "hex encoded RLE8 data")
	server := fs.String("server", serverURL, "server URL")
	fs.Parse(args)

	if *width <= 0 {
		fmt.Fprintln(os.Stderr, "error: width must be positive")
		os.Exit(1)
	}
	if *dataStr == "" {
		fmt.Fprintln(os.Stderr, "error: data is required")
		os.Exit(1)
	}

	req := api.DecodeRequest{
		Data:  *dataStr,
		Width: *width,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(*server+"/decode", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error calling server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "server error (%d): %s\n", resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	var result api.DecodeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing response: %v\n", err)
		os.Exit(1)
	}

	output, _ := json.Marshal(result.Pixels)
	fmt.Println(string(output))
}

func parsePixels(s string) ([][]byte, error) {
	var raw [][]int
	if err := json.Unmarshal([]byte(s), &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %v", err)
	}

	result := make([][]byte, len(raw))
	for i, row := range raw {
		result[i] = make([]byte, len(row))
		for j, val := range row {
			if val < 0 || val > 255 {
				return nil, fmt.Errorf("pixel value out of range at [%d][%d]: %d", i, j, val)
			}
			result[i][j] = byte(val)
		}
	}
	return result, nil
}

func parsePixelList(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return nil, nil
	}

	parts := strings.Split(s, ",")
	result := make([]byte, len(parts))
	for i, p := range parts {
		p = strings.TrimSpace(p)
		val, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid pixel value: %s", p)
		}
		if val < 0 || val > 255 {
			return nil, fmt.Errorf("pixel value out of range: %d", val)
		}
		result[i] = byte(val)
	}
	return result, nil
}
