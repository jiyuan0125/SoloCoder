package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"ringbuffer/common"
	"ringbuffer/ringbuffer"
)

type bufferEntry struct {
	rb *ringbuffer.RingBuffer
}

type Server struct {
	mu       sync.RWMutex
	buffers  map[string]*bufferEntry
	nextID   int
}

func NewServer() *Server {
	return &Server{
		buffers: make(map[string]*bufferEntry),
		nextID:  1,
	}
}

func (s *Server) generateID() string {
	id := s.nextID
	s.nextID++
	return fmt.Sprintf("buf-%d", id)
}

func (s *Server) createBuffer(w http.ResponseWriter, r *http.Request) {
	var req common.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Capacity < 0 {
		http.Error(w, "capacity must be non-negative", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	id := s.generateID()
	rb := ringbuffer.New(req.Capacity)
	s.buffers[id] = &bufferEntry{rb: rb}
	s.mu.Unlock()

	resp := common.CreateResponse{
		ID:       id,
		Capacity: req.Capacity,
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) writeBuffer(w http.ResponseWriter, r *http.Request) {
	var req common.WriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	entry, ok := s.buffers[req.ID]
	s.mu.RUnlock()

	if !ok {
		resp := common.WriteResponse{
			Error: "buffer not found",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	var timeout time.Duration
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Millisecond
	} else if req.Timeout == 0 {
		timeout = 0
	} else {
		timeout = -1
	}

	written, err := entry.rb.Write(req.Data, timeout)

	resp := common.WriteResponse{
		Written: written,
	}
	if err != nil {
		resp.Error = err.Error()
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) readBuffer(w http.ResponseWriter, r *http.Request) {
	var req common.ReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Length <= 0 {
		http.Error(w, "length must be positive", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	entry, ok := s.buffers[req.ID]
	s.mu.RUnlock()

	if !ok {
		resp := common.ReadResponse{
			Error: "buffer not found",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	var timeout time.Duration
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Millisecond
	} else if req.Timeout == 0 {
		timeout = 0
	} else {
		timeout = -1
	}

	data := make([]byte, req.Length)
	read, err := entry.rb.Read(data, timeout)

	resp := common.ReadResponse{
		Data: data[:read],
		Read: read,
	}
	if err != nil {
		if errors.Is(err, io.EOF) {
			resp.EOF = true
		} else {
			resp.Error = err.Error()
		}
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) statusBuffer(w http.ResponseWriter, r *http.Request) {
	var req common.StatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	entry, ok := s.buffers[req.ID]
	s.mu.RUnlock()

	if !ok {
		resp := common.StatusResponse{
			Error: "buffer not found",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := common.StatusResponse{
		ID:       req.ID,
		Capacity: entry.rb.Capacity(),
		Size:     entry.rb.Size(),
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) closeBuffer(w http.ResponseWriter, r *http.Request) {
	var req common.CloseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	entry, ok := s.buffers[req.ID]
	s.mu.RUnlock()

	if !ok {
		resp := common.CloseResponse{
			Error: "buffer not found",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	entry.rb.Close()

	resp := common.CloseResponse{}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) listBuffers(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	buffers := make([]common.StatusResponse, 0, len(s.buffers))
	for id, entry := range s.buffers {
		buffers = append(buffers, common.StatusResponse{
			ID:       id,
			Capacity: entry.rb.Capacity(),
			Size:     entry.rb.Size(),
		})
	}
	s.mu.RUnlock()

	resp := common.ListResponse{
		Buffers: buffers,
	}
	json.NewEncoder(w).Encode(resp)
}

func main() {
	server := NewServer()

	http.HandleFunc("/create", server.createBuffer)
	http.HandleFunc("/write", server.writeBuffer)
	http.HandleFunc("/read", server.readBuffer)
	http.HandleFunc("/status", server.statusBuffer)
	http.HandleFunc("/close", server.closeBuffer)
	http.HandleFunc("/list", server.listBuffers)

	port := 8080
	fmt.Printf("Server starting on port %d...\n", port)
	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
