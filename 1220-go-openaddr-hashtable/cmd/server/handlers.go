package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	
	"hashtable/pkg/api"
	"hashtable/pkg/hashtable"
)

func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, api.ErrorResponse{Success: false, Error: "Method not allowed"})
		return
	}
	
	var req api.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, api.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	
	method, err := hashtable.ParseProbeMethod(req.Method)
	if err != nil {
		sendJSON(w, http.StatusBadRequest, api.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	
	if req.Capacity < 1 {
		sendJSON(w, http.StatusBadRequest, api.ErrorResponse{Success: false, Error: "capacity must be at least 1"})
		return
	}
	
	s.mutex.Lock()
	s.ht = hashtable.NewHashTable(method, req.Capacity)
	s.mutex.Unlock()
	
	capacity := s.ht.Capacity()
	sendJSON(w, http.StatusOK, api.CreateResponse{
		Success: true,
		Message: fmt.Sprintf("Created hash table with %s probing and capacity %d", method.String(), capacity),
	})
}

func (s *Server) handlePut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, api.ErrorResponse{Success: false, Error: "Method not allowed"})
		return
	}
	
	var req api.PutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, api.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	
	if req.Key == "" {
		sendJSON(w, http.StatusBadRequest, api.ErrorResponse{Success: false, Error: "key cannot be empty"})
		return
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if s.ht == nil {
		sendJSON(w, http.StatusBadRequest, api.ErrorResponse{Success: false, Error: "Hash table not created. Call /create first."})
		return
	}
	
	if err := s.ht.Put(req.Key, req.Value); err != nil {
		sendJSON(w, http.StatusInternalServerError, api.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	
	sendJSON(w, http.StatusOK, api.PutResponse{
		Success: true,
		Message: "Put successful",
	})
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, api.ErrorResponse{Success: false, Error: "Method not allowed"})
		return
	}
	
	var req api.GetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, api.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	if s.ht == nil {
		sendJSON(w, http.StatusBadRequest, api.ErrorResponse{Success: false, Error: "Hash table not created. Call /create first."})
		return
	}
	
	value, err := s.ht.Get(req.Key)
	if err != nil {
		sendJSON(w, http.StatusNotFound, api.GetResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	
	sendJSON(w, http.StatusOK, api.GetResponse{
		Success: true,
		Value:   value,
	})
}

func (s *Server) handleRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, api.ErrorResponse{Success: false, Error: "Method not allowed"})
		return
	}
	
	var req api.RemoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, api.ErrorResponse{Success: false, Error: err.Error()})
		return
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if s.ht == nil {
		sendJSON(w, http.StatusBadRequest, api.ErrorResponse{Success: false, Error: "Hash table not created. Call /create first."})
		return
	}
	
	if err := s.ht.Remove(req.Key); err != nil {
		sendJSON(w, http.StatusNotFound, api.RemoveResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	
	sendJSON(w, http.StatusOK, api.RemoveResponse{
		Success: true,
		Message: "Remove successful",
	})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, api.ErrorResponse{Success: false, Error: "Method not allowed"})
		return
	}
	
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	if s.ht == nil {
		sendJSON(w, http.StatusBadRequest, api.ErrorResponse{Success: false, Error: "Hash table not created. Call /create first."})
		return
	}
	
	sendJSON(w, http.StatusOK, api.StatsResponse{
		Success:     true,
		Size:        s.ht.Size(),
		Tombstones:  s.ht.Tombstones(),
		Capacity:    s.ht.Capacity(),
		MaxProbeLen: s.ht.MaxProbeLen(),
	})
}
