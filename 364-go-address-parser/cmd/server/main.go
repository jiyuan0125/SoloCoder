package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	
	"address-parser/common"
	"address-parser/parser"
)

func parseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		http.Error(w, `{"success": false, "message": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	
	var req common.ParseRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, `{"success": false, "message": "invalid request body"}`, http.StatusBadRequest)
		return
	}
	
	result, err := parser.Parse(req.Address)
	if err != nil {
		response := common.ParseResponse{
			Success: false,
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	
	response := common.ParseResponse{
		Success:  true,
		Province: result.Province,
		City:     result.City,
		District: result.District,
		Detail:   result.Detail,
	}
	
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/parse", parseHandler)
	
	fmt.Println("Server starting on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
