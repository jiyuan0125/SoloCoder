package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"punycode-idn/pkg/api"
	"punycode-idn/pkg/punycode"
)

func encodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.EncodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.EncodeResponse{Error: err.Error()})
		return
	}

	encoded, err := punycode.EncodeDomain(req.Domain)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.EncodeResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.EncodeResponse{Domain: encoded})
}

func decodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.DecodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.DecodeResponse{Error: err.Error()})
		return
	}

	decoded, err := punycode.DecodeDomain(req.Domain)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.DecodeResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.DecodeResponse{Domain: decoded})
}

func getPort() string {
	port := os.Getenv("PORT")
	if port != "" {
		return ":" + port
	}

	portFlag := flag.String("port", "8310", "server port")
	flag.Parse()
	return ":" + *portFlag
}

func main() {
	port := getPort()

	http.HandleFunc("/encode", encodeHandler)
	http.HandleFunc("/decode", decodeHandler)

	fmt.Printf("server listening on %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
