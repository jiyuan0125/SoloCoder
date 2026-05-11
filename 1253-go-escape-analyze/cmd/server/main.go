package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"

	"go-escape-analyze/internal/parser"
	"go-escape-analyze/pkg/api"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "", "Server port (default: 8500)")
	flag.Parse()

	if port == "" {
		port = os.Getenv("ESCAPE_ANALYZE_PORT")
	}
	if port == "" {
		port = "8500"
	}

	http.HandleFunc("/analyze", analyzeHandler)

	log.Printf("Escape Analysis Server listening on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeResponse(w, http.StatusMethodNotAllowed, api.AnalyzeResponse{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	var req api.AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeResponse(w, http.StatusBadRequest, api.AnalyzeResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	if req.OutputText == "" {
		writeResponse(w, http.StatusBadRequest, api.AnalyzeResponse{
			Success: false,
			Error:   "output_text is required",
		})
		return
	}

	report := parser.ParseOutput(req.OutputText, req.Filter)

	writeResponse(w, http.StatusOK, api.AnalyzeResponse{
		Success: true,
		Report:  report,
	})
}

func writeResponse(w http.ResponseWriter, statusCode int, resp api.AnalyzeResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}
