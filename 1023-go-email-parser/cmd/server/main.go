package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/solocoder/email-parser/pkg/api"
	"github.com/solocoder/email-parser/pkg/email"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/api/parse", handleParse)
	http.HandleFunc("/api/validate", handleValidate)
	http.HandleFunc("/api/batch", handleBatchParse)

	log.Printf("Email parser server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	addr, err := email.Parse(req.Email)
	if err != nil {
		sendJSON(w, api.ParseResponse{
			Success: true,
			Valid:   false,
			Error:   err.Error(),
		})
		return
	}

	sendJSON(w, api.ParseResponse{
		Success: true,
		Valid:   true,
		Address: toAddressInfo(addr),
	})
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	valid := email.Validate(req.Email)
	sendJSON(w, api.ValidateResponse{
		Success: true,
		Valid:   valid,
	})
}

func handleBatchParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.BatchParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := email.ParseBatch(req.Emails)
	if err != nil {
		sendJSON(w, api.BatchParseResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	results := make([]api.BatchParseResult, len(result.Results))
	for i, r := range result.Results {
		batchResult := api.BatchParseResult{
			Valid: r.Valid,
		}
		if r.Address != nil {
			batchResult.Email = r.Address.Original
			batchResult.Address = toAddressInfo(r.Address)
		}
		if r.Error != nil {
			batchResult.Error = r.Error.Error()
			// 尝试从错误中获取原始邮箱
			// 这里简化处理，不存储原始邮箱
		}
		results[i] = batchResult
	}

	sendJSON(w, api.BatchParseResponse{
		Success: true,
		Total:   result.Total,
		Valid:   result.Valid,
		Invalid: result.Invalid,
		Results: results,
	})
}

func toAddressInfo(addr *email.Address) *api.AddressInfo {
	if addr == nil {
		return nil
	}
	return &api.AddressInfo{
		Original:      addr.Original,
		Local:         addr.Local,
		Domain:        addr.Domain,
		HasAlias:      addr.HasAlias,
		BaseLocal:     addr.BaseLocal,
		BaseAddress:   addr.BaseAddress,
		IsQuotedLocal: addr.IsQuotedLocal,
	}
}

func sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}
