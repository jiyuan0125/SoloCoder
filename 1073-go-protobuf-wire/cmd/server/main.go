package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"protobuf-wire/internal/common"
	"protobuf-wire/internal/core"
)

func convertParsedFields(fields []*core.ParsedField) []*common.Field {
	result := make([]*common.Field, len(fields))
	for i, f := range fields {
		result[i] = &common.Field{
			FieldNumber:  f.FieldNumber,
			WireType:     int(f.WireType),
			WireTypeName: f.WireType.String(),
			RawBytes:     f.RawBytes,
			RawByteSize:  f.RawByteSize,
			Value:        convertValue(f.Value),
			ValueType:    f.ValueType,
		}
	}
	return result
}

func convertValue(v interface{}) interface{} {
	switch val := v.(type) {
	case []*core.ParsedField:
		return convertParsedFields(val)
	default:
		return val
	}
}

func parseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	var data []byte

	if contentType == "application/json" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			sendError(w, "failed to read request body: "+err.Error(), http.StatusBadRequest)
			return
		}
		var req common.ParseRequest
		if err := json.Unmarshal(body, &req); err != nil {
			sendError(w, "failed to parse JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		data = req.Data
	} else {
		var err error
		data, err = ioutil.ReadAll(r.Body)
		if err != nil {
			sendError(w, "failed to read request body: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	parser := core.NewParser(nil)
	fields, err := parser.Parse(data, 0)
	if err != nil {
		sendError(w, "parse error: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp := common.ParseResponse{
		Success: true,
		Fields:  convertParsedFields(fields),
	}

	w.Header().Set("Content-Type", "application/json")
	jsonBytes, err := resp.ToJSON()
	if err != nil {
		sendError(w, "failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(jsonBytes)
}

func sendError(w http.ResponseWriter, errMsg string, status int) {
	resp := common.ParseResponse{
		Success: false,
		Error:   errMsg,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	jsonBytes, _ := resp.ToJSON()
	w.Write(jsonBytes)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/parse", parseHandler)

	fmt.Printf("server starting on :%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
