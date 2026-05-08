package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/solocoder/base64codec/base64"
	"github.com/solocoder/base64codec/pkg"
)

func main() {
	http.HandleFunc("/api/encode", handleEncode)
	http.HandleFunc("/api/decode", handleDecode)

	port := ":8080"
	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func handleEncode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	var req pkg.EncodeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	if req.Mode == "" {
		req.Mode = pkg.ModeStandard
	}

	opts := base64.NewEncodeOptions(base64.Mode(req.Mode))
	opts.Padding = req.Padding
	if req.MIMELineWidth > 0 {
		opts.MIMELineWidth = req.MIMELineWidth
	}

	data := []byte(req.Data)
	result, err := base64.Encode(data, opts)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := pkg.EncodeResponse{
		Success: true,
		Result:  result,
	}
	sendJSONResponse(w, http.StatusOK, resp)
}

func handleDecode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	var req pkg.DecodeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	if req.Mode == "" {
		req.Mode = pkg.ModeStandard
	}

	opts := base64.NewDecodeOptions(base64.Mode(req.Mode))
	result, err := base64.Decode(req.Data, opts)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := pkg.DecodeResponse{
		Success: true,
		Result:  string(result),
	}
	sendJSONResponse(w, http.StatusOK, resp)
}

func sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	resp := map[string]interface{}{
		"success": false,
		"error":   message,
	}
	sendJSONResponse(w, statusCode, resp)
}

func sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Printf("Failed to encode response: %v\n", err)
	}
}
