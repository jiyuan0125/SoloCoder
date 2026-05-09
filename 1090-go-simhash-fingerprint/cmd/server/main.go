package main

import (
	"encoding/json"
	"net/http"

	"simhash/pkg/common"
	"simhash/pkg/fingerprint"
)

func main() {
	http.HandleFunc("/fingerprint", handleFingerprint)
	http.HandleFunc("/hamming", handleHamming)
	http.HandleFunc("/similar", handleSimilar)
	http.HandleFunc("/dedup", handleDedup)

	http.ListenAndServe(":8080", nil)
}

func handleFingerprint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req common.FingerprintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	cfg := &fingerprint.Config{
		Threshold: fingerprint.DefaultConfig.Threshold,
		NGram:     fingerprint.DefaultConfig.NGram,
	}
	if req.NGram > 0 {
		cfg.NGram = req.NGram
	}
	fp := fingerprint.FingerprintText(req.Text, cfg)
	resp := common.FingerprintResponse{Fingerprint: fp.String()}
	json.NewEncoder(w).Encode(resp)
}

func handleHamming(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req common.HammingDistanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	fp1, err := fingerprint.Parse(req.FP1)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	fp2, err := fingerprint.Parse(req.FP2)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	dist := fingerprint.HammingDistance(fp1, fp2)
	resp := common.HammingDistanceResponse{Distance: dist}
	json.NewEncoder(w).Encode(resp)
}

func handleSimilar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req common.IsSimilarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	fp1, err := fingerprint.Parse(req.FP1)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	fp2, err := fingerprint.Parse(req.FP2)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	cfg := &fingerprint.Config{
		Threshold: fingerprint.DefaultConfig.Threshold,
		NGram:     fingerprint.DefaultConfig.NGram,
	}
	if req.Threshold > 0 {
		cfg.Threshold = req.Threshold
	}
	similar := fingerprint.IsSimilar(fp1, fp2, cfg)
	resp := common.IsSimilarResponse{Similar: similar}
	json.NewEncoder(w).Encode(resp)
}

func handleDedup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req common.DedupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	cfg := &fingerprint.Config{
		Threshold: fingerprint.DefaultConfig.Threshold,
		NGram:     fingerprint.DefaultConfig.NGram,
	}
	if req.Threshold > 0 {
		cfg.Threshold = req.Threshold
	}
	if req.NGram > 0 {
		cfg.NGram = req.NGram
	}
	result := fingerprint.Dedup(req.Docs, cfg)
	resp := common.DedupResponse{Docs: result}
	json.NewEncoder(w).Encode(resp)
}

func sendError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(common.ErrorResponse{Error: msg})
}
