package main

import (
	"encoding/json"
	"io"
	"net/http"

	"file-type-detector/common"
	"file-type-detector/filedetector"
)

func DetectHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, "Failed to read request body")
		return
	}

	var req common.DetectRequest
	if err := json.Unmarshal(body, &req); err != nil {
		respondWithError(w, "Invalid JSON format")
		return
	}

	var mimeType string
	var detectErr error

	if len(req.Data) > 0 {
		mimeType = filedetector.DetectFromBytes(req.Data)
	} else if req.Path != "" {
		mimeType, detectErr = filedetector.DetectFromPath(req.Path)
	} else {
		respondWithError(w, "No data or path provided")
		return
	}

	resp := common.DetectResponse{
		MimeType: mimeType,
	}

	if detectErr != nil {
		resp.Error = detectErr.Error()
	}

	json.NewEncoder(w).Encode(resp)
}

func respondWithError(w http.ResponseWriter, msg string) {
	resp := common.DetectResponse{
		MimeType: "unknown",
		Error:    msg,
	}
	json.NewEncoder(w).Encode(resp)
}
