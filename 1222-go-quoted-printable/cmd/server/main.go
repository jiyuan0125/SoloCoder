package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"quoted-printable/pkg/api"
	"quoted-printable/pkg/qp"
)

func encodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(api.EncodeResponse{Error: "method not allowed"})
		return
	}

	var req api.EncodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.EncodeResponse{Error: "invalid request body"})
		return
	}

	encoded, err := qp.EncodeString(req.Text)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(api.EncodeResponse{Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(api.EncodeResponse{Encoded: encoded})
}

func decodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(api.DecodeResponse{Error: "method not allowed"})
		return
	}

	var req api.DecodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.DecodeResponse{Error: "invalid request body"})
		return
	}

	decoded, err := qp.DecodeString(req.Encoded)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.DecodeResponse{Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(api.DecodeResponse{Decoded: decoded})
}

func main() {
	port := flag.String("port", "", "server port")
	flag.Parse()

	if *port == "" {
		*port = os.Getenv("QP_PORT")
	}

	if *port == "" {
		*port = "8304"
	}

	http.HandleFunc("/encode", encodeHandler)
	http.HandleFunc("/decode", decodeHandler)

	fmt.Printf("server listening on port %s\n", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
