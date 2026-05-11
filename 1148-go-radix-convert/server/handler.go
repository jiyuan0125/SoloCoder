package main

import (
	"encoding/json"
	"net/http"

	"baseconv/baseconv"
	"baseconv/common"
)

func convertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(common.ConvertResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req common.ConvertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ConvertResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	if req.Precision <= 0 {
		req.Precision = 32
	}

	result, err := baseconv.Convert(req.Input, req.FromBase, req.ToBase, req.Precision)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ConvertResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(common.ConvertResponse{
		Success: true,
		Result:  result,
	})
}
