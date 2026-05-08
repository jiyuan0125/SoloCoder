package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/solocoder/consistenthash/api"
	"github.com/solocoder/consistenthash/pkg/consistenthash"
)

func main() {
	addr := flag.String("addr", ":8080", "Server address")
	flag.Parse()

	ring := consistenthash.NewHashRing(nil)

	http.HandleFunc("POST /hash/nodes", handleAddNode(ring))
	http.HandleFunc("DELETE /hash/nodes/{id}", handleRemoveNode(ring))
	http.HandleFunc("GET /hash/lookup", handleLookup(ring))
	http.HandleFunc("GET /hash/stats", handleStats(ring))

	log.Printf("Server listening on %s", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func handleAddNode(ring *consistenthash.HashRing) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req api.AddNodeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Error: "invalid request body"})
			return
		}
		defer r.Body.Close()

		if strings.TrimSpace(req.ID) == "" {
			writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Error: "node id is required"})
			return
		}

		ring.AddNode(req.ID, req.Weight)

		writeJSON(w, http.StatusOK, api.AddNodeResponse{
			Success: true,
			Message: "node added successfully",
		})
	}
}

func handleRemoveNode(ring *consistenthash.HashRing) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if strings.TrimSpace(id) == "" {
			writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Error: "node id is required"})
			return
		}

		removed := ring.RemoveNode(id)
		if !removed {
			writeJSON(w, http.StatusNotFound, api.ErrorResponse{Error: "node not found"})
			return
		}

		writeJSON(w, http.StatusOK, api.RemoveNodeResponse{
			Success: true,
			Message: "node removed successfully",
		})
	}
}

func handleLookup(ring *consistenthash.HashRing) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if strings.TrimSpace(key) == "" {
			writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Error: "key is required"})
			return
		}

		node, err := ring.GetNode(key)
		if err != nil {
			if errors.Is(err, consistenthash.ErrEmptyRing) {
				writeJSON(w, http.StatusServiceUnavailable, api.LookupResponse{
					Error: "no nodes available",
				})
				return
			}
			writeJSON(w, http.StatusInternalServerError, api.LookupResponse{
				Error: err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, api.LookupResponse{
			Node: node,
		})
	}
}

func handleStats(ring *consistenthash.HashRing) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := ring.Stats()

		nodeInfos := make(map[string]api.NodeInfo, len(stats.NodeStats))
		for name, ns := range stats.NodeStats {
			nodeInfos[name] = api.NodeInfo{
				Weight:           ns.Weight,
				VirtualNodeCount: ns.VirtualNodeCount,
				Percentage:       ns.Percentage,
			}
		}

		writeJSON(w, http.StatusOK, api.StatsResponse{
			TotalVirtualNodes: stats.TotalVirtualNodes,
			Nodes:             nodeInfos,
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}
