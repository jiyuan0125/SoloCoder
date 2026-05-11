package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"goroutinelab/analyzer"
	"goroutinelab/api"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "", "Port to listen on (e.g., :8080)")
	flag.Parse()

	if port == "" {
		port = os.Getenv("GOROUTINE_LAB_PORT")
	}
	if port == "" {
		port = ":8080"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	analyzer := analyzer.New()

	http.HandleFunc("/analyze", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req api.AnalyzeRequest
		body, err := io.ReadAll(r.Body)
		if err != nil {
			sendResponse(w, api.AnalyzeResponse{Success: false, Error: err.Error()})
			return
		}
		defer r.Body.Close()

		var contentType = r.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/json") {
			if err := json.Unmarshal(body, &req); err != nil {
				sendResponse(w, api.AnalyzeResponse{Success: false, Error: "Invalid JSON: " + err.Error()})
				return
			}
		} else {
			req.StackTrace = string(body)
		}

		report := analyzer.Analyze(req.StackTrace, req.Options)
		sendResponse(w, api.AnalyzeResponse{Success: true, Report: *report})
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	fmt.Printf("Goroutine Lab server listening on %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}

func sendResponse(w http.ResponseWriter, resp api.AnalyzeResponse) {
	w.Header().Set("Content-Type", "application/json")
	if !resp.Success {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(resp)
}

func intEnv(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
