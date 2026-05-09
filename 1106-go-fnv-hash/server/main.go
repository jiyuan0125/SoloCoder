package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/example/fnv/pkg/api"
	"github.com/example/fnv/pkg/fnv"
)

const (
	DefaultReplicas = 150
)

var consistentHash *fnv.ConsistentHash

func main() {
	consistentHash = fnv.NewConsistentHash(DefaultReplicas)

	http.HandleFunc("/hash", hashHandler)
	http.HandleFunc("/consistent", consistentHandler)
	http.HandleFunc("/nodes/add", addNodeHandler)
	http.HandleFunc("/nodes/remove", removeNodeHandler)
	http.HandleFunc("/nodes", listNodesHandler)

	addr := ":8080"
	fmt.Printf("FNV hash server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func hashHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.HashResponse{
			Success: false,
			Error:   "method not allowed, use POST",
		})
		return
	}
	var req api.HashRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.HashResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}
	if req.Algorithm == "" {
		req.Algorithm = "fnv1a"
	}
	hashVal, err := fnv.Hash(req.Algorithm, req.Bits, req.Input)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.HashResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, api.HashResponse{
		Success: true,
		Hash:    hashVal,
	})
}

func consistentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.ConsistentResponse{
			Success: false,
			Error:   "method not allowed, use POST",
		})
		return
	}
	var req api.ConsistentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ConsistentResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}
	if len(req.Nodes) == 0 {
		req.Nodes = consistentHash.Nodes()
	}
	tempCh := fnv.NewConsistentHash(DefaultReplicas)
	for _, node := range req.Nodes {
		if err := tempCh.Add(node); err != nil {
			writeJSON(w, http.StatusInternalServerError, api.ConsistentResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to add node %s: %v", node, err),
			})
			return
		}
	}
	node, err := tempCh.Get(req.Key)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.ConsistentResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, api.ConsistentResponse{
		Success: true,
		Node:    node,
	})
}

func addNodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.NodeOperationResponse{
			Success: false,
			Error:   "method not allowed, use POST",
		})
		return
	}
	var req api.NodeOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.NodeOperationResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}
	if req.Node == "" {
		writeJSON(w, http.StatusBadRequest, api.NodeOperationResponse{
			Success: false,
			Error:   "node is required",
		})
		return
	}
	if err := consistentHash.Add(req.Node); err != nil {
		writeJSON(w, http.StatusInternalServerError, api.NodeOperationResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to add node: %v", err),
		})
		return
	}
	writeJSON(w, http.StatusOK, api.NodeOperationResponse{
		Success: true,
		Nodes:   consistentHash.Nodes(),
	})
}

func removeNodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.NodeOperationResponse{
			Success: false,
			Error:   "method not allowed, use POST",
		})
		return
	}
	var req api.NodeOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.NodeOperationResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}
	if req.Node == "" {
		writeJSON(w, http.StatusBadRequest, api.NodeOperationResponse{
			Success: false,
			Error:   "node is required",
		})
		return
	}
	consistentHash.Remove(req.Node)
	writeJSON(w, http.StatusOK, api.NodeOperationResponse{
		Success: true,
		Nodes:   consistentHash.Nodes(),
	})
}

func listNodesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, api.NodeOperationResponse{
			Success: false,
			Error:   "method not allowed, use GET",
		})
		return
	}
	writeJSON(w, http.StatusOK, api.NodeOperationResponse{
		Success: true,
		Nodes:   consistentHash.Nodes(),
	})
}
