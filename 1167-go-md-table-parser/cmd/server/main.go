package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"mdtableparser/pkg/common"
	"mdtableparser/pkg/tableparser"
)

func getPort() string {
	port := "8080"

	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	flagPort := flag.String("port", "", "Server port (default: 8080, can also be set via PORT env variable)")
	flag.Parse()

	if *flagPort != "" {
		port = *flagPort
	}

	return port
}

func handleExtractTables(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req common.ExtractTablesRequest
	if err := json.Unmarshal(body, &req); err != nil {
		sendError(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	tables := tableparser.ExtractTables(req.Markdown)
	tableInfos := make([]common.TableInfo, len(tables))

	for i, table := range tables {
		alignments := make([]string, len(table.Alignments))
		for j, alignment := range table.Alignments {
			alignments[j] = alignment.String()
		}
		tableInfos[i] = common.TableInfo{
			Headers:    table.Headers,
			Rows:       table.Rows,
			Alignments: alignments,
		}
	}

	resp := common.ExtractTablesResponse{
		Tables: tableInfos,
	}

	sendJSON(w, resp, http.StatusOK)
}

func handleMarkdownToCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req common.MarkdownToCSVRequest
	if err := json.Unmarshal(body, &req); err != nil {
		sendError(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	tables := tableparser.ExtractTables(req.Markdown)
	if len(tables) == 0 {
		sendJSON(w, common.MarkdownToCSVResponse{
			CSVContent: "",
			TableCount: 0,
		}, http.StatusOK)
		return
	}

	var csvParts []string
	for _, table := range tables {
		csvText, err := table.ToCSV()
		if err != nil {
			sendError(w, "Failed to convert table to CSV: "+err.Error(), http.StatusInternalServerError)
			return
		}
		csvParts = append(csvParts, csvText)
	}

	resp := common.MarkdownToCSVResponse{
		CSVContent: strings.Join(csvParts, "\n\n"),
		TableCount: len(tables),
	}

	sendJSON(w, resp, http.StatusOK)
}

func handleCSVToMarkdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req common.CSVToMarkdownRequest
	if err := json.Unmarshal(body, &req); err != nil {
		sendError(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	table, err := tableparser.ParseCSVWithOptions(req.CSVContent, req.HasHeader)
	if err != nil {
		sendError(w, "Failed to parse CSV: "+err.Error(), http.StatusBadRequest)
		return
	}

	table.InferAlignments()
	markdown := table.ToMarkdown()

	resp := common.CSVToMarkdownResponse{
		Markdown: markdown,
	}

	sendJSON(w, resp, http.StatusOK)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	errorResp := common.ErrorResponse{
		Error: message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResp)
}

func sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func main() {
	port := getPort()

	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/api/extract-tables", handleExtractTables)
	http.HandleFunc("/api/markdown-to-csv", handleMarkdownToCSV)
	http.HandleFunc("/api/csv-to-markdown", handleCSVToMarkdown)

	log.Printf("Server starting on port %s...", port)
	log.Printf("Endpoints:")
	log.Printf("  GET  /health")
	log.Printf("  POST /api/extract-tables")
	log.Printf("  POST /api/markdown-to-csv")
	log.Printf("  POST /api/csv-to-markdown")

	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
