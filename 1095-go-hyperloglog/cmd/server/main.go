package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/example/hyperloglog/pkg/api"
	"github.com/example/hyperloglog/pkg/hll"
)

type HLLServer struct {
	hlls map[string]*hll.HLL
	mu   sync.RWMutex
}

func NewHLLServer() *HLLServer {
	return &HLLServer{
		hlls: make(map[string]*hll.HLL),
	}
}

func (s *HLLServer) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(api.ErrorResponse{
		Success: false,
		Message: message,
	})
}

func (s *HLLServer) writeSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func (s *HLLServer) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req api.CreateHLLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	if req.Name == "" {
		s.writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.hlls[req.Name]; exists {
		s.writeError(w, http.StatusConflict, fmt.Sprintf("HLL instance '%s' already exists", req.Name))
		return
	}

	hllInstance, err := hll.NewHLL(req.Precision)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.hlls[req.Name] = hllInstance
	s.writeSuccess(w, api.CreateHLLResponse{
		Success: true,
		Message: fmt.Sprintf("HLL instance '%s' created with precision %d", req.Name, req.Precision),
	})
}

func (s *HLLServer) handleAdd(w http.ResponseWriter, r *http.Request) {
	var req api.AddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	if req.Name == "" {
		s.writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	s.mu.RLock()
	hllInstance, exists := s.hlls[req.Name]
	s.mu.RUnlock()

	if !exists {
		s.writeError(w, http.StatusNotFound, fmt.Sprintf("HLL instance '%s' not found", req.Name))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, value := range req.Values {
		hllInstance.Add(value)
	}

	s.writeSuccess(w, api.AddResponse{
		Success: true,
		Message: fmt.Sprintf("added %d values", len(req.Values)),
	})
}

func (s *HLLServer) handleCount(w http.ResponseWriter, r *http.Request) {
	var req api.CountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	if req.Name == "" {
		s.writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	s.mu.RLock()
	hllInstance, exists := s.hlls[req.Name]
	s.mu.RUnlock()

	if !exists {
		s.writeError(w, http.StatusNotFound, fmt.Sprintf("HLL instance '%s' not found", req.Name))
		return
	}

	count := hllInstance.Count()

	s.writeSuccess(w, api.CountResponse{
		Success: true,
		Count:   count,
	})
}

func (s *HLLServer) handleMerge(w http.ResponseWriter, r *http.Request) {
	var req api.MergeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	if req.NewName == "" {
		s.writeError(w, http.StatusBadRequest, "new_name is required")
		return
	}

	if len(req.Sources) < 2 {
		s.writeError(w, http.StatusBadRequest, "at least 2 source HLLs are required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.hlls[req.NewName]; exists {
		s.writeError(w, http.StatusConflict, fmt.Sprintf("HLL instance '%s' already exists", req.NewName))
		return
	}

	var precision int
	for i, name := range req.Sources {
		hllInstance, exists := s.hlls[name]
		if !exists {
			s.writeError(w, http.StatusNotFound, fmt.Sprintf("HLL instance '%s' not found", name))
			return
		}

		if i == 0 {
			precision = hllInstance.Precision()
		} else if hllInstance.Precision() != precision {
			s.writeError(w, http.StatusBadRequest, fmt.Sprintf("precision mismatch: '%s' has different precision", name))
			return
		}
	}

	serialized, err := s.hlls[req.Sources[0]].Serialize()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to serialize: %v", err))
		return
	}

	newHLL, err := hll.Deserialize(serialized)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to deserialize: %v", err))
		return
	}

	for i := 1; i < len(req.Sources); i++ {
		if err := newHLL.Merge(s.hlls[req.Sources[i]]); err != nil {
			s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to merge: %v", err))
			return
		}
	}

	s.hlls[req.NewName] = newHLL
	s.writeSuccess(w, api.MergeResponse{
		Success: true,
		Message: fmt.Sprintf("merged %d HLLs into '%s'", len(req.Sources), req.NewName),
	})
}

func (s *HLLServer) handleList(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	infos := make([]api.HLLInstanceInfo, 0, len(s.hlls))
	for name, hllInstance := range s.hlls {
		infos = append(infos, api.HLLInstanceInfo{
			Name:      name,
			Precision: hllInstance.Precision(),
			Count:     hllInstance.Count(),
		})
	}

	s.writeSuccess(w, api.ListResponse{
		Success: true,
		HLLs:    infos,
	})
}

func (s *HLLServer) handleExport(w http.ResponseWriter, r *http.Request) {
	var req api.ExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	if req.Name == "" {
		s.writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	s.mu.RLock()
	hllInstance, exists := s.hlls[req.Name]
	s.mu.RUnlock()

	if !exists {
		s.writeError(w, http.StatusNotFound, fmt.Sprintf("HLL instance '%s' not found", req.Name))
		return
	}

	data, err := hllInstance.Serialize()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to serialize: %v", err))
		return
	}

	s.writeSuccess(w, api.ExportResponse{
		Success: true,
		Data:    data,
	})
}

func (s *HLLServer) handleImport(w http.ResponseWriter, r *http.Request) {
	var req api.ImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	if req.Name == "" {
		s.writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.hlls[req.Name]; exists {
		s.writeError(w, http.StatusConflict, fmt.Sprintf("HLL instance '%s' already exists", req.Name))
		return
	}

	hllInstance, err := hll.Deserialize(req.Data)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("failed to deserialize: %v", err))
		return
	}

	s.hlls[req.Name] = hllInstance
	s.writeSuccess(w, api.ImportResponse{
		Success: true,
		Message: fmt.Sprintf("HLL instance '%s' imported successfully", req.Name),
	})
}

func main() {
	server := NewHLLServer()

	http.HandleFunc("/api/hll/create", server.handleCreate)
	http.HandleFunc("/api/hll/add", server.handleAdd)
	http.HandleFunc("/api/hll/count", server.handleCount)
	http.HandleFunc("/api/hll/merge", server.handleMerge)
	http.HandleFunc("/api/hll/list", server.handleList)
	http.HandleFunc("/api/hll/export", server.handleExport)
	http.HandleFunc("/api/hll/import", server.handleImport)

	fmt.Println("HLL Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
