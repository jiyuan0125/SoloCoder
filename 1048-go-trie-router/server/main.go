package main

import (
	"encoding/json"
	"log"
	"net/http"

	"trie-router/common"
	"trie-router/core"
)

var router *core.Router

func main() {
	router = core.NewRouter()

	http.HandleFunc("/api/routes/register", handleRegister)
	http.HandleFunc("/api/routes/query", handleQuery)
	http.HandleFunc("/api/routes/list", handleList)

	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.RegisterResponse{
			Success: false,
			Error:   err.Error(),
		}
		writeJSON(w, http.StatusBadRequest, resp)
		return
	}

	if err := router.Register(req.Path, req.Method, req.HandlerName); err != nil {
		resp := common.RegisterResponse{
			Success: false,
			Error:   err.Error(),
		}
		writeJSON(w, http.StatusBadRequest, resp)
		return
	}

	resp := common.RegisterResponse{
		Success: true,
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.QueryResponse{
			Found:  false,
			Status: http.StatusBadRequest,
			Error:  err.Error(),
		}
		writeJSON(w, http.StatusBadRequest, resp)
		return
	}

	result := router.Match(req.Path, req.Method)

	resp := common.QueryResponse{
		Found:          result.Found,
		Route:          result.Route,
		Params:         result.Params,
		AllowedMethods: result.AllowedMethods,
		Status:         result.Status,
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var sortBy string
	if r.Method == http.MethodPost {
		var req common.ListRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			sortBy = req.SortBy
		}
	} else {
		sortBy = r.URL.Query().Get("sort_by")
	}

	routes := router.List(sortBy)
	resp := common.ListResponse{
		Routes: routes,
		Count:  len(routes),
	}
	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
