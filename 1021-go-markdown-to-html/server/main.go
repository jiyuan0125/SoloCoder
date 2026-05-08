package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"md2html/common"
	"md2html/mdparser"
)

func convertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		response := common.ConvertResponse{
			Success: false,
			Error:   "Failed to read request body",
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	defer r.Body.Close()
	
	var req common.ConvertRequest
	if err := json.Unmarshal(body, &req); err != nil {
		response := common.ConvertResponse{
			Success: false,
			Error:   "Invalid JSON",
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	
	html := mdparser.Convert(req.Markdown)
	
	response := common.ConvertResponse{
		HTML:    html,
		Success: true,
	}
	json.NewEncoder(w).Encode(response)
}

func main() {
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}
	
	http.HandleFunc("/convert", convertHandler)
	
	fmt.Printf("Server listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}
