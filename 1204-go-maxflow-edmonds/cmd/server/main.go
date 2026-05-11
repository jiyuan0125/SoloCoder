package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"maxflow/pkg/api"
	"maxflow/pkg/maxflow"
)

type flowHandler struct{}

func (h *flowHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.FlowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	g := maxflow.BuildGraphFromRequest(&req)
	result, err := g.ComputeMaxFlow(req.Source, req.Sink)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := api.FlowResponse{
		MaxFlow: result.MaxFlow,
		Flows:   make([]api.FlowResult, len(result.Flows)),
	}

	for i, f := range result.Flows {
		resp.Flows[i] = api.FlowResult{
			From: f.From,
			To:   f.To,
			Flow: f.Flow,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getPort() string {
	port := "8080"

	if envPort := os.Getenv("MAXFLOW_PORT"); envPort != "" {
		port = envPort
	}

	flagPort := flag.String("port", "", "server port (overrides MAXFLOW_PORT env)")
	flag.Parse()

	if *flagPort != "" {
		port = *flagPort
	}

	return port
}

func main() {
	port := getPort()
	addr := ":" + port

	http.Handle("/maxflow", &flowHandler{})

	log.Printf("server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
