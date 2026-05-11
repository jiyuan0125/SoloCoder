package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"unicode-normalizer/pkg/api"
	"unicode-normalizer/pkg/uninorm"
)

const version = "1.0.0"

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/api/normalize", handleNormalize)
	mux.HandleFunc("/api/analyze", handleAnalyze)

	fmt.Printf("Unicode Normalizer Server v%s starting on :8101\n", version)
	fmt.Printf("Unicode Version: %s\n", uninorm.UnicodeVersion)
	
	if err := http.ListenAndServe(":8101", mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := api.HealthResponse{
		Status:  "ok",
		Version: version,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleNormalize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.NormalizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request body: "+err.Error())
		return
	}

	if strings.TrimSpace(req.Form) == "" {
		req.Form = "NFC"
	}

	form, err := uninorm.ParseForm(req.Form)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	normalized := uninorm.Normalize(req.Text, form)

	response := api.NormalizeResponse{
		Success:        true,
		NormalizedText: normalized,
		OriginalText:   req.Text,
		Form:           form.String(),
		UnicodeVersion: uninorm.UnicodeVersion,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "Invalid request body: "+err.Error())
		return
	}

	if strings.TrimSpace(req.Form) == "" {
		req.Form = "NFC"
	}

	form, err := uninorm.ParseForm(req.Form)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	changes, normalized := uninorm.AnalyzeChanges(req.Text, form)

	changeInfos := make([]api.CharChangeInfo, 0, len(changes))
	for _, change := range changes {
		normalizedCodes := make([]string, 0, len(change.Normalized))
		for _, r := range change.Normalized {
			normalizedCodes = append(normalizedCodes, fmt.Sprintf("U+%04X", r))
		}

		originalCodes := make([]string, 0, len(change.OriginalRunes))
		for _, r := range change.OriginalRunes {
			originalCodes = append(originalCodes, fmt.Sprintf("U+%04X", r))
		}

		changeInfos = append(changeInfos, api.CharChangeInfo{
			Original:        change.OriginalStr,
			OriginalCode:    strings.Join(originalCodes, " "),
			Normalized:      change.NormalizedStr,
			NormalizedCodes: normalizedCodes,
		})
	}

	response := api.AnalyzeResponse{
		Success:        true,
		NormalizedText: normalized,
		OriginalText:   req.Text,
		Form:           form.String(),
		UnicodeVersion: uninorm.UnicodeVersion,
		Changes:        changeInfos,
		ChangeCount:    len(changeInfos),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func sendErrorResponse(w http.ResponseWriter, errMsg string) {
	response := api.NormalizeResponse{
		Success: false,
		Error:   errMsg,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}
