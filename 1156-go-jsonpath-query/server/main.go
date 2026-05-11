package main

import (
	"encoding/json"
	"log"
	"net/http"

	"jsonpath/api"
	"jsonpath/jsonpath"
)

func main() {
	http.HandleFunc("/query", handleQuery)
	log.Println("JSONPath server starting on :8600")
	log.Fatal(http.ListenAndServe(":8600", nil))
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.JSON == "" || req.Path == "" {
		sendError(w, "missing json or path", http.StatusBadRequest)
		return
	}

	var data interface{}
	if err := json.Unmarshal([]byte(req.JSON), &data); err != nil {
		sendError(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}

	matches, err := jsonpath.QueryWithPaths(data, req.Path)
	if err != nil {
		sendError(w, "jsonpath error: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp := api.QueryResponse{
		MatchCount: len(matches),
	}
	if len(matches) > 0 {
		paths := make([]string, len(matches))
		values := make([]interface{}, len(matches))
		for i, m := range matches {
			values[i] = m.Value
			paths[i] = m.Path
		}
		if len(matches) == 1 {
			resp.Results = values[0]
		} else {
			resp.Results = values
		}
		resp.Paths = paths
	}

	_ = json.NewEncoder(w).Encode(resp)
}

func sendError(w http.ResponseWriter, msg string, code int) {
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(api.QueryResponse{Error: msg})
}
