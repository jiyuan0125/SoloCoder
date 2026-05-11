package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"tar-archiver/common"
	"tar-archiver/core"
)

func main() {
	http.HandleFunc("/archive", handleArchive)
	http.HandleFunc("/extract", handleExtract)

	addr := ":8302"
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func handleArchive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.ArchiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	format := core.FormatPOSIX
	if req.Format == common.FormatGNU {
		format = core.FormatGNU
	}

	pr, err := core.ArchivePaths(req.Paths, format)
	if err != nil {
		http.Error(w, fmt.Sprintf("Archive error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=archive.tar")
	io.Copy(w, pr)
}

func handleExtract(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	destDir := r.URL.Query().Get("dest")
	if destDir == "" {
		destDir = "."
	}

	if err := core.ExtractArchive(r.Body, destDir); err != nil {
		http.Error(w, fmt.Sprintf("Extract error: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
