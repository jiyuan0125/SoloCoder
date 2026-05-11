package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"raftlog/common"
	"raftlog/raftlog"
)

var cluster *raftlog.Cluster

func initCluster() {
	nodeInfos := []common.NodeInfo{
		{ID: "node1", Address: "localhost:8081"},
		{ID: "node2", Address: "localhost:8082"},
		{ID: "node3", Address: "localhost:8083"},
	}
	cluster = raftlog.NewCluster(nodeInfos)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.SubmitResponse{
			Success: false,
			Message: "method not allowed",
		})
		return
	}
	var req common.SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.SubmitResponse{
			Success: false,
			Message: fmt.Sprintf("invalid request: %v", err),
		})
		return
	}
	if err := cluster.Submit(req.NodeID, req.Command); err != nil {
		writeJSON(w, http.StatusNotFound, common.SubmitResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, common.SubmitResponse{
		Success: true,
		Message: "command submitted",
	})
}

func logHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.LogResponse{
			Success: false,
			Message: "method not allowed",
		})
		return
	}
	nodeID := r.URL.Query().Get("node_id")
	if nodeID == "" {
		writeJSON(w, http.StatusBadRequest, common.LogResponse{
			Success: false,
			Message: "node_id is required",
		})
		return
	}
	entries, err := cluster.GetLog(nodeID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, common.LogResponse{
			Success: false,
			NodeID:  nodeID,
			Message: err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, common.LogResponse{
		Success: true,
		NodeID:  nodeID,
		Entries: entries,
	})
}

func checkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.ConsistencyCheckResponse{
			Success: false,
			Message: "method not allowed",
		})
		return
	}
	node1 := r.URL.Query().Get("node_id1")
	node2 := r.URL.Query().Get("node_id2")
	if node1 == "" || node2 == "" {
		writeJSON(w, http.StatusBadRequest, common.ConsistencyCheckResponse{
			Success: false,
			Message: "node_id1 and node_id2 are required",
		})
		return
	}
	result, err := cluster.CheckConsistency(node1, node2)
	if err != nil {
		writeJSON(w, http.StatusNotFound, common.ConsistencyCheckResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	resp := common.ConsistencyCheckResponse{
		Success:      true,
		Consistent:   result.Consistent,
		MatchedCount: result.MatchedCount,
		Node1ID:      result.Node1ID,
		Node1Total:   result.Node1Total,
		Node2ID:      result.Node2ID,
		Node2Total:   result.Node2Total,
	}
	if !result.Consistent {
		resp.MismatchIndex = result.MismatchIndex
		resp.Node1Entry = result.Node1Entry
		resp.Node2Entry = result.Node2Entry
	}
	writeJSON(w, http.StatusOK, resp)
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.ResetResponse{
			Success: false,
			Message: "method not allowed",
		})
		return
	}
	var req common.ResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.ResetResponse{
			Success: false,
			Message: fmt.Sprintf("invalid request: %v", err),
		})
		return
	}
	if err := cluster.Reset(req.NodeID); err != nil {
		writeJSON(w, http.StatusNotFound, common.ResetResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, common.ResetResponse{
		Success: true,
		Message: "node log reset",
	})
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.StatusResponse{
			Success: false,
			Message: "method not allowed",
		})
		return
	}
	nodes := cluster.AllNodeStatus()
	writeJSON(w, http.StatusOK, common.StatusResponse{
		Success: true,
		Nodes:   nodes,
	})
}

func main() {
	var port string
	flag.StringVar(&port, "port", "", "server port (overrides env)")
	flag.Parse()

	if port == "" {
		port = os.Getenv("SERVER_PORT")
	}
	if port == "" {
		port = "8080"
	}

	initCluster()

	http.HandleFunc("/submit", submitHandler)
	http.HandleFunc("/log", logHandler)
	http.HandleFunc("/check", checkHandler)
	http.HandleFunc("/reset", resetHandler)
	http.HandleFunc("/status", statusHandler)

	addr := ":" + port
	log.Printf("server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
