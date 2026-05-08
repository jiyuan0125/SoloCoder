package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/example/crdt/pkg/api"
	"github.com/example/crdt/pkg/crdt"
	"github.com/google/uuid"
)

type CRDTInstance struct {
	ID        string
	Type      api.CRDTType
	GCounter  *crdt.GCounter
	PNCounter *crdt.PNCounter
	GSet      *crdt.GSet
	CreatedAt time.Time
}

type Server struct {
	instances map[string]*CRDTInstance
	mu        sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		instances: make(map[string]*CRDTInstance),
	}
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(api.ErrorResponse{Error: message})
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) createInstanceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.CreateInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Type == "" {
		s.writeError(w, http.StatusBadRequest, "CRDT type is required")
		return
	}

	instance := &CRDTInstance{
		ID:        uuid.New().String(),
		Type:      req.Type,
		CreatedAt: time.Now(),
	}

	switch req.Type {
	case api.TypeGCounter:
		instance.GCounter = crdt.NewGCounter()
	case api.TypePNCounter:
		instance.PNCounter = crdt.NewPNCounter()
	case api.TypeGSet:
		instance.GSet = crdt.NewGSet()
	default:
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Unknown CRDT type: %s", req.Type))
		return
	}

	s.mu.Lock()
	s.instances[instance.ID] = instance
	s.mu.Unlock()

	var state interface{}
	switch instance.Type {
	case api.TypeGCounter:
		state = api.GCounterToState(instance.GCounter)
	case api.TypePNCounter:
		state = api.PNCounterToState(instance.PNCounter)
	case api.TypeGSet:
		state = api.GSetToState(instance.GSet)
	}

	response := api.CreateInstanceResponse{
		ID:        instance.ID,
		Type:      instance.Type,
		State:     state,
		CreatedAt: instance.CreatedAt.Format(time.RFC3339),
	}

	s.writeJSON(w, http.StatusCreated, response)
}

func (s *Server) getInstance(id string) (*CRDTInstance, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	instance, exists := s.instances[id]
	return instance, exists
}

func (s *Server) getInstanceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id := r.URL.Path[len("/instances/"):]
	instance, exists := s.getInstance(id)
	if !exists {
		s.writeError(w, http.StatusNotFound, "Instance not found")
		return
	}

	response := s.buildInstanceResponse(instance)
	s.writeJSON(w, http.StatusOK, response)
}

func (s *Server) buildInstanceResponse(instance *CRDTInstance) api.GetInstanceResponse {
	var value int
	var elements []string
	var state interface{}

	switch instance.Type {
	case api.TypeGCounter:
		value = instance.GCounter.Value()
		state = api.GCounterToState(instance.GCounter)
	case api.TypePNCounter:
		value = instance.PNCounter.Value()
		state = api.PNCounterToState(instance.PNCounter)
	case api.TypeGSet:
		elements = instance.GSet.ElementsList()
		state = api.GSetToState(instance.GSet)
	}

	return api.GetInstanceResponse{
		ID:       instance.ID,
		Type:     instance.Type,
		Value:    value,
		Elements: elements,
		State:    state,
	}
}

func (s *Server) operationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id := r.URL.Path[len("/instances/"):]
	id = id[:len(id)-len("/operations")]

	instance, exists := s.getInstance(id)
	if !exists {
		s.writeError(w, http.StatusNotFound, "Instance not found")
		return
	}

	var req api.OperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var err error
	switch instance.Type {
	case api.TypeGCounter:
		err = s.handleGCounterOperation(instance, req)
	case api.TypePNCounter:
		err = s.handlePNCounterOperation(instance, req)
	case api.TypeGSet:
		err = s.handleGSetOperation(instance, req)
	}

	if err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := s.buildOperationResponse(instance)
	s.writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleGCounterOperation(instance *CRDTInstance, req api.OperationRequest) error {
	switch req.Operation {
	case "increment":
		if req.Delta > 0 {
			return instance.GCounter.IncrementBy(req.NodeID, req.Delta)
		}
		return instance.GCounter.Increment(req.NodeID)
	default:
		return fmt.Errorf("unknown operation for G-Counter: %s", req.Operation)
	}
}

