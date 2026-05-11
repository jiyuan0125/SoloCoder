package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"raft-sim/internal/api"
	"raft-sim/internal/raft"
)

var cluster = raft.NewCluster()

func main() {
	http.HandleFunc("/api/cluster/create", createClusterHandler)
	http.HandleFunc("/api/cluster/add-node", addNodeHandler)
	http.HandleFunc("/api/cluster/nodes", listNodesHandler)
	http.HandleFunc("/api/cluster/nodes/", getNodeHandler)
	http.HandleFunc("/api/log/submit", submitLogHandler)
	http.HandleFunc("/api/election/trigger", triggerElectionHandler)

	log.Println("Raft server starting on :8501")
	log.Fatal(http.ListenAndServe(":8501", nil))
}

func createClusterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.CreateClusterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, api.CreateClusterResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if req.NodeCount < 3 {
		sendJSON(w, http.StatusBadRequest, api.CreateClusterResponse{
			Success: false,
			Message: "Cluster must have at least 3 nodes",
		})
		return
	}

	nodeIDs := cluster.CreateNodes(req.NodeCount)
	cluster.StartAll()

	sendJSON(w, http.StatusOK, api.CreateClusterResponse{
		Success: true,
		Message: fmt.Sprintf("Created cluster with %d nodes", req.NodeCount),
		Nodes:   nodeIDs,
	})
}

func addNodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.AddNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, api.AddNodeResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if err := cluster.AddNode(req.NodeID); err != nil {
		sendJSON(w, http.StatusBadRequest, api.AddNodeResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	sendJSON(w, http.StatusOK, api.AddNodeResponse{
		Success: true,
		Message: fmt.Sprintf("Added node %d to cluster", req.NodeID),
		NodeID:  req.NodeID,
	})
}

func listNodesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nodeInfos := cluster.GetAllNodeInfos()
	var nodes []api.NodeInfo
	for _, info := range nodeInfos {
		nodes = append(nodes, convertNodeInfo(info))
	}

	sendJSON(w, http.StatusOK, api.ListNodesResponse{
		Success: true,
		Message: fmt.Sprintf("Found %d nodes", len(nodes)),
		Nodes:   nodes,
	})
}

func getNodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	nodeID, err := strconv.Atoi(parts[4])
	if err != nil {
		http.Error(w, "Invalid node ID", http.StatusBadRequest)
		return
	}

	info, exists := cluster.GetNodeInfo(nodeID)
	if !exists {
		sendJSON(w, http.StatusNotFound, api.GetNodeResponse{
			Success: false,
			Message: fmt.Sprintf("Node %d not found", nodeID),
		})
		return
	}

	sendJSON(w, http.StatusOK, api.GetNodeResponse{
		Success: true,
		Message: "Node found",
		Node:    convertNodeInfo(info),
	})
}

func submitLogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.SubmitLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, api.SubmitLogResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if req.Command == "" {
		sendJSON(w, http.StatusBadRequest, api.SubmitLogResponse{
			Success: false,
			Message: "Command cannot be empty",
		})
		return
	}

	term, index, err := cluster.SubmitLog(req.Command)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, api.SubmitLogResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	sendJSON(w, http.StatusOK, api.SubmitLogResponse{
		Success: true,
		Message: "Log submitted successfully",
		Term:    term,
		Index:   index,
	})
}

func triggerElectionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if cluster.NodeCount() == 0 {
		sendJSON(w, http.StatusBadRequest, api.TriggerElectionResponse{
			Success: false,
			Message: "No cluster exists. Create one first.",
		})
		return
	}

	sendJSON(w, http.StatusOK, api.TriggerElectionResponse{
		Success: true,
		Message: "Election triggered. Check nodes status for results.",
	})
}

func convertNodeInfo(info raft.NodeInfo) api.NodeInfo {
	logEntries := make([]api.LogEntry, len(info.LogEntries))
	for i, e := range info.LogEntries {
		logEntries[i] = api.LogEntry{
			Term:    e.Term,
			Command: e.Command,
		}
	}

	return api.NodeInfo{
		ID:              info.ID,
		State:           api.NodeState(info.State),
		CurrentTerm:     info.CurrentTerm,
		VotedFor:        info.VotedFor,
		CommitIndex:     info.CommitIndex,
		LastApplied:     info.LastApplied,
		LogEntries:      logEntries,
		LeaderID:        info.LeaderID,
		LastHeartbeatAt: info.LastHeartbeatAt,
	}
}

func sendJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
