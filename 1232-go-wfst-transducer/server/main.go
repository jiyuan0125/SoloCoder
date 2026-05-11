package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"wfst/api"
	"wfst/wfst"

	"github.com/google/uuid"
)

type WFSTStore struct {
	sync.RWMutex
	wfsts map[string]*wfst.WFST
}

func NewWFSTStore() *WFSTStore {
	return &WFSTStore{
		wfsts: make(map[string]*wfst.WFST),
	}
}

func (s *WFSTStore) Create(f *wfst.WFST) string {
	s.Lock()
	defer s.Unlock()
	id := uuid.New().String()
	s.wfsts[id] = f
	return id
}

func (s *WFSTStore) Get(id string) *wfst.WFST {
	s.RLock()
	defer s.RUnlock()
	return s.wfsts[id]
}

func (s *WFSTStore) Count() int {
	s.RLock()
	defer s.RUnlock()
	return len(s.wfsts)
}

var store = NewWFSTStore()

func countArcs(f *wfst.WFST) int {
	count := 0
	for _, state := range f.States() {
		count += len(f.GetArcs(state))
	}
	return count
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(api.ErrorResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req api.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.ErrorResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
		return
	}

	f := wfst.New()
	
	maxStateID := 0
	
	if len(req.States) > 0 {
		for _, state := range req.States {
			if state > maxStateID {
				maxStateID = state
			}
		}
	}
	
	for _, arc := range req.Arcs {
		if arc.From > maxStateID {
			maxStateID = arc.From
		}
		if arc.Next > maxStateID {
			maxStateID = arc.Next
		}
	}
	
	for f.NumStates() <= maxStateID {
		f.AddState()
	}

	for _, arc := range req.Arcs {
		wfstArc := wfst.Arc{
			Input:  arc.Input,
			Output: arc.Output,
			Weight: arc.Weight,
			Next:   arc.Next,
		}
		f.AddArc(arc.From, wfstArc)
	}

	id := store.Create(f)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.CreateResponse{
		Success:   true,
		WFSTID:    id,
		NumStates: f.NumStates(),
		NumArcs:   countArcs(f),
		Message:   "WFST created successfully",
	})
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(api.ErrorResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req api.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.ErrorResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
		return
	}

	f := store.Get(req.WFSTID)
	if f == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(api.ErrorResponse{
			Success: false,
			Error:   "WFST not found: " + req.WFSTID,
		})
		return
	}

	result := wfst.Viterbi(f, req.Input)

	w.Header().Set("Content-Type", "application/json")
	if result == nil {
		json.NewEncoder(w).Encode(api.SearchResponse{
			Success:     false,
			Message:     "no valid path found for input sequence",
		})
		return
	}

	json.NewEncoder(w).Encode(api.SearchResponse{
		Success:     true,
		Path:        result.Path,
		InputSeq:    result.InputSeq,
		OutputSeq:   result.OutputSeq,
		Weights:     result.Weights,
		TotalWeight: result.TotalWeight,
		Message:     "best path found",
	})
}

func composeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(api.ErrorResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req api.ComposeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.ErrorResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
		return
	}

	a := store.Get(req.WFSTIDA)
	if a == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(api.ErrorResponse{
			Success: false,
			Error:   "WFST A not found: " + req.WFSTIDA,
		})
		return
	}

	b := store.Get(req.WFSTIDB)
	if b == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(api.ErrorResponse{
			Success: false,
			Error:   "WFST B not found: " + req.WFSTIDB,
		})
		return
	}

	composed := wfst.Compose(a, b)
	id := store.Create(composed)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.ComposeResponse{
		Success:   true,
		WFSTID:    id,
		NumStates: composed.NumStates(),
		NumArcs:   countArcs(composed),
		Message:   "WFST composed successfully",
	})
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8213, "HTTP server port")
	flag.Parse()

	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	http.HandleFunc("/wfst/create", createHandler)
	http.HandleFunc("/wfst/search", searchHandler)
	http.HandleFunc("/wfst/compose", composeHandler)

	log.Printf("WFST server starting on port %d...", port)
	log.Printf("Total WFSTs in store: %d", store.Count())
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
