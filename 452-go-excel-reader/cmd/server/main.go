package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"xlsx-reader/pkg/common"
	"xlsx-reader/pkg/xlsxreader"
)

const uploadDir = "./uploads"

func main() {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		fmt.Printf("Failed to create upload directory: %v\n", err)
		os.Exit(1)
	}

	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/read-sheet", readSheetHandler)
	http.HandleFunc("/list-sheets", listSheetsHandler)

	fmt.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
		os.Exit(1)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := filepath.Join(uploadDir, header.Filename)
	dst, err := os.Create(filename)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"success":  "true",
		"filename": header.Filename,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func readSheetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.ReadSheetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request body")
		return
	}

	filename := r.URL.Query().Get("filename")
	if filename == "" {
		sendErrorResponse(w, "Filename is required")
		return
	}

	filepath := filepath.Join(uploadDir, filename)
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		sendErrorResponse(w, "File not found: "+filename)
		return
	}

	reader, err := xlsxreader.Open(filepath)
	if err != nil {
		sendErrorResponse(w, "Failed to open file: "+err.Error())
		return
	}
	defer reader.Close()

	options := xlsxreader.ReadOptions{
		StartRow:      req.StartRow,
		EndRow:        req.EndRow,
		SkipEmptyRows: req.SkipEmptyRows,
	}

	if options.SkipEmptyRows && req.StartRow == 0 && req.EndRow == 0 {
		options.SkipEmptyRows = true
	}

	data, err := reader.ReadSheet(req.SheetName, options)
	if err != nil {
		if sheetErr, ok := err.(*xlsxreader.SheetNotFoundError); ok {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(common.ReadSheetResponse{
				Success:    false,
				SheetNames: sheetErr.AvailableSheets,
				Error:      sheetErr.Error(),
			})
			return
		}
		sendErrorResponse(w, "Failed to read sheet: "+err.Error())
		return
	}

	response := common.ReadSheetResponse{
		Success: true,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func listSheetsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := r.URL.Query().Get("filename")
	if filename == "" {
		sendListSheetsError(w, "Filename is required")
		return
	}

	filepath := filepath.Join(uploadDir, filename)
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		sendListSheetsError(w, "File not found: "+filename)
		return
	}

	reader, err := xlsxreader.Open(filepath)
	if err != nil {
		sendListSheetsError(w, "Failed to open file: "+err.Error())
		return
	}
	defer reader.Close()

	sheetNames := reader.GetSheetNames()

	response := common.ListSheetsResponse{
		Success:    true,
		SheetNames: sheetNames,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func sendErrorResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.ReadSheetResponse{
		Success: false,
		Error:   message,
	})
}

func sendListSheetsError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.ListSheetsResponse{
		Success: false,
		Error:   message,
	})
}
