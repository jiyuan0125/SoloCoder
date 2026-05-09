package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"bencode-parser/api"
	"bencode-parser/bencode"
)

func main() {
	http.HandleFunc("/decode", handleDecode)
	http.HandleFunc("/encode", handleEncode)
	http.HandleFunc("/info", handleInfo)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleDecode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.DecodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := api.DecodeResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	result, err := bencode.Decode([]byte(req.Data))
	if err != nil {
		resp := api.DecodeResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := api.DecodeResponse{
		Success: true,
		Result:  convertResult(result),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleEncode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.EncodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := api.EncodeResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	converted := convertForEncode(req.Data)
	result, err := bencode.Encode(converted)
	if err != nil {
		resp := api.EncodeResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := api.EncodeResponse{
		Success: true,
		Result:  string(result),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.InfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := api.InfoResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	info, err := bencode.ParseInfo([]byte(req.Data))
	if err != nil {
		resp := api.InfoResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := api.InfoResponse{
		Success:       true,
		OuterType:     info.OuterType,
		DictKeyCount:  info.DictKeyCount,
		ListElemCount: info.ListElemCount,
		MaxDepth:      info.MaxDepth,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func convertResult(v interface{}) interface{} {
	switch val := v.(type) {
	case []byte:
		return string(val)
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, elem := range val {
			result[i] = convertResult(elem)
		}
		return result
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, v := range val {
			result[k] = convertResult(v)
		}
		return result
	default:
		return v
	}
}

func convertForEncode(v interface{}) interface{} {
	switch val := v.(type) {
	case string:
		return []byte(val)
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, elem := range val {
			result[i] = convertForEncode(elem)
		}
		return result
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, v := range val {
			result[k] = convertForEncode(v)
		}
		return result
	default:
		return v
	}
}
