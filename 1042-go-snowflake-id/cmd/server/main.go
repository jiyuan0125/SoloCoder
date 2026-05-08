package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"snowflake-id/pkg/common"
	"snowflake-id/pkg/snowflake"
)

type Server struct {
	generator *snowflake.Generator
}

func NewServer(generator *snowflake.Generator) *Server {
	return &Server{generator: generator}
}

func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	var req common.GenerateRequest
	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "invalid request body", http.StatusBadRequest)
			return
		}
	} else if r.Method == http.MethodGet {
		countStr := r.URL.Query().Get("count")
		if countStr != "" {
			count, err := strconv.Atoi(countStr)
			if err != nil {
				writeError(w, "invalid count parameter", http.StatusBadRequest)
				return
			}
			req.Count = count
		}
	} else {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if req.Count <= 1 {
		id, err := s.generator.Next()
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		resp := common.GenerateResponse{
			Success: true,
			ID:      id,
			Count:   1,
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	ids, err := s.generator.NextBatch(req.Count)
	if err != nil {
		resp := common.GenerateResponse{
			Success: false,
			Error:   err.Error(),
			Count:   len(ids),
		}
		if len(ids) > 0 {
			resp.IDs = ids
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := common.GenerateResponse{
		Success: true,
		IDs:     ids,
		Count:   len(ids),
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := s.generator.Status()
	resp := common.StatusResponse{
		Success:          true,
		LastTimestamp:    status.LastTimestamp,
		CurrentTimestamp: status.CurrentTimestamp,
		MachineID:        status.MachineID,
		LastSequence:     status.LastSequence,
		Epoch:            status.Epoch,
		MaxSequence:      status.MaxSequence,
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleParse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	var req common.ParseRequest
	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "invalid request body", http.StatusBadRequest)
			return
		}
	} else if r.Method == http.MethodGet {
		idStr := r.URL.Query().Get("id")
		epochStr := r.URL.Query().Get("epoch")
		
		if idStr == "" {
			writeError(w, "id parameter is required", http.StatusBadRequest)
			return
		}
		
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			writeError(w, "invalid id parameter", http.StatusBadRequest)
			return
		}
		req.ID = id
		
		if epochStr != "" {
			epoch, err := strconv.ParseInt(epochStr, 10, 64)
			if err != nil {
				writeError(w, "invalid epoch parameter", http.StatusBadRequest)
				return
			}
			req.Epoch = epoch
		}
	} else {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	epoch := req.Epoch
	if epoch == 0 {
		epoch = snowflake.DefaultEpoch
	}

	parsed := snowflake.ParseID(req.ID, epoch)
	resp := common.ParseResponse{
		Success:       true,
		ID:            parsed.ID,
		Timestamp:     parsed.Timestamp,
		RealTimestamp: parsed.RealTimestamp,
		MachineID:     parsed.MachineID,
		Sequence:      parsed.Sequence,
	}
	json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(common.ErrorResponse{
		Success: false,
		Error:   message,
	})
}

func main() {
	var (
		port         int
		epoch        int64
		machineID    int64
		autoMachineID bool
		maxWaitMillis int64
	)

	flag.IntVar(&port, "port", 8080, "HTTP server port")
	flag.Int64Var(&epoch, "epoch", snowflake.DefaultEpoch, "Epoch timestamp in milliseconds")
	flag.Int64Var(&machineID, "machine-id", 0, "Machine ID (0-1023)")
	flag.BoolVar(&autoMachineID, "auto-machine-id", true, "Auto-generate machine ID from hostname and IP")
	flag.Int64Var(&maxWaitMillis, "max-wait", snowflake.DefaultMaxWait, "Max milliseconds to wait for clock to catch up")
	flag.Parse()

	config := &snowflake.Config{
		Epoch:         epoch,
		MachineID:     machineID,
		AutoMachineID: autoMachineID,
		MaxWaitMillis: maxWaitMillis,
	}

	generator, err := snowflake.NewGenerator(config)
	if err != nil {
		log.Fatalf("Failed to create generator: %v", err)
	}

	server := NewServer(generator)
	
	http.HandleFunc("/generate", server.handleGenerate)
	http.HandleFunc("/status", server.handleStatus)
	http.HandleFunc("/parse", server.handleParse)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Snowflake ID server starting on %s...", addr)
	log.Printf("  - Epoch: %d", config.Epoch)
	log.Printf("  - Machine ID: %d", generator.MachineID())
	log.Printf("  - Max wait for clock: %dms", config.MaxWaitMillis)
	
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
