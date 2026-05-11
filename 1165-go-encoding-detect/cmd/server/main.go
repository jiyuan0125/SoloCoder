package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"encoding-detector/pkg/api"
	"encoding-detector/pkg/detector"
	"encoding-detector/pkg/encoder"
)

func main() {
	var port int
	var portStr string

	flag.IntVar(&port, "port", 0, "Server port")
	flag.StringVar(&portStr, "p", "", "Server port (short flag)")
	flag.Parse()

	actualPort := getPort(port, portStr)

	http.HandleFunc("/api/detect", handleDetect)
	http.HandleFunc("/api/convert", handleConvert)

	addr := fmt.Sprintf(":%d", actualPort)
	fmt.Printf("Server starting on port %d...\n", actualPort)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}

func getPort(flagPort int, flagPortStr string) int {
	if flagPort > 0 {
		return flagPort
	}
	if flagPortStr != "" {
		var p int
		fmt.Sscanf(flagPortStr, "%d", &p)
		if p > 0 {
			return p
		}
	}

	envPort := os.Getenv("PORT")
	if envPort != "" {
		var p int
		fmt.Sscanf(envPort, "%d", &p)
		if p > 0 {
			return p
		}
	}

	return 8200
}

func handleDetect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req api.DetectRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	result := detector.Detect(req.Content)

	resp := api.DetectResponse{
		Encoding:   result.Encoding,
		Confidence: result.Confidence,
		HasBOM:     result.HasBOM,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req api.ConvertRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	toEncoding := strings.ToLower(req.ToEncoding)
	if toEncoding == "" {
		toEncoding = "utf-8"
	}

	fromEncoding := req.FromEncoding
	if fromEncoding == "" {
		detectResult := detector.Detect(req.Content)
		fromEncoding = detectResult.Encoding
	}

	converted, err := encoder.Convert(req.Content, fromEncoding, toEncoding)
	if err != nil {
		resp := api.ConvertResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := api.ConvertResponse{
		Content: converted,
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
