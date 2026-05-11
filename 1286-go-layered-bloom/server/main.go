package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"layered-bloom/common"
	"layered-bloom/core"
)

var filter *core.LayeredBloomFilter

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, common.ErrorResponse{
		Success: false,
		Message: message,
	})
}

func insertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req common.InsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Element == "" {
		writeError(w, http.StatusBadRequest, "element is required")
		return
	}
	result, err := filter.Insert([]byte(req.Element))
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, common.InsertResponse{
		Success:     true,
		Inserted:    result.Inserted,
		LayerIndex:  result.LayerIndex,
		WasExisting: result.WasExisting,
	})
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	element := r.URL.Query().Get("element")
	if element == "" {
		writeError(w, http.StatusBadRequest, "element query parameter is required")
		return
	}
	exists := filter.Contains([]byte(element))
	writeJSON(w, http.StatusOK, common.QueryResponse{
		Exists: exists,
	})
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req common.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Element == "" {
		writeError(w, http.StatusBadRequest, "element is required")
		return
	}
	removed := filter.Remove([]byte(req.Element))
	writeJSON(w, http.StatusOK, common.DeleteResponse{
		Success: true,
		Removed: removed,
	})
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	stats := filter.Stats()
	writeJSON(w, http.StatusOK, common.StatsResponse{
		Success: true,
		Stats:   stats,
	})
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	filter.Reset()
	writeJSON(w, http.StatusOK, common.ResetResponse{
		Success: true,
		Message: "filter reset successfully",
	})
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req common.ConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	layers := make([]core.LayerConfig, len(req.Layers))
	for i, l := range req.Layers {
		layers[i] = core.LayerConfig{
			Capacity:      l.Capacity,
			HashFunctions: l.HashFunctions,
			TargetFPR:     l.TargetFPR,
		}
	}
	cfg := &core.FilterConfig{
		Layers:           layers,
		AutoExpandOnFull: req.AutoExpandOnFull,
		WarningFPRFactor: req.WarningFPRFactor,
		MaxAutoLayers:    req.MaxAutoLayers,
	}
	if err := filter.Reconfigure(cfg); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, common.ConfigResponse{
		Success: true,
		Message: "configuration updated successfully",
	})
}

func batchInsertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req common.BatchInsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	results := make([]common.BatchInsertResult, 0, len(req.Elements))
	inserted := 0
	existing := 0
	failed := 0
	for _, elem := range req.Elements {
		res, err := filter.Insert([]byte(elem))
		if err != nil {
			results = append(results, common.BatchInsertResult{
				Element: elem,
				Success: false,
			})
			failed++
			continue
		}
		if res.Inserted {
			inserted++
		} else if res.WasExisting {
			existing++
		}
		results = append(results, common.BatchInsertResult{
			Element:     elem,
			Success:     true,
			Inserted:    res.Inserted,
			LayerIndex:  res.LayerIndex,
			WasExisting: res.WasExisting,
		})
	}
	writeJSON(w, http.StatusOK, common.BatchInsertResponse{
		Success:  true,
		Results:  results,
		Total:    len(req.Elements),
		Inserted: inserted,
		Existing: existing,
		Failed:   failed,
	})
}

func batchQueryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req common.BatchQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	results := make([]common.BatchQueryResult, 0, len(req.Elements))
	exists := 0
	notExists := 0
	for _, elem := range req.Elements {
		found := filter.Contains([]byte(elem))
		results = append(results, common.BatchQueryResult{
			Element: elem,
			Exists:  found,
		})
		if found {
			exists++
		} else {
			notExists++
		}
	}
	writeJSON(w, http.StatusOK, common.BatchQueryResponse{
		Success:   true,
		Results:   results,
		Total:     len(req.Elements),
		Exists:    exists,
		NotExists: notExists,
	})
}

func batchDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req common.BatchDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	results := make([]common.BatchDeleteResult, 0, len(req.Elements))
	removed := 0
	notFound := 0
	for _, elem := range req.Elements {
		ok := filter.Remove([]byte(elem))
		results = append(results, common.BatchDeleteResult{
			Element: elem,
			Success: true,
			Removed: ok,
		})
		if ok {
			removed++
		} else {
			notFound++
		}
	}
	writeJSON(w, http.StatusOK, common.BatchDeleteResponse{
		Success:  true,
		Results:  results,
		Total:    len(req.Elements),
		Removed:  removed,
		NotFound: notFound,
	})
}

func getPort() string {
	port := flag.String("port", "", "server port")
	flag.Parse()
	if *port != "" {
		return *port
	}
	if envPort := os.Getenv("BLOOM_PORT"); envPort != "" {
		if _, err := strconv.Atoi(envPort); err == nil {
			return envPort
		}
	}
	return "8513"
}

func main() {
	var err error
	filter, err = core.NewLayeredBloomFilter(core.DefaultConfig())
	if err != nil {
		log.Fatalf("failed to create filter: %v", err)
	}
	http.HandleFunc("/insert", insertHandler)
	http.HandleFunc("/query", queryHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/stats", statsHandler)
	http.HandleFunc("/reset", resetHandler)
	http.HandleFunc("/config", configHandler)
	http.HandleFunc("/batch/insert", batchInsertHandler)
	http.HandleFunc("/batch/query", batchQueryHandler)
	http.HandleFunc("/batch/delete", batchDeleteHandler)
	port := getPort()
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Layered Bloom Filter server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
