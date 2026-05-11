package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/ip-manager/pkg/cidr"
	"github.com/ip-manager/pkg/common"
)

type IPManagerServer struct {
	pools map[string]*cidr.Pool
	mu    sync.RWMutex
}

func NewIPManagerServer() *IPManagerServer {
	return &IPManagerServer{
		pools: make(map[string]*cidr.Pool),
	}
}

func (s *IPManagerServer) handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.ParseResponse{Success: false, Error: "invalid request body"}
		json.NewEncoder(w).Encode(resp)
		return
	}

	parsed, err := cidr.Parse(req.CIDR)
	if err != nil {
		resp := common.ParseResponse{Success: false, Error: err.Error()}
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := common.ParseResponse{
		Success:     true,
		CIDR:        parsed.String(),
		Network:     parsed.Network().String(),
		Broadcast:   parsed.Broadcast().String(),
		FirstUsable: parsed.FirstUsable().String(),
		LastUsable:  parsed.LastUsable().String(),
		Prefix:      parsed.Prefix,
		Size:        parsed.Size(),
		Version:     fmt.Sprintf("IPv%d", parsed.Version),
	}

	json.NewEncoder(w).Encode(resp)
}

func (s *IPManagerServer) handleContains(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.ContainsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.ContainsResponse{Success: false, Error: "invalid request body"}
		json.NewEncoder(w).Encode(resp)
		return
	}

	parsed, err := cidr.Parse(req.CIDR)
	if err != nil {
		resp := common.ContainsResponse{Success: false, Error: err.Error()}
		json.NewEncoder(w).Encode(resp)
		return
	}

	ip := net.ParseIP(req.IP)
	if ip == nil {
		resp := common.ContainsResponse{Success: false, Error: "invalid IP address"}
		json.NewEncoder(w).Encode(resp)
		return
	}

	contained := parsed.Contains(ip)
	resp := common.ContainsResponse{Success: true, Contained: contained}
	json.NewEncoder(w).Encode(resp)
}

func (s *IPManagerServer) handleMerge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.MergeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.MergeResponse{Success: false, Error: "invalid request body"}
		json.NewEncoder(w).Encode(resp)
		return
	}

	var list cidr.CIDRList
	for _, cidrStr := range req.CIDRs {
		parsed, err := cidr.Parse(cidrStr)
		if err != nil {
			resp := common.MergeResponse{Success: false, Error: fmt.Sprintf("invalid CIDR %s: %v", cidrStr, err)}
			json.NewEncoder(w).Encode(resp)
			return
		}
		list = append(list, parsed)
	}

	merged := cidr.Merge(list)
	result := make([]string, len(merged))
	for i, m := range merged {
		result[i] = m.String()
	}

	resp := common.MergeResponse{Success: true, Merged: result}
	json.NewEncoder(w).Encode(resp)
}

func (s *IPManagerServer) handleCreatePool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.CreatePoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.CreatePoolResponse{Success: false, Error: "invalid request body"}
		json.NewEncoder(w).Encode(resp)
		return
	}

	parsed, err := cidr.Parse(req.CIDR)
	if err != nil {
		resp := common.CreatePoolResponse{Success: false, Error: err.Error()}
		json.NewEncoder(w).Encode(resp)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.pools[req.PoolID]; exists {
		resp := common.CreatePoolResponse{Success: false, Error: fmt.Sprintf("pool %s already exists", req.PoolID)}
		json.NewEncoder(w).Encode(resp)
		return
	}

	pool := cidr.NewPool()
	pool.AddAvailable(parsed)
	s.pools[req.PoolID] = pool

	resp := common.CreatePoolResponse{Success: true, PoolID: req.PoolID}
	json.NewEncoder(w).Encode(resp)
}

