package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"varint-codec/api"
	"varint-codec/codec"
)

func main() {
	http.HandleFunc("/encode", handleEncode)
	http.HandleFunc("/decode", handleDecode)
	http.HandleFunc("/validate", handleValidate)
	http.HandleFunc("/benchmark", handleBenchmark)

	fmt.Println("Server starting on :8500")
	http.ListenAndServe(":8500", nil)
}

func handleEncode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		respondError(w, "method not allowed")
		return
	}

	var req api.EncodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "invalid request body")
		return
	}

	var encoded []byte
	switch req.NumberType {
	case codec.TypeInt32:
		values := make([]int32, len(req.Numbers))
		for i, v := range req.Numbers {
			values[i] = int32(v)
		}
		encoded = codec.EncodeBatchInt32(values, req.Mode)
	case codec.TypeInt64:
		values := make([]int64, len(req.Numbers))
		for i, v := range req.Numbers {
			values[i] = v
		}
		encoded = codec.EncodeBatchInt64(values, req.Mode)
	case codec.TypeUint32:
		values := make([]uint32, len(req.Numbers))
		for i, v := range req.Numbers {
			values[i] = uint32(v)
		}
		encoded = codec.EncodeBatchUint32(values, req.Mode)
	case codec.TypeUint64:
		fallthrough
	default:
		values := make([]uint64, len(req.Numbers))
		for i, v := range req.Numbers {
			values[i] = uint64(v)
		}
		encoded = codec.EncodeBatchUint64(values, req.Mode)
	}

	resp := api.EncodeResponse{
		Success: true,
		Data:    base64.StdEncoding.EncodeToString(encoded),
	}
	json.NewEncoder(w).Encode(resp)
}

func handleDecode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		respondError(w, "method not allowed")
		return
	}

	var req api.DecodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "invalid request body")
		return
	}

	data, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		respondError(w, "invalid base64 data")
		return
	}

	var numbers []int64
	switch req.NumberType {
	case codec.TypeInt32:
		values, err := codec.DecodeBatchInt32(data, req.Mode)
		if err != nil {
			respondError(w, err.Error())
			return
		}
		numbers = make([]int64, len(values))
		for i, v := range values {
			numbers[i] = int64(v)
		}
	case codec.TypeInt64:
		values, err := codec.DecodeBatchInt64(data, req.Mode)
		if err != nil {
			respondError(w, err.Error())
			return
		}
		numbers = values
	case codec.TypeUint32:
		values, err := codec.DecodeBatchUint32(data, req.Mode)
		if err != nil {
			respondError(w, err.Error())
			return
		}
		numbers = make([]int64, len(values))
		for i, v := range values {
			numbers[i] = int64(v)
		}
	case codec.TypeUint64:
		fallthrough
	default:
		values, err := codec.DecodeBatchUint64(data, req.Mode)
		if err != nil {
			respondError(w, err.Error())
			return
		}
		numbers = make([]int64, len(values))
		for i, v := range values {
			numbers[i] = int64(v)
		}
	}

	resp := api.DecodeResponse{
		Success: true,
		Numbers: numbers,
	}
	json.NewEncoder(w).Encode(resp)
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		respondError(w, "method not allowed")
		return
	}

	var req api.ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "invalid request body")
		return
	}

	data, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		respondError(w, "invalid base64 data")
		return
	}

	var valid bool
	var msg string
	if req.Mode == codec.ModeLEB128 {
		valid, msg = codec.ValidateLEB128(data, req.NumberType)
	} else {
		valid, msg = codec.ValidateVarint(data, req.NumberType)
	}

	resp := api.ValidateResponse{
		Success: true,
		Valid:   valid,
		Message: msg,
	}
	json.NewEncoder(w).Encode(resp)
}

func handleBenchmark(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		respondError(w, "method not allowed")
		return
	}

	var req api.BenchmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "invalid request body")
		return
	}

	if req.Size <= 0 {
		respondError(w, "size must be positive")
		return
	}

	testData := make([]int64, req.Size)
	for i := range testData {
		testData[i] = int64(i * 1234567)
	}

	startEncode := time.Now()
	encoded := codec.EncodeBatchInt64(testData, codec.ModeVarint)
	encodeTime := time.Since(startEncode)

	startDecode := time.Now()
	_, _ = codec.DecodeBatchInt64(encoded, codec.ModeVarint)
	decodeTime := time.Since(startDecode)

	resp := api.BenchmarkResponse{
		Success:      true,
		Size:         req.Size,
		EncodeTimeNs: encodeTime.Nanoseconds(),
		DecodeTimeNs: decodeTime.Nanoseconds(),
		EncodePerSec: float64(req.Size) / encodeTime.Seconds(),
		DecodePerSec: float64(req.Size) / decodeTime.Seconds(),
	}
	json.NewEncoder(w).Encode(resp)
}

func respondError(w http.ResponseWriter, msg string) {
	resp := map[string]interface{}{
		"success": false,
		"error":   msg,
	}
	json.NewEncoder(w).Encode(resp)
}
