package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/example/gossip-simulator/pkg/api"
	"github.com/example/gossip-simulator/pkg/gossip"
)

type Server struct {
	simulator *gossip.Simulator
}

func NewServer() *Server {
	return &Server{
		simulator: gossip.NewSimulator(),
	}
}

func (s *Server) HandleAddNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req api.AddNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.NodeID == "" {
		sendError(w, "node_id is required", http.StatusBadRequest)
		return
	}
	
	s.simulator.AddNode(req.NodeID)
	
	resp := api.SuccessResponse{
		Success: true,
		Message: fmt.Sprintf("Node %s added", req.NodeID),
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) HandleRemoveNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req api.RemoveNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.NodeID == "" {
		sendError(w, "node_id is required", http.StatusBadRequest)
		return
	}
	
	s.simulator.RemoveNode(req.NodeID)
	
	resp := api.SuccessResponse{
		Success: true,
		Message: fmt.Sprintf("Node %s removed", req.NodeID),
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) HandleInjectData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req api.InjectDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.NodeID == "" || req.Key == "" {
		sendError(w, "node_id and key are required", http.StatusBadRequest)
		return
	}
	
	s.simulator.InjectData(req.NodeID, req.Key, req.Value)
	
	resp := api.SuccessResponse{
		Success: true,
		Message: fmt.Sprintf("Data injected into node %s", req.NodeID),
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) HandleStep(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req api.StepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Steps = 1
	}
	
	if req.Steps <= 0 {
		req.Steps = 1
	}
	
	for i := 0; i < req.Steps; i++ {
		s.simulator.Step()
	}
	
	resp := api.StepResponse{
		Success:     true,
		StepsDone: req.Steps,
		Consistency: s.simulator.GetConsistency(),
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) HandleGetState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodGet {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	state := s.simulator.GetState()
	
	resp := api.StateResponse{
		Round: state["round"].(int),
		Nodes: make(map[string]api.NodeState),
	}
	
	nodesMap := state["nodes"].(map[string]interface{})
	for nodeID, nodeData := range nodesMap {
		nodeMap := nodeData.(map[string]interface{})
		nodeState := api.NodeState{
			Status:  nodeMap["status"].(string),
			Storage: make(map[string]api.NodeStorageItem),
		}
		
		storageMap := nodeMap["storage"].(map[string]interface{})
		for k, v := range storageMap {
			itemMap := v.(map[string]interface{})
			nodeState.Storage[k] = api.NodeStorageItem{
				Value:   itemMap["value"].(string),
				Version: itemMap["version"].(int64),
			}
		}
		
		resp.Nodes[nodeID] = nodeState
	}
	
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) HandleGetConsistency(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodGet {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	resp := api.ConsistencyResponse{
		Consistency: s.simulator.GetConsistency(),
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) HandleSimulateFailure(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req api.SimulateFailureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.NodeID == "" {
		sendError(w, "node_id is required", http.StatusBadRequest)
		return
	}
	
	s.simulator.SimulateNodeFailure(req.NodeID)
	
	resp := api.SuccessResponse{
		Success: true,
		Message: fmt.Sprintf("Node %s failure simulated", req.NodeID),
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) HandleSimulateRecovery(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req api.SimulateRecoveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.NodeID == "" {
		sendError(w, "node_id is required", http.StatusBadRequest)
		return
	}
	
	s.simulator.SimulateNodeRecovery(req.NodeID)
	
	resp := api.SuccessResponse{
		Success: true,
		Message: fmt.Sprintf("Node %s recovery simulated", req.NodeID),
	}
	json.NewEncoder(w).Encode(resp)
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	resp := api.ErrorResponse{
		Success: false,
		Error:   message,
	}
	json.NewEncoder(w).Encode(resp)
}

func main() {
	server := NewServer()
	
	http.HandleFunc("/nodes/add", server.HandleAddNode)
	http.HandleFunc("/nodes/remove", server.HandleRemoveNode)
	http.HandleFunc("/data/inject", server.HandleInjectData)
	http.HandleFunc("/step", server.HandleStep)
	http.HandleFunc("/state", server.HandleGetState)
	http.HandleFunc("/consistency", server.HandleGetConsistency)
	http.HandleFunc("/nodes/fail", server.HandleSimulateFailure)
	http.HandleFunc("/nodes/recover", server.HandleSimulateRecovery)
	
	port := 8080
	fmt.Printf("Gossip Simulator Server starting on port %d...\n", port)
	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
