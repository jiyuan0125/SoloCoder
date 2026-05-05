package main

import (
	"encoding/json"
	"net/http"

	"bankcard/cardparser"
	"bankcard/common"
)

func ParseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ErrorResponse{Error: "Invalid request body"})
		return
	}

	if req.CardNumber == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ErrorResponse{Error: "Card number is required"})
		return
	}

	info := cardparser.ParseCardNumber(req.CardNumber)

	resp := common.ParseResponse{
		CardNumber:      info.CardNumber,
		BankName:        info.BankName,
		CardType:        common.CardType(info.CardType),
		IsValid:         info.IsValid,
		Error:           info.Error,
		MaskedNumber:    info.MaskedNumber,
		FormattedNumber: info.FormattedNumber,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
