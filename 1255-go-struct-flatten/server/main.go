package main

import (
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"os"
	"strings"

	"go-struct-flatten/common"
	"go-struct-flatten/flatten"
)

func main() {
	port := getPort()

	http.HandleFunc("/api/list", handleList)
	http.HandleFunc("/api/flatten", handleFlatten)
	http.HandleFunc("/health", handleHealth)

	println("Server listening on :" + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}

func getPort() string {
	port := flag.String("port", "", "port to listen on")
	flag.Parse()

	if *port != "" {
		return *port
	}

	if envPort := os.Getenv("PORT"); envPort != "" {
		return envPort
	}

	return "8080"
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.ListStructsResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.ListStructsResponse{
			Success: false,
			Error:   "failed to read request body",
		})
		return
	}

	var req common.ListStructsRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.ListStructsResponse{
			Success: false,
			Error:   "invalid JSON: " + err.Error(),
		})
		return
	}

	if strings.TrimSpace(req.SourceCode) == "" {
		writeJSON(w, http.StatusBadRequest, common.ListStructsResponse{
			Success: false,
			Error:   "source_code is empty",
		})
		return
	}

	flattener, err := flatten.NewFlattener(req.SourceCode)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.ListStructsResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, common.ListStructsResponse{
		Success: true,
		Structs: flattener.ListStructs(),
	})
}

func handleFlatten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.FlattenResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.FlattenResponse{
			Success: false,
			Error:   "failed to read request body",
		})
		return
	}

	var req common.FlattenRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.FlattenResponse{
			Success: false,
			Error:   "invalid JSON: " + err.Error(),
		})
		return
	}

	if strings.TrimSpace(req.SourceCode) == "" {
		writeJSON(w, http.StatusBadRequest, common.FlattenResponse{
			Success: false,
			Error:   "source_code is empty",
		})
		return
	}

	if strings.TrimSpace(req.StructName) == "" {
		writeJSON(w, http.StatusBadRequest, common.FlattenResponse{
			Success: false,
			Error:   "struct_name is empty",
		})
		return
	}

	flattener, err := flatten.NewFlattener(req.SourceCode)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.FlattenResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	fields, err := flattener.Flatten(req.StructName)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.FlattenResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, common.FlattenResponse{
		Success: true,
		Fields:  fields,
	})
}