func (s *Server) handlePNCounterOperation(instance *CRDTInstance, req api.OperationRequest) error {
	switch req.Operation {
	case "increment":
		if req.Delta > 0 {
			return instance.PNCounter.IncrementBy(req.NodeID, req.Delta)
		}
		return instance.PNCounter.Increment(req.NodeID)
	case "decrement":
		if req.Delta > 0 {
			return instance.PNCounter.DecrementBy(req.NodeID, req.Delta)
		}
		return instance.PNCounter.Decrement(req.NodeID)
	default:
		return fmt.Errorf("unknown operation for PN-Counter: %s", req.Operation)
	}
}

func (s *Server) handleGSetOperation(instance *CRDTInstance, req api.OperationRequest) error {
	switch req.Operation {
	case "add":
		if req.Element == "" {
			return fmt.Errorf("element is required for add operation")
		}
		instance.GSet.Add(req.Element)
		return nil
	default:
		return fmt.Errorf("unknown operation for G-Set: %s", req.Operation)
	}
}

func (s *Server) buildOperationResponse(instance *CRDTInstance) api.OperationResponse {
	var value int
	var elements []string
	var state interface{}

	switch instance.Type {
	case api.TypeGCounter:
		value = instance.GCounter.Value()
		state = api.GCounterToState(instance.GCounter)
	case api.TypePNCounter:
		value = instance.PNCounter.Value()
		state = api.PNCounterToState(instance.PNCounter)
	case api.TypeGSet:
		elements = instance.GSet.ElementsList()
		state = api.GSetToState(instance.GSet)
	}

	return api.OperationResponse{
		ID:       instance.ID,
		Type:     instance.Type,
		Value:    value,
		Elements: elements,
		State:    state,
	}
}

func (s *Server) mergeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.MergeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.TargetID == "" || req.SourceID == "" {
		s.writeError(w, http.StatusBadRequest, "target_id and source_id are required")
		return
	}

	targetInstance, targetExists := s.getInstance(req.TargetID)
	sourceInstance, sourceExists := s.getInstance(req.SourceID)

	if !targetExists {
		s.writeError(w, http.StatusNotFound, "Target instance not found")
		return
	}
	if !sourceExists {
		s.writeError(w, http.StatusNotFound, "Source instance not found")
		return
	}

	if targetInstance.Type != sourceInstance.Type {
		s.writeError(w, http.StatusBadRequest, "Cannot merge different CRDT types")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.mergeInstances(targetInstance, sourceInstance)

	response := s.buildMergeResponse(targetInstance, req)
	s.writeJSON(w, http.StatusOK, response)
}

func (s *Server) mergeInstances(target, source *CRDTInstance) {
	switch target.Type {
	case api.TypeGCounter:
		merged := target.GCounter.Merge(source.GCounter)
		target.GCounter = merged
	case api.TypePNCounter:
		merged := target.PNCounter.Merge(source.PNCounter)
		target.PNCounter = merged
	case api.TypeGSet:
		merged := target.GSet.Merge(source.GSet)
		target.GSet = merged
	}
}

func (s *Server) buildMergeResponse(instance *CRDTInstance, req api.MergeRequest) api.MergeResponse {
	var value int
	var elements []string
	var state interface{}

	switch instance.Type {
	case api.TypeGCounter:
		value = instance.GCounter.Value()
		state = api.GCounterToState(instance.GCounter)
	case api.TypePNCounter:
		value = instance.PNCounter.Value()
		state = api.PNCounterToState(instance.PNCounter)
	case api.TypeGSet:
		elements = instance.GSet.ElementsList()
		state = api.GSetToState(instance.GSet)
	}

	return api.MergeResponse{
		TargetID: req.TargetID,
		SourceID: req.SourceID,
		Value:    value,
		Elements: elements,
		State:    state,
	}
}

func (s *Server) listInstancesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	response := make([]api.GetInstanceResponse, 0, len(s.instances))
	for _, instance := range s.instances {
		response = append(response, s.buildInstanceResponse(instance))
	}

	s.writeJSON(w, http.StatusOK, response)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": "1.0.0",
	})
}
