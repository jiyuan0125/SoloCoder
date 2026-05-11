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

	"snappy-compress/internal/api"
)

const serverURL = "http://localhost:8301"

func main() {
	decompress := flag.Bool("decompress", false, "decompress mode")
	inputFile := flag.String("file", "", "input file (defaults to stdin)")
	flag.Parse()

	var inputData []byte
	var err error

	if *inputFile != "" {
		inputData, err = os.ReadFile(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
	} else {
		inputData, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
		}
	}

	if *decompress {
		output, err := doDecompress(inputData)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Decompress error: %v\n", err)
			os.Exit(1)
		}
		os.Stdout.Write(output)
	} else {
		output, err := doCompress(inputData)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Compress error: %v\n", err)
			os.Exit(1)
		}
		os.Stdout.Write(output)
	}
}

func doCompress(data []byte) ([]byte, error) {
	req := api.CompressRequest{
		Data: base64.StdEncoding.EncodeToString(data),
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(serverURL+"/compress", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("server error: %s", errResp.Error)
	}

	var respBody api.CompressResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, err
	}

	return base64.StdEncoding.DecodeString(respBody.Data)
}

func doDecompress(data []byte) ([]byte, error) {
	req := api.DecompressRequest{
		Data: base64.StdEncoding.EncodeToString(data),
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(serverURL+"/decompress", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("server error: %s", errResp.Error)
	}

	var respBody api.DecompressResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, err
	}

	return base64.StdEncoding.DecodeString(respBody.Data)
}
