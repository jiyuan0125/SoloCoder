package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/example/msgpack-codec/pkg/msgpack"
	"github.com/example/msgpack-codec/pkg/protocol"
)

func decodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(body) == 0 {
		writeError(w, "empty request body", http.StatusBadRequest)
		return
	}

	v, err := msgpack.Unmarshal(body)
	if err != nil {
		writeError(w, "failed to decode msgpack: "+err.Error(), http.StatusBadRequest)
		return
	}

	jsonData, err := msgpack.ToJSON(v)
	if err != nil {
		writeError(w, "failed to convert to JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func encodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(body) == 0 {
		writeError(w, "empty request body", http.StatusBadRequest)
		return
	}

	v, err := msgpack.FromJSON(body)
	if err != nil {
		writeError(w, "failed to parse JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	msgpackData, err := msgpack.Marshal(v)
	if err != nil {
		writeError(w, "failed to encode to msgpack: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.Write(msgpackData)
}

func writeError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(protocol.ErrorResponse{Error: message})
}

func main() {
	http.HandleFunc(protocol.EndpointDecode, decodeHandler)
	http.HandleFunc(protocol.EndpointEncode, encodeHandler)

	addr := ":8080"
	log.Printf("server listening on %s", addr)
	log.Printf("  POST %s - decode msgpack to json", protocol.EndpointDecode)
	log.Printf("  POST %s - encode json to msgpack", protocol.EndpointEncode)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
