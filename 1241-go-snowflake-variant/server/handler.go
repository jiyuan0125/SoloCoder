package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"idgen/api"
	"idgen/idgen"
)

type Server struct {
	nodeManager *idgen.NodeManager
}

func NewServer() *Server {
	return &Server{
		nodeManager: idgen.NewNodeManager(),
	}
}

func (s *Server) RegisterNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.RegisterNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "node name is required")
		return
	}

	nodeID, err := s.nodeManager.RegisterNode(req.Name, req.Remark, time.Now().UnixMilli())
	if err != nil {
		if errors.Is(err, idgen.ErrNodeAlreadyExists) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, idgen.ErrNoAvailableNodeID) {
			writeError(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, api.RegisterNodeResponse{
		Success: true,
		NodeID:  nodeID,
	})
}

func (s *Server) UnregisterNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.UnregisterNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.NodeID == api.ReservedNodeID {
		writeError(w, http.StatusBadRequest, "invalid node ID")
		return
	}

	if err := s.nodeManager.UnregisterNode(req.NodeID); err != nil {
		if errors.Is(err, idgen.ErrNodeNotRegistered) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, api.UnregisterNodeResponse{
		Success: true,
	})
}

func (s *Server) GenerateID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.GenerateIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	node, exists := s.nodeManager.GetNode(req.NodeID)
	if !exists {
		writeError(w, http.StatusNotFound, "node not registered")
		return
	}

	id, err := node.Generator.NextID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, api.GenerateIDResponse{
		Success: true,
		ID:      id,
	})
}

func (s *Server) BatchGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.BatchGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Count <= 0 || req.Count > api.MaxBatchCount {
		writeError(w, http.StatusBadRequest, "count must be between 1 and 1000")
		return
	}

	node, exists := s.nodeManager.GetNode(req.NodeID)
	if !exists {
		writeError(w, http.StatusNotFound, "node not registered")
		return
	}

	ids, err := node.Generator.BatchNextID(req.Count)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, api.BatchGenerateResponse{
		Success: true,
		IDs:     ids,
	})
}

func (s *Server) ListNodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	nodes := s.nodeManager.ListNodes()
	nodeInfos := make([]*api.NodeInfo, 0, len(nodes))
	for _, node := range nodes {
		nodeInfos = append(nodeInfos, &api.NodeInfo{
			NodeID:       node.NodeID,
			Name:         node.Name,
			Remark:       node.Remark,
			RegisteredAt: node.RegisteredAt,
		})
	}

	writeJSON(w, http.StatusOK, api.ListNodesResponse{
		Success: true,
		Nodes:   nodeInfos,
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, api.ErrorResponse{
		Success: false,
		Message: message,
	})
}
