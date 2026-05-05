package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"go-log-parser/logparser"
	"go-log-parser/protocol"
)

const version = "1.0.0"

var addr string

func init() {
	flag.StringVar(&addr, "addr", ":8080", "HTTP server address")
}

func main() {
	flag.Parse()

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/formats", formatsHandler)
	http.HandleFunc("/parse", parseHandler)
	http.HandleFunc("/register", registerHandler)

	fmt.Printf("Log Parser Server v%s listening on %s\n", version, addr)
	fmt.Println("Endpoints:")
	fmt.Println("  GET  /health          - Health check")
	fmt.Println("  GET  /formats         - List supported formats")
	fmt.Println("  POST /parse           - Parse logs")
	fmt.Println("  POST /register        - Register custom format")

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := protocol.HealthResponse{
		Status:  "ok",
		Version: version,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func formatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := protocol.ListFormatsResponse{
		Formats: logparser.ListBuiltInFormats(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func parseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if req.Format == "" {
		http.Error(w, "Format is required", http.StatusBadRequest)
		return
	}

	if req.LogContent == "" && req.FilePath == "" {
		http.Error(w, "LogContent or FilePath is required", http.StatusBadRequest)
		return
	}

	parser, err := logparser.NewParser(req.Format, req.Fields)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create parser: %v", err), http.StatusBadRequest)
		return
	}

	var entries []protocol.LogEntryResponse
	var errors []protocol.ParseErrorInfo
	var totalCount int

	if req.LogContent != "" {
		resultChan, err := parser.Parse(strings.NewReader(req.LogContent))
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse: %v", err), http.StatusInternalServerError)
			return
		}

		for result := range resultChan {
			if result.Entry != nil {
				entries = append(entries, logEntryToResponse(result.Entry))
				totalCount++
			}
			for _, e := range result.Errors {
				errors = append(errors, protocol.ParseErrorInfo{
					LineNum: e.LineNum,
					Reason:  e.Reason,
				})
			}
		}
	} else if req.FilePath != "" {
		file, err := os.Open(req.FilePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to open file: %v", err), http.StatusBadRequest)
			return
		}
		defer file.Close()

		resultChan, err := parser.Parse(file)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse: %v", err), http.StatusInternalServerError)
			return
		}

		for result := range resultChan {
			if result.Entry != nil {
				entries = append(entries, logEntryToResponse(result.Entry))
				totalCount++
			}
			for _, e := range result.Errors {
				errors = append(errors, protocol.ParseErrorInfo{
					LineNum: e.LineNum,
					Reason:  e.Reason,
				})
			}
		}
	}

	resp := protocol.ParseResponse{
		Success:    true,
		Entries:    entries,
		TotalCount: totalCount,
		Errors:     errors,
	}

	if len(errors) > 0 {
		resp.Message = fmt.Sprintf("Completed with %d errors", len(errors))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.FormatRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Format name is required", http.StatusBadRequest)
		return
	}

	if req.FormatString == "" {
		http.Error(w, "Format string is required", http.StatusBadRequest)
		return
	}

	logparser.RegisterFormat(req.Name, logparser.FormatDefinition{
		FormatString: req.FormatString,
		Fields:       req.Fields,
	})

	resp := protocol.FormatRegisterResponse{
		Success: true,
		Message: fmt.Sprintf("Format '%s' registered successfully", req.Name),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func logEntryToResponse(entry *logparser.LogEntry) protocol.LogEntryResponse {
	resp := protocol.LogEntryResponse{
		LineNum: entry.LineNum,
		Fields:  entry.Fields,
		Line:    entry.Line,
	}

	if entry.Timestamp != nil {
		switch ts := entry.Timestamp.(type) {
		case time.Time:
			resp.Timestamp = ts.Format(time.RFC3339Nano)
		case string:
			resp.Timestamp = ts
		case int64:
			resp.Timestamp = time.Unix(ts, 0).Format(time.RFC3339)
		case float64:
			sec := int64(ts)
			nsec := int64((ts - float64(sec)) * 1e9)
			resp.Timestamp = time.Unix(sec, nsec).Format(time.RFC3339Nano)
		}
	}

	return resp
}
