package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"cardcheck/internal/cardverify"
	"cardcheck/internal/models"
)

func main() {
	http.HandleFunc("/validate", validateHandler)
	http.HandleFunc("/batch-validate", batchValidateHandler)

	port := ":8500"
	fmt.Printf("银行卡号校验服务启动中，监听端口 %s...\n", port)
	fmt.Println("接口说明:")
	fmt.Println("  POST /validate - 单卡校验")
	fmt.Println("  POST /batch-validate - 批量校验")
	
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	var req models.ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "请求参数解析失败")
		return
	}

	if req.CardNumber == "" {
		writeErrorResponse(w, http.StatusBadRequest, "银行卡号不能为空")
		return
	}

	result, _ := cardverify.ValidateCardNumber(req.CardNumber)
	
	response := models.ValidateResponse{
		Success:          true,
		CardNumber:       result.RawNumber,
		CleanNumber:      result.CleanNumber,
		IsValid:          result.IsValid,
		ErrorReason:      result.ErrorReason,
		CardOrganization: models.CardOrganization(result.CardOrganization),
		BankName:         result.BankName,
		CardType:         models.CardType(result.CardType),
		BIN:             result.BIN,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func batchValidateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	var req models.BatchValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "请求参数解析失败")
		return
	}

	if len(req.CardNumbers) == 0 {
		writeErrorResponse(w, http.StatusBadRequest, "银行卡号列表不能为空")
		return
	}

	results := make([]models.ValidateResponse, 0, len(req.CardNumbers))
	for _, cardNumber := range req.CardNumbers {
		result, _ := cardverify.ValidateCardNumber(cardNumber)
		results = append(results, models.ValidateResponse{
			Success:          true,
			CardNumber:       result.RawNumber,
			CleanNumber:      result.CleanNumber,
			IsValid:          result.IsValid,
			ErrorReason:      result.ErrorReason,
			CardOrganization: models.CardOrganization(result.CardOrganization),
			BankName:         result.BankName,
			CardType:         models.CardType(result.CardType),
			BIN:             result.BIN,
		})
	}

	response := models.BatchValidateResponse{
		Success: true,
		Results: results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(models.ErrorResponse{
		Success: false,
		Error:   message,
	})
}
