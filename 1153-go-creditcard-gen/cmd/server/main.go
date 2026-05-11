package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/creditcard-gen/pkg/api"
	"github.com/creditcard-gen/pkg/creditcard"
)

const (
	TestWarningMessage = "仅用于测试的虚拟卡号，不可用于真实交易"
)

func main() {
	port := flag.String("port", "8502", "server port")
	flag.Parse()

	mux := http.NewServeMux()

	mux.HandleFunc("/generate", handleGenerate)
	mux.HandleFunc("/validate", handleValidate)
	mux.HandleFunc("/health", handleHealth)

	addr := ":" + *port
	log.Printf("Credit Card Generator Server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func setWarningHeaders(w http.ResponseWriter) {
	w.Header().Set(api.HeaderWarning, api.HeaderWarningValue)
	w.Header().Set("Content-Type", "application/json")
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	setWarningHeaders(w)
	w.WriteHeader(statusCode)
	resp := map[string]interface{}{
		"success": false,
		"error":   message,
		"warning": TestWarningMessage,
	}
	json.NewEncoder(w).Encode(resp)
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ct, err := creditcard.ParseCardType(req.CardType)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("unsupported card type: %s", req.CardType))
		return
	}

	if req.Count <= 0 {
		writeErrorResponse(w, http.StatusBadRequest, "count must be positive")
		return
	}

	if req.Count > creditcard.MaxGenerateCount {
		writeErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("count exceeds maximum limit of %d", creditcard.MaxGenerateCount))
		return
	}

	numbers, err := creditcard.GenerateCardNumbers(ct, req.Count)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	cards := make([]api.CardInfo, len(numbers))
	for i, num := range numbers {
		cards[i] = api.CardInfo{
			Number:      num,
			CardType:    string(ct),
			DisplayName: creditcard.CardTypeDisplayName(ct),
		}
	}

	resp := api.GenerateResponse{
		Success:  true,
		CardType: string(ct),
		Count:    req.Count,
		Cards:    cards,
		Warning:  TestWarningMessage,
	}

	setWarningHeaders(w)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cardNumber := strings.TrimSpace(req.CardNumber)
	if cardNumber == "" {
		writeErrorResponse(w, http.StatusBadRequest, "card number is required")
		return
	}

	result := creditcard.ValidateCardNumber(cardNumber)

	resp := api.ValidateResponse{
		Success:    true,
		Valid:      result.Valid,
		CardNumber: cardNumber,
		Message:    result.Message,
		Warning:    TestWarningMessage,
	}

	if result.Valid {
		resp.CardType = result.CardTypeStr
	}

	setWarningHeaders(w)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
