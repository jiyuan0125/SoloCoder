package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"go-reflect-tags/pkg/common"
	"go-reflect-tags/pkg/tagparser"
)

var parser = tagparser.NewParser()

func main() {
	var port string
	flag.StringVar(&port, "port", "", "server port (default: 8080)")
	flag.Parse()

	if port == "" {
		port = os.Getenv("SERVER_PORT")
		if port == "" {
			port = "8080"
		}
	}

	http.HandleFunc("/register", handleRegister)
	http.HandleFunc("/query", handleQuery)

	log.Printf("server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		sendError(w, http.StatusBadRequest, "struct name is required")
		return
	}

	parser.RegisterStruct(req.Name, req.Fields)
	sendSuccess(w, "struct registered successfully")
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.StructName) == "" {
		sendError(w, http.StatusBadRequest, "struct name is required")
		return
	}

	info, err := parser.QueryField(req.StructName, req.FieldPath)
	if err != nil {
		if tagparser.IsNotFoundError(err) {
			sendError(w, http.StatusNotFound, err.Error())
		} else {
			sendError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.QueryResponse{
		Success: true,
		Data:    info,
	})
}

func sendSuccess(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.RegisterResponse{
		Success: true,
		Message: message,
	})
}

func sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"success":false,"message":%q}`, message)
}
