package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"yenc-codec/common"
	"yenc-codec/yenc"
)

const (
	defaultPort = 8080
	envPortKey  = "YENC_PORT"
)

func getPort() int {
	if envPort := os.Getenv(envPortKey); envPort != "" {
		if port, err := strconv.Atoi(envPort); err == nil && port > 0 && port < 65536 {
			return port
		}
	}

	if len(os.Args) >= 2 {
		if port, err := strconv.Atoi(os.Args[1]); err == nil && port > 0 && port < 65536 {
			return port
		}
	}

	return defaultPort
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func encodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, &common.ErrorResponse{
			Success: false,
			Message: "method not allowed",
		})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &common.ErrorResponse{
			Success: false,
			Message: "failed to read request body",
		})
		return
	}

	var req common.EncodeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, &common.ErrorResponse{
			Success: false,
			Message: "invalid JSON request",
		})
		return
	}

	data, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &common.ErrorResponse{
			Success: false,
			Message: "invalid base64 data",
		})
		return
	}

	var opts *yenc.Options
	if req.LineSize != nil && *req.LineSize > 0 {
		opts = &yenc.Options{LineSize: *req.LineSize}
	}

	result, err := yenc.EncodeWithOptions(data, opts)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, &common.ErrorResponse{
			Success: false,
			Message: fmt.Sprintf("encode error: %v", err),
		})
		return
	}

	base64Encoded := base64.StdEncoding.EncodeToString([]byte(result.Text))

	writeJSON(w, http.StatusOK, &common.EncodeResponse{
		Success: true,
		Encoded: base64Encoded,
	})
}

func decodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, &common.ErrorResponse{
			Success: false,
			Message: "method not allowed",
		})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &common.ErrorResponse{
			Success: false,
			Message: "failed to read request body",
		})
		return
	}

	var req common.DecodeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, &common.ErrorResponse{
			Success: false,
			Message: "invalid JSON request",
		})
		return
	}

	yencData, err := base64.StdEncoding.DecodeString(req.Encoded)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &common.ErrorResponse{
			Success: false,
			Message: "invalid base64 encoded yEnc data",
		})
		return
	}

	decoded, err := yenc.Decode(string(yencData))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &common.ErrorResponse{
			Success: false,
			Message: fmt.Sprintf("decode error: %v", err),
		})
		return
	}

	base64Data := base64.StdEncoding.EncodeToString(decoded)

	writeJSON(w, http.StatusOK, &common.DecodeResponse{
		Success: true,
		Data:    base64Data,
	})
}

func main() {
	port := getPort()

	http.HandleFunc("/encode", encodeHandler)
	http.HandleFunc("/decode", decodeHandler)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("yEnc server listening on %s\n", addr)
	fmt.Printf("  POST /encode - encode base64 data to yEnc\n")
	fmt.Printf("  POST /decode - decode yEnc to base64 data\n")

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
