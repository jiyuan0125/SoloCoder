package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"varint-codec/api"
	"varint-codec/codec"
)

func main() {
	mode := flag.String("mode", "varint", "Encoding mode: varint, leb128, zigzag")
	numType := flag.String("type", "int64", "Number type: int32, int64, uint32, uint64")
	action := flag.String("action", "", "Action: encode, decode, validate, benchmark")
	input := flag.String("input", "", "Input file path")
	output := flag.String("output", "", "Output file path")
	server := flag.String("server", "", "Server URL for remote mode (e.g., http://localhost:8080)")
	size := flag.Int("size", 10000, "Benchmark size")
	flag.Parse()

	if *action == "" {
		fmt.Println("Error: action is required")
		printUsage()
		os.Exit(1)
	}

	encodeMode := parseMode(*mode)
	numberType := parseNumberType(*numType)

	switch *action {
	case "encode":
		if *server != "" {
			encodeRemote(*input, *output, *server, encodeMode, numberType)
		} else {
			encodeLocal(*input, *output, encodeMode, numberType)
		}
	case "decode":
		if *server != "" {
			decodeRemote(*input, *output, *server, encodeMode, numberType)
		} else {
			decodeLocal(*input, *output, encodeMode, numberType)
		}
	case "validate":
		if *server != "" {
			validateRemote(*input, *server, encodeMode, numberType)
		} else {
			validateLocal(*input, encodeMode, numberType)
		}
	case "benchmark":
		if *server != "" {
			benchmarkRemote(*server, *size)
		} else {
			fmt.Println("Benchmark requires server URL")
			os.Exit(1)
		}
	default:
		fmt.Printf("Error: unknown action %s\n", *action)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Usage:
  client -action encode -input numbers.txt -output encoded.bin
  client -action decode -input encoded.bin -output numbers.txt
  client -action validate -input encoded.bin
  client -action benchmark -server http://localhost:8080 -size 100000
  
Options:
  -action     encode|decode|validate|benchmark
  -mode       varint|leb128|zigzag (default: varint)
  -type       int32|int64|uint32|uint64 (default: int64)
  -input      Input file path
  -output     Output file path
  -server     Server URL for remote operations
  -size       Benchmark size (default: 10000)`)
}

func parseMode(m string) codec.EncodeMode {
	switch strings.ToLower(m) {
	case "leb128":
		return codec.ModeLEB128
	case "zigzag":
		return codec.ModeZigZag
	default:
		return codec.ModeVarint
	}
}

func parseNumberType(t string) codec.NumberType {
	switch strings.ToLower(t) {
	case "int32":
		return codec.TypeInt32
	case "uint32":
		return codec.TypeUint32
	case "uint64":
		return codec.TypeUint64
	default:
		return codec.TypeInt64
	}
}

func readNumbers(file string) ([]int64, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var numbers []int64
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		n, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", line)
		}
		numbers = append(numbers, n)
	}
	return numbers, scanner.Err()
}

func writeNumbers(file string, numbers []int64) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, n := range numbers {
		fmt.Fprintln(w, n)
	}
	return w.Flush()
}

func encodeLocal(input, output string, mode codec.EncodeMode, numType codec.NumberType) {
	numbers, err := readNumbers(input)
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		os.Exit(1)
	}

	var encoded []byte
	switch numType {
	case codec.TypeInt32:
		vals := make([]int32, len(numbers))
		for i, n := range numbers {
			vals[i] = int32(n)
		}
		encoded = codec.EncodeBatchInt32(vals, mode)
	case codec.TypeInt64:
		vals := make([]int64, len(numbers))
		for i, n := range numbers {
			vals[i] = n
		}
		encoded = codec.EncodeBatchInt64(vals, mode)
	case codec.TypeUint32:
		vals := make([]uint32, len(numbers))
		for i, n := range numbers {
			vals[i] = uint32(n)
		}
		encoded = codec.EncodeBatchUint32(vals, mode)
	case codec.TypeUint64:
		vals := make([]uint64, len(numbers))
		for i, n := range numbers {
			vals[i] = uint64(n)
		}
		encoded = codec.EncodeBatchUint64(vals, mode)
	}

	if err := os.WriteFile(output, encoded, 0644); err != nil {
		fmt.Printf("Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Encoded %d numbers, %d bytes written\n", len(numbers), len(encoded))
}

func decodeLocal(input, output string, mode codec.EncodeMode, numType codec.NumberType) {
	data, err := os.ReadFile(input)
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		os.Exit(1)
	}

	var numbers []int64
	switch numType {
	case codec.TypeInt32:
		vals, err := codec.DecodeBatchInt32(data, mode)
		if err != nil {
			fmt.Printf("Decode error: %v\n", err)
			os.Exit(1)
		}
		numbers = make([]int64, len(vals))
		for i, v := range vals {
			numbers[i] = int64(v)
		}
	case codec.TypeInt64:
		vals, err := codec.DecodeBatchInt64(data, mode)
		if err != nil {
			fmt.Printf("Decode error: %v\n", err)
			os.Exit(1)
		}
		numbers = vals
	case codec.TypeUint32:
		vals, err := codec.DecodeBatchUint32(data, mode)
		if err != nil {
			fmt.Printf("Decode error: %v\n", err)
			os.Exit(1)
		}
		numbers = make([]int64, len(vals))
		for i, v := range vals {
			numbers[i] = int64(v)
		}
	case codec.TypeUint64:
		vals, err := codec.DecodeBatchUint64(data, mode)
		if err != nil {
			fmt.Printf("Decode error: %v\n", err)
			os.Exit(1)
		}
		numbers = make([]int64, len(vals))
		for i, v := range vals {
			numbers[i] = int64(v)
		}
	}

	if err := writeNumbers(output, numbers); err != nil {
		fmt.Printf("Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Decoded %d numbers\n", len(numbers))
}

func validateLocal(input string, mode codec.EncodeMode, numType codec.NumberType) {
	data, err := os.ReadFile(input)
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		os.Exit(1)
	}

	var valid bool
	var msg string
	if mode == codec.ModeLEB128 {
		valid, msg = codec.ValidateLEB128(data, numType)
	} else {
		valid, msg = codec.ValidateVarint(data, numType)
	}

	if valid {
		fmt.Printf("Valid encoding: %s\n", msg)
	} else {
		fmt.Printf("Invalid encoding: %s\n", msg)
		os.Exit(1)
	}
}

func postJSON(url string, req interface{}, resp interface{}) error {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpResp, err := http.Post(url, "application/json", strings.NewReader(string(reqBody)))
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, resp)
}

func encodeRemote(input, output, server string, mode codec.EncodeMode, numType codec.NumberType) {
	numbers, err := readNumbers(input)
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		os.Exit(1)
	}

	req := api.EncodeRequest{
		Numbers:    numbers,
		Mode:       mode,
		NumberType: numType,
	}

	var resp api.EncodeResponse
	if err := postJSON(server+"/encode", req, &resp); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		os.Exit(1)
	}

	data, err := base64.StdEncoding.DecodeString(resp.Data)
	if err != nil {
		fmt.Printf("Base64 decode error: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(output, data, 0644); err != nil {
		fmt.Printf("Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Encoded %d numbers, %d bytes written\n", len(numbers), len(data))
}

func decodeRemote(input, output, server string, mode codec.EncodeMode, numType codec.NumberType) {
	rawData, err := os.ReadFile(input)
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		os.Exit(1)
	}

	req := api.DecodeRequest{
		Data:       base64.StdEncoding.EncodeToString(rawData),
		Mode:       mode,
		NumberType: numType,
	}

	var resp api.DecodeResponse
	if err := postJSON(server+"/decode", req, &resp); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		os.Exit(1)
	}

	if err := writeNumbers(output, resp.Numbers); err != nil {
		fmt.Printf("Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Decoded %d numbers\n", len(resp.Numbers))
}

func validateRemote(input, server string, mode codec.EncodeMode, numType codec.NumberType) {
	rawData, err := os.ReadFile(input)
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		os.Exit(1)
	}

	req := api.ValidateRequest{
		Data:       base64.StdEncoding.EncodeToString(rawData),
		Mode:       mode,
		NumberType: numType,
	}

	var resp api.ValidateResponse
	if err := postJSON(server+"/validate", req, &resp); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		os.Exit(1)
	}

	if resp.Valid {
		fmt.Printf("Valid encoding: %s\n", resp.Message)
	} else {
		fmt.Printf("Invalid encoding: %s\n", resp.Message)
		os.Exit(1)
	}
}

func benchmarkRemote(server string, size int) {
	req := api.BenchmarkRequest{Size: size}
	var resp api.BenchmarkResponse
	if err := postJSON(server+"/benchmark", req, &resp); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Printf("Benchmark Results (size=%d):\n", resp.Size)
	fmt.Printf("  Encode: %.2f ms (%.0f numbers/sec)\n", float64(resp.EncodeTimeNs)/1e6, resp.EncodePerSec)
	fmt.Printf("  Decode: %.2f ms (%.0f numbers/sec)\n", float64(resp.DecodeTimeNs)/1e6, resp.DecodePerSec)
}
