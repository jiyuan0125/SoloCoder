package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"isbn-lookup/api"
	"isbn-lookup/core"
)

var store = core.NewBookStore()

func main() {
	addr := flag.String("addr", ":8602", "server address")
	flag.Parse()

	http.HandleFunc("/api/query", handleQuery)
	http.HandleFunc("/api/load", handleLoad)
	http.HandleFunc("/api/add", handleAdd)

	log.Printf("Server starting on %s", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.ErrorResponse{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	var req api.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	query := &core.Query{
		ISBN:      strings.TrimSpace(req.ISBN),
		Title:     strings.TrimSpace(req.Title),
		Publisher: strings.TrimSpace(req.Publisher),
	}

	if req.Authors != "" {
		parts := strings.Split(req.Authors, ";")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				query.Authors = append(query.Authors, p)
			}
		}
	}

	pageReq := &core.PageRequest{
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	result := store.Query(query, pageReq)

	books := make([]api.BookDTO, 0, len(result.Books))
	for _, b := range result.Books {
		books = append(books, api.BookDTO{
			ISBN:         b.ISBN,
			Title:        b.Title,
			Authors:      b.Authors,
			Publisher:    b.Publisher,
			Year:         b.Year,
			CategoryCode: b.CategoryCode,
		})
	}

	resp := api.QueryResponse{
		Success:     true,
		Total:       result.Total,
		TotalPages:  result.TotalPages,
		CurrentPage: result.CurrentPage,
		PageSize:    result.PageSize,
		Books:       books,
	}

	if result.Total == 0 {
		resp.Message = "No matching records found"
	}

	writeJSON(w, http.StatusOK, resp)
}

func handleLoad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.ErrorResponse{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	var req api.LoadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	file, err := os.Open(req.FilePath)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
			Success: false,
			Message: "Failed to open file: " + err.Error(),
		})
		return
	}
	defer file.Close()

	books, result, err := core.LoadFromReader(file)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, api.ErrorResponse{
			Success: false,
			Message: "Failed to load file: " + err.Error(),
		})
		return
	}

	for _, warn := range result.Warnings {
		log.Printf("WARNING: %s", warn)
	}

	store.ReplaceAll(books)

	invalidRecords := make([]api.InvalidRecordDTO, 0, len(result.InvalidRecords))
	for _, ir := range result.InvalidRecords {
		invalidRecords = append(invalidRecords, api.InvalidRecordDTO{
			LineNumber: ir.LineNumber,
			Line:       ir.Line,
			Error:      ir.Error,
		})
	}

	resp := api.LoadResponse{
		Success:        true,
		TotalRecords:   result.TotalRecords,
		LoadedRecords:  result.LoadedRecords,
		InvalidRecords: invalidRecords,
		Warnings:       result.Warnings,
	}

	writeJSON(w, http.StatusOK, resp)
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.ErrorResponse{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	var req api.AddBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	book := &core.Book{
		ISBN:         strings.TrimSpace(req.ISBN),
		Title:        strings.TrimSpace(req.Title),
		Authors:      core.ParseAuthors(req.Authors),
		Publisher:    strings.TrimSpace(req.Publisher),
		Year:         req.Year,
		CategoryCode: strings.TrimSpace(req.CategoryCode),
	}

	if ok := store.Add(book); !ok {
		writeJSON(w, http.StatusBadRequest, api.AddBookResponse{
			Success: false,
			Message: "Failed to add book: validation failed",
		})
		return
	}

	writeJSON(w, http.StatusOK, api.AddBookResponse{
		Success: true,
		Message: "Book added successfully",
	})
}
