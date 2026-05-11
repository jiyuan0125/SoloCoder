package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/asn1-der-parser/pkg/api"
	"github.com/asn1-der-parser/pkg/asn1"
)

func main() {
	http.HandleFunc("/decode", decodeHandler)
	http.HandleFunc("/encode", encodeHandler)

	port := ":8450"
	fmt.Printf("ASN.1 DER Parser Server listening on %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func decodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.DecodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	data, err := req.GetBytes()
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid base64 data")
		return
	}

	node, _, err := asn1.Decode(data)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := api.DecodeResponse{
		Success: true,
		Result:  asn1.NodeToJSON(node),
		Dump:    node.Dump(0),
	}

	json.NewEncoder(w).Encode(resp)
}

func encodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.EncodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Node == nil {
		sendError(w, http.StatusBadRequest, "node is required")
		return
	}

	node, err := asn1.NodeFromJSON(req.Node)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	encoded, err := asn1.Encode(node)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := api.EncodeResponse{
		Success: true,
	}
	resp.SetBytes(encoded)

	json.NewEncoder(w).Encode(resp)
}

func sendError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(api.DecodeResponse{
		Success: false,
		Error:   msg,
	})
}
