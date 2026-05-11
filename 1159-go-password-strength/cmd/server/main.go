package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/passwordstrength/common"
	"github.com/passwordstrength/passwordstrength"
)

func maskPassword(pwd string) string {
	if len(pwd) == 0 {
		return ""
	}
	return strings.Repeat("*", len(pwd))
}

func evaluateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req common.EvaluateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	log.Printf("[INFO] Received evaluate request: password_len=%d, masked=%s", len(req.Password), maskPassword(req.Password))
	resp := passwordstrength.Evaluate(req.Password)
	log.Printf("[INFO] Evaluation complete: score=%.1f, level=%s, is_common=%v", resp.TotalScore, resp.Level, resp.IsCommon)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	http.HandleFunc("/api/evaluate", evaluateHandler)
	http.HandleFunc("/health", healthHandler)
	log.Println("[INFO] Password strength evaluation server starting on :8603")
	log.Println("[INFO] Security: Passwords are never logged in plaintext")
	if err := http.ListenAndServe(":8603", nil); err != nil {
		log.Fatalf("[ERROR] Server failed to start: %v", err)
	}
}
