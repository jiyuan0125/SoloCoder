package main

import (
	"encoding/json"
	"net/http"

	"github.com/example/externalsort/pkg/api"
)

func (s *server) handleAddData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.AddDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	s.sorter.AddData(req.Data)

	resp := api.AddDataResponse{
		Success:   true,
		TotalRows: s.sorter.GetStats().TotalRecords,
	}
	sendJSON(w, http.StatusOK, resp)
}

func (s *server) handleSetMemoryLimit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.SetMemoryLimitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	s.sorter.SetMemoryLimit(req.Limit)

	resp := api.SetMemoryLimitResponse{
		Success: true,
		Limit:   s.sorter.GetMemoryLimit(),
	}
	sendJSON(w, http.StatusOK, resp)
}

func (s *server) handleSetMergeWays(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.SetMergeWaysRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	s.sorter.SetMergeWays(req.Ways)

	resp := api.SetMergeWaysResponse{
		Success: true,
		Ways:    s.sorter.GetMergeWays(),
	}
	sendJSON(w, http.StatusOK, resp)
}

func (s *server) handleSort(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	go func() {
		_ = s.sorter.Sort()
	}()

	resp := api.SortResponse{
		Success: true,
		Message: "sorting started",
	}
	sendJSON(w, http.StatusOK, resp)
}

func (s *server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	stats := s.sorter.GetStats()

	resp := api.StatsResponse{
		Success:       true,
		TotalRecords:  stats.TotalRecords,
		ChunksCreated: stats.ChunksCreated,
		MergePasses:   stats.MergePasses,
		DiskReads:     stats.DiskReads,
		DiskWrites:    stats.DiskWrites,
		ChunkSizes:    stats.ChunkSizes,
		MemoryLimit:   s.sorter.GetMemoryLimit(),
		MergeWays:     s.sorter.GetMergeWays(),
		IsSorted:      s.sorter.IsSorted(),
		MergeDetails:  s.sorter.GetMergePassDetails(),
	}
	sendJSON(w, http.StatusOK, resp)
}

func (s *server) handleResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if !s.sorter.IsSorted() {
		resp := api.GetResultResponse{
			Success: false,
			Message: "sorting not completed",
		}
		sendJSON(w, http.StatusOK, resp)
		return
	}

	data := s.sorter.GetFinalResult()
	resp := api.GetResultResponse{
		Success: true,
		Data:    data,
	}
	sendJSON(w, http.StatusOK, resp)
}

func (s *server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	resp := api.GetConfigResponse{
		Success:     true,
		MemoryLimit: s.sorter.GetMemoryLimit(),
		MergeWays:   s.sorter.GetMergeWays(),
	}
	sendJSON(w, http.StatusOK, resp)
}

func (s *server) handleReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	s.sorter.Reset()

	resp := api.ResetResponse{
		Success: true,
	}
	sendJSON(w, http.StatusOK, resp)
}

func sendJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func sendError(w http.ResponseWriter, status int, message string) {
	resp := api.ErrorResponse{
		Success: false,
		Error:   message,
	}
	sendJSON(w, status, resp)
}
