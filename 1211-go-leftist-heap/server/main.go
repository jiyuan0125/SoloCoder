package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"leftistheap/api"
	"leftistheap/leftistheap"
)

type HeapManager struct {
	heaps map[string]*leftistheap.Heap
	mu    sync.RWMutex
}

func NewHeapManager() *HeapManager {
	return &HeapManager{
		heaps: make(map[string]*leftistheap.Heap),
	}
}

func generateHeapID() string {
	rand.Seed(time.Now().UnixNano())
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func (hm *HeapManager) CreateHeap(w http.ResponseWriter, r *http.Request) {
	var req api.CreateHeapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var heapType leftistheap.HeapType
	if req.HeapType == "max" {
		heapType = leftistheap.MaxHeap
	} else {
		heapType = leftistheap.MinHeap
	}

	heap := leftistheap.NewHeap(heapType)
	heapID := generateHeapID()

	hm.mu.Lock()
	hm.heaps[heapID] = heap
	hm.mu.Unlock()

	resp := api.CreateHeapResponse{
		HeapID: heapID,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (hm *HeapManager) Insert(w http.ResponseWriter, r *http.Request) {
	var req api.InsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hm.mu.RLock()
	heap, exists := hm.heaps[req.HeapID]
	hm.mu.RUnlock()

	if !exists {
		resp := api.InsertResponse{
			Error: fmt.Sprintf("heap %s not found", req.HeapID),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(resp)
		return
	}

	hm.mu.Lock()
	heap.Insert(req.Value)
	hm.mu.Unlock()

	resp := api.InsertResponse{}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (hm *HeapManager) Top(w http.ResponseWriter, r *http.Request) {
	var req api.TopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hm.mu.RLock()
	heap, exists := hm.heaps[req.HeapID]
	hm.mu.RUnlock()

	if !exists {
		resp := api.TopResponse{
			Error: fmt.Sprintf("heap %s not found", req.HeapID),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(resp)
		return
	}

	hm.mu.RLock()
	value, err := heap.Top()
	hm.mu.RUnlock()

	resp := api.TopResponse{}
	if err != nil {
		resp.Error = err.Error()
	} else {
		v := value
		resp.Value = &v
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (hm *HeapManager) Pop(w http.ResponseWriter, r *http.Request) {
	var req api.PopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hm.mu.RLock()
	heap, exists := hm.heaps[req.HeapID]
	hm.mu.RUnlock()

	if !exists {
		resp := api.PopResponse{
			Error: fmt.Sprintf("heap %s not found", req.HeapID),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(resp)
		return
	}

	hm.mu.Lock()
	value, err := heap.Pop()
	hm.mu.Unlock()

	resp := api.PopResponse{}
	if err != nil {
		resp.Error = err.Error()
	} else {
		v := value
		resp.Value = &v
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (hm *HeapManager) Merge(w http.ResponseWriter, r *http.Request) {
	var req api.MergeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hm.mu.Lock()
	defer hm.mu.Unlock()

	heap1, exists1 := hm.heaps[req.HeapID1]
	heap2, exists2 := hm.heaps[req.HeapID2]

	if !exists1 || !exists2 {
		resp := api.MergeResponse{
			Error: "one or both heaps not found",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(resp)
		return
	}

	heap1.Merge(heap2)
	delete(hm.heaps, req.HeapID2)

	resp := api.MergeResponse{
		NewHeapID: req.HeapID1,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	var port string
	flag.StringVar(&port, "port", "", "Server port")
	flag.Parse()

	if port == "" {
		port = os.Getenv("HEAP_SERVER_PORT")
		if port == "" {
			port = "8080"
		}
	}

	manager := NewHeapManager()

	http.HandleFunc("/create", manager.CreateHeap)
	http.HandleFunc("/insert", manager.Insert)
	http.HandleFunc("/top", manager.Top)
	http.HandleFunc("/pop", manager.Pop)
	http.HandleFunc("/merge", manager.Merge)

	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}
