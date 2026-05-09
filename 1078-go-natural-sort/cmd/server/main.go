package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"naturalsort/pkg/api"
	"naturalsort/pkg/natsort"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	http.HandleFunc("/sort", handleSort)

	log.Printf("server listening on %s", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func handleSort(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.SortRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := api.SortResponse{Error: fmt.Sprintf("invalid request: %v", err)}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	opts := natsort.Options{
		Ascending:          req.Ascending,
		CaseSensitive:      req.CaseSensitive,
		IgnoreLeadingZeros: req.IgnoreLeadingZeros,
	}

	sorted := natsort.Sort(req.Strings, opts)

	resp := api.SortResponse{Strings: sorted}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
