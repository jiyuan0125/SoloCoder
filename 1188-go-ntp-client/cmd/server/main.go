package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/ntp-client/pkg/api"
	"github.com/ntp-client/pkg/ntp"
)

const defaultPort = 8202
const maxHistory = 100

type server struct {
	history []api.HistoryEntry
	mu      sync.RWMutex
}

func main() {
	port := defaultPort

	flag.IntVar(&port, "port", defaultPort, "HTTP server port")
	flag.Parse()

	if envPort := os.Getenv("NTP_SERVER_PORT"); envPort != "" {
		var p int
		if _, err := fmt.Sscanf(envPort, "%d", &p); err == nil && p > 0 {
			port = p
		}
	}

	s := &server{
		history: make([]api.HistoryEntry, 0, maxHistory),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/query", s.handleQuery)
	mux.HandleFunc("/sync", s.handleSync)
	mux.HandleFunc("/history", s.handleHistory)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("NTP server listening on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func (s *server) handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Server == "" {
		sendError(w, http.StatusBadRequest, "server is required")
		return
	}

	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	result, err := ntp.Query(req.Server, req.Port, timeout)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := api.QueryResponse{
		Server:       req.Server,
		ServerTime:   result.ServerTime,
		Offset:       result.Offset,
		Delay:        result.Delay,
		Stratum:      result.Stratum,
		PollSec:      ntp.PollSec(result.Poll),
		PrecisionSec: ntp.PrecisionSec(result.Precision),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *server) handleSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Server == "" {
		sendError(w, http.StatusBadRequest, "server is required")
		return
	}

	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	result, err := ntp.Query(req.Server, req.Port, timeout)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	localTime := time.Now()
	entry := api.HistoryEntry{
		Timestamp: time.Now(),
		Server:    req.Server,
		Offset:    result.Offset,
		Delay:     result.Delay,
		Stratum:   result.Stratum,
	}
	s.addHistory(entry)

	adjusted, err := setSystemTime(result.ServerTime)
	msg := "system time adjusted successfully"
	if err != nil {
		msg = fmt.Sprintf("adjustment failed (may require root privileges): %v", err)
	} else if !adjusted {
		msg = "time adjustment not supported on this platform"
	}

	resp := api.SyncResponse{
		Server:     req.Server,
		ServerTime: result.ServerTime,
		LocalTime:  localTime,
		Offset:     result.Offset,
		Delay:      result.Delay,
		Adjusted:   adjusted,
		Message:    msg,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *server) handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	entries := make([]api.HistoryEntry, len(s.history))
	copy(entries, s.history)
	s.mu.RUnlock()

	resp := api.HistoryResponse{Entries: entries}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *server) addHistory(entry api.HistoryEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.history = append(s.history, entry)
	if len(s.history) > maxHistory {
		s.history = s.history[len(s.history)-maxHistory:]
	}
}

func sendError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(api.ErrorResponse{Error: msg})
}
