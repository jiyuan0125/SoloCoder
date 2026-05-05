package main

import (
	"encoding/json"
	"id-number-parser/common"
	"id-number-parser/idparser"
	"net/http"
	"time"
)

func parseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持 POST 方法", http.StatusMethodNotAllowed)
		return
	}

	var req common.ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.ParseResponse{
			Success: false,
			Message: "请求体解析失败",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	info, err := idparser.Parse(req.IDNumber)
	if err != nil {
		resp := common.ParseResponse{
			Success: false,
			Message: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := common.ParseResponse{
		Success:        true,
		OriginalID:     info.OriginalID,
		StandardizedID: info.StandardizedID,
		BirthDate:      info.BirthDate.Format(time.RFC3339),
		Gender:         info.Gender,
		ProvinceCode:   info.ProvinceCode,
		ProvinceName:   info.ProvinceName,
		Age:            info.Age,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持 POST 方法", http.StatusMethodNotAllowed)
		return
	}

	var req common.ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.ValidateResponse{
			Success: false,
			Message: "请求体解析失败",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	_, err := idparser.Parse(req.IDNumber)
	if err != nil {
		resp := common.ValidateResponse{
			Success: true,
			Valid:   false,
			Message: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := common.ValidateResponse{
		Success: true,
		Valid:   true,
		Message: "身份证号有效",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
