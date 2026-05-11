package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/spf-router/common"
	"github.com/spf-router/spf"
)

var (
	topology     *spf.Topology
	spfCalc      *spf.SPFCalculator
	port         int
)

func init() {
	flag.IntVar(&port, "port", 8600, "Server port")
}

func main() {
	flag.Parse()

	topology = spf.NewTopology()
	spfCalc = spf.NewSPFCalculator(topology)

	http.HandleFunc("/topology/update", updateTopologyHandler)
	http.HandleFunc("/spf/run", runSPFHandler)
	http.HandleFunc("/spf/paths", getPathsHandler)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("SPF Router Server starting on port %d", port)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func updateTopologyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.TopologyUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, node := range req.AddNodes {
		topology.AddNode(node.ID)
	}

	for _, nodeID := range req.RemoveNodes {
		topology.RemoveNode(nodeID)
	}

	for _, link := range req.AddLinks {
		topology.AddLink(spf.Link{
			From:   link.From,
			To:     link.To,
			Cost:   link.Cost,
			SeqNum: link.SeqNum,
		})
	}

	for _, link := range req.RemoveLinks {
		topology.RemoveLink(spf.Link{
			From: link.From,
			To:   link.To,
		})
	}

	resp := common.TopologyUpdateResponse{
		Success: true,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func runSPFHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.SPFRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Source == "" {
		http.Error(w, "Source node is required", http.StatusBadRequest)
		return
	}

	err := spfCalc.ComputeSPF(req.Source)
	if err != nil {
		resp := common.SPFRunResponse{
			Success: false,
			Message: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := common.SPFRunResponse{
		Success: true,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getPathsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.ShortestPathQueryRequest
	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		req.Source = r.URL.Query().Get("source")
	}

	if req.Source == "" {
		http.Error(w, "Source node is required", http.StatusBadRequest)
		return
	}

	paths, err := spfCalc.GetShortestPaths(req.Source)
	if err != nil {
		resp := common.ShortestPathQueryResponse{
			Success: false,
			Message: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	var entries []common.PathEntry
	for _, p := range paths {
		entry := common.PathEntry{
			Destination: p.Destination,
			Paths:       p.Paths,
			TotalCost:   p.TotalCost,
			Valid:       p.Valid,
		}
		entries = append(entries, entry)
	}

	resp := common.ShortestPathQueryResponse{
		Success: true,
		Paths:   entries,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
