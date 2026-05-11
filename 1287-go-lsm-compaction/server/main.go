package main

import (
	"encoding/json"
	"flag"
	"lsm-tree/common"
	"lsm-tree/lsm"
	"log"
	"net/http"
	"os"
	"strings"
)

type Server struct {
	tree *lsm.LSMTree
}

func NewServer(tree *lsm.LSMTree) *Server {
	return &Server{tree: tree}
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) readJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func (s *Server) PutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeJSON(w, http.StatusMethodNotAllowed, common.PutResponse{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}
	
	var req common.PutRequest
	if err := s.readJSON(r, &req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, common.PutResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}
	
	if req.Key == "" {
		s.writeJSON(w, http.StatusBadRequest, common.PutResponse{
			Success: false,
			Error:   "Key cannot be empty",
		})
		return
	}
	
	if err := s.tree.Put(req.Key, req.Value); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, common.PutResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	
	s.writeJSON(w, http.StatusOK, common.PutResponse{Success: true})
}

func (s *Server) GetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeJSON(w, http.StatusMethodNotAllowed, common.GetResponse{
			Exists: false,
			Error:  "Method not allowed",
		})
		return
	}
	
	key := r.URL.Query().Get("key")
	if key == "" {
		s.writeJSON(w, http.StatusBadRequest, common.GetResponse{
			Exists: false,
			Error:  "Key is required",
		})
		return
	}
	
	value, exists, err := s.tree.Get(key)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, common.GetResponse{
			Exists: false,
			Error:  err.Error(),
		})
		return
	}
	
	s.writeJSON(w, http.StatusOK, common.GetResponse{
		Exists: exists,
		Value:  value,
	})
}

func (s *Server) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeJSON(w, http.StatusMethodNotAllowed, common.DeleteResponse{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}
	
	var req common.DeleteRequest
	if err := s.readJSON(r, &req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, common.DeleteResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}
	
	if req.Key == "" {
		s.writeJSON(w, http.StatusBadRequest, common.DeleteResponse{
			Success: false,
			Error:   "Key cannot be empty",
		})
		return
	}
	
	if err := s.tree.Delete(req.Key); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, common.DeleteResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	
	s.writeJSON(w, http.StatusOK, common.DeleteResponse{Success: true})
}

func (s *Server) RangeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeJSON(w, http.StatusMethodNotAllowed, common.RangeResponse{
			Error: "Method not allowed",
		})
		return
	}
	
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")
	
	entries, err := s.tree.Range(start, end)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, common.RangeResponse{
			Error: err.Error(),
		})
		return
	}
	
	results := make([]common.KVPair, 0)
	for _, e := range entries {
		if !e.Deleted {
			results = append(results, common.KVPair{Key: e.Key, Value: e.Value})
		}
	}
	
	s.writeJSON(w, http.StatusOK, common.RangeResponse{Results: results})
}

func (s *Server) CompactHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeJSON(w, http.StatusMethodNotAllowed, common.CompactResponse{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}
	
	if err := s.tree.Compact(); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, common.CompactResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	
	s.writeJSON(w, http.StatusOK, common.CompactResponse{Success: true})
}

func (s *Server) SwitchStrategyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeJSON(w, http.StatusMethodNotAllowed, common.SwitchStrategyResponse{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}
	
	var req common.SwitchStrategyRequest
	if err := s.readJSON(r, &req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, common.SwitchStrategyResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}
	
	if req.Strategy != common.SizeTieredStrategy && req.Strategy != common.LeveledStrategy {
		s.writeJSON(w, http.StatusBadRequest, common.SwitchStrategyResponse{
			Success: false,
			Error:   "Invalid strategy",
		})
		return
	}
	
	if err := s.tree.SwitchStrategy(req.Strategy); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, common.SwitchStrategyResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	
	s.writeJSON(w, http.StatusOK, common.SwitchStrategyResponse{
		Success:  true,
		Strategy: req.Strategy,
	})
}

func (s *Server) StatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeJSON(w, http.StatusMethodNotAllowed, common.StatsResponse{})
		return
	}
	
	levelStats := s.tree.GetLevelStats()
	stats := s.tree.GetStats()
	total, user, wa := stats.GetStats()
	
	s.writeJSON(w, http.StatusOK, common.StatsResponse{
		Levels:             levelStats,
		Strategy:           string(s.tree.GetStrategy()),
		WriteAmplification: wa,
		TotalWrites:        total,
		UserWrites:         user,
	})
}

func (s *Server) ConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeJSON(w, http.StatusMethodNotAllowed, common.ConfigResponse{})
		return
	}
	
	config := s.tree.Config()
	
	s.writeJSON(w, http.StatusOK, common.ConfigResponse{
		Current: common.Config{
			MemTableSize: config.MemTableSize,
			Strategy:     config.Strategy,
			DataDir:      config.DataDir,
		},
	})
}

func (s *Server) BatchPutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeJSON(w, http.StatusMethodNotAllowed, common.BatchPutResponse{
			Error: "Method not allowed",
		})
		return
	}
	
	var req common.BatchPutRequest
	if err := s.readJSON(r, &req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, common.BatchPutResponse{
			Error: "Invalid request: " + err.Error(),
		})
		return
	}
	
	successCount := 0
	for _, pair := range req.Pairs {
		if pair.Key == "" {
			continue
		}
		if err := s.tree.Put(pair.Key, pair.Value); err != nil {
			continue
		}
		successCount++
	}
	
	s.writeJSON(w, http.StatusOK, common.BatchPutResponse{
		SuccessCount: successCount,
	})
}

func getPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "8514"
}

func main() {
	portFlag := flag.String("port", "", "Server port (overrides PORT env var)")
	memTableSize := flag.Int("memtable-size", lsm.DefaultConfig.MemTableSize, "MemTable size in bytes")
	dataDir := flag.String("data-dir", lsm.DefaultConfig.DataDir, "Data directory")
	strategyFlag := flag.String("strategy", string(common.SizeTieredStrategy), "Compaction strategy: size-tiered or leveled")
	flag.Parse()
	
	port := *portFlag
	if port == "" {
		port = getPort()
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	
	strategy := common.CompactionStrategy(*strategyFlag)
	if strategy != common.SizeTieredStrategy && strategy != common.LeveledStrategy {
		log.Fatalf("Invalid strategy: %s. Use 'size-tiered' or 'leveled'", strategy)
	}
	
	config := lsm.DefaultConfig
	config.MemTableSize = *memTableSize
	config.DataDir = *dataDir
	config.Strategy = strategy
	
	tree, err := lsm.NewLSMTree(config)
	if err != nil {
		log.Fatalf("Failed to create LSM tree: %v", err)
	}
	defer tree.Close()
	
	server := NewServer(tree)
	
	mux := http.NewServeMux()
	mux.HandleFunc("/put", server.PutHandler)
	mux.HandleFunc("/get", server.GetHandler)
	mux.HandleFunc("/delete", server.DeleteHandler)
	mux.HandleFunc("/range", server.RangeHandler)
	mux.HandleFunc("/compact", server.CompactHandler)
	mux.HandleFunc("/strategy", server.SwitchStrategyHandler)
	mux.HandleFunc("/stats", server.StatsHandler)
	mux.HandleFunc("/config", server.ConfigHandler)
	mux.HandleFunc("/batch", server.BatchPutHandler)
	
	log.Printf("LSM Tree server starting on port %s", port)
	log.Printf("Config: MemTableSize=%d, DataDir=%s, Strategy=%s",
		config.MemTableSize, config.DataDir, config.Strategy)
	
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