func (s *IPManagerServer) handlePoolStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.PoolStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.PoolStatusResponse{Success: false, Error: "invalid request body"}
		json.NewEncoder(w).Encode(resp)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	pool, exists := s.pools[req.PoolID]
	if !exists {
		resp := common.PoolStatusResponse{Success: false, Error: fmt.Sprintf("pool %s not found", req.PoolID)}
		json.NewEncoder(w).Encode(resp)
		return
	}

	available := make([]string, len(pool.Available))
	for i, c := range pool.Available {
		available[i] = c.String()
	}

	used := make([]string, len(pool.Used))
	for i, c := range pool.Used {
		used[i] = c.String()
	}

	resp := common.PoolStatusResponse{
		Success:   true,
		PoolID:    req.PoolID,
		Available: available,
		Used:      used,
	}

	json.NewEncoder(w).Encode(resp)
}

func (s *IPManagerServer) handleAllocate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.AllocateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.AllocateResponse{Success: false, Error: "invalid request body"}
		json.NewEncoder(w).Encode(resp)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	pool, exists := s.pools[req.PoolID]
	if !exists {
		resp := common.AllocateResponse{Success: false, Error: fmt.Sprintf("pool %s not found", req.PoolID)}
		json.NewEncoder(w).Encode(resp)
		return
	}

	allocated, err := pool.Allocate(req.Prefix)
	if err != nil {
		resp := common.AllocateResponse{Success: false, Error: err.Error()}
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := common.AllocateResponse{
		Success:   true,
		Allocated: allocated.String(),
		PoolID:    req.PoolID,
	}

	json.NewEncoder(w).Encode(resp)
}

func (s *IPManagerServer) handleRelease(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.ReleaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.ReleaseResponse{Success: false, Error: "invalid request body"}
		json.NewEncoder(w).Encode(resp)
		return
	}

	parsed, err := cidr.Parse(req.CIDR)
	if err != nil {
		resp := common.ReleaseResponse{Success: false, Error: err.Error()}
		json.NewEncoder(w).Encode(resp)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	pool, exists := s.pools[req.PoolID]
	if !exists {
		resp := common.ReleaseResponse{Success: false, Error: fmt.Sprintf("pool %s not found", req.PoolID)}
		json.NewEncoder(w).Encode(resp)
		return
	}

	if err := pool.Release(parsed); err != nil {
		resp := common.ReleaseResponse{Success: false, Error: err.Error()}
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := common.ReleaseResponse{Success: true, PoolID: req.PoolID}
	json.NewEncoder(w).Encode(resp)
}

func (s *IPManagerServer) handleExclude(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.ExcludeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.ExcludeResponse{Success: false, Error: "invalid request body"}
		json.NewEncoder(w).Encode(resp)
		return
	}

	parsed, err := cidr.Parse(req.CIDR)
	if err != nil {
		resp := common.ExcludeResponse{Success: false, Error: err.Error()}
		json.NewEncoder(w).Encode(resp)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	pool, exists := s.pools[req.PoolID]
	if !exists {
		resp := common.ExcludeResponse{Success: false, Error: fmt.Sprintf("pool %s not found", req.PoolID)}
		json.NewEncoder(w).Encode(resp)
		return
	}

	if err := pool.Exclude(parsed); err != nil {
		resp := common.ExcludeResponse{Success: false, Error: err.Error()}
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := common.ExcludeResponse{Success: true, PoolID: req.PoolID}
	json.NewEncoder(w).Encode(resp)
}

func main() {
	server := NewIPManagerServer()

	http.HandleFunc("/parse", server.handleParse)
	http.HandleFunc("/contains", server.handleContains)
	http.HandleFunc("/merge", server.handleMerge)
	http.HandleFunc("/pool/create", server.handleCreatePool)
	http.HandleFunc("/pool/status", server.handlePoolStatus)
	http.HandleFunc("/pool/allocate", server.handleAllocate)
	http.HandleFunc("/pool/release", server.handleRelease)
	http.HandleFunc("/pool/exclude", server.handleExclude)

	port := ":8104"
	log.Printf("IP Manager Server starting on %s...", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
