package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"bloom/bloomfilter"
	"bloom/common"
)

type FilterManager struct {
	filters map[string]*bloomfilter.BloomFilter
	mu      sync.RWMutex
}

func NewFilterManager() *FilterManager {
	return &FilterManager{
		filters: make(map[string]*bloomfilter.BloomFilter),
	}
}

func (fm *FilterManager) Create(name string, expectedCapacity uint64, falsePositiveRate float64) (*bloomfilter.BloomFilter, error) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if _, exists := fm.filters[name]; exists {
		return nil, fmt.Errorf("filter with name '%s' already exists", name)
	}

	bf, err := bloomfilter.New(expectedCapacity, falsePositiveRate)
	if err != nil {
		return nil, err
	}

	fm.filters[name] = bf
	return bf, nil
}

func (fm *FilterManager) Get(name string) (*bloomfilter.BloomFilter, bool) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	bf, ok := fm.filters[name]
	return bf, ok
}

func (fm *FilterManager) Delete(name string) bool {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	if _, exists := fm.filters[name]; exists {
		delete(fm.filters, name)
		return true
	}
	return false
}

func (fm *FilterManager) List() []string {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	names := make([]string, 0, len(fm.filters))
	for name := range fm.filters {
		names = append(names, name)
	}
	return names
}

func (fm *FilterManager) Set(name string, bf *bloomfilter.BloomFilter) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.filters[name] = bf
}

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, statusCode int, err error) {
	writeJSON(w, statusCode, common.ErrorResponse{
		Success: false,
		Error:   err.Error(),
	})
}

func handleCreateFilter(fm *FilterManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req common.CreateFilterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
			return
		}

		if req.Name == "" {
			writeError(w, http.StatusBadRequest, fmt.Errorf("name is required"))
			return
		}

		bf, err := fm.Create(req.Name, req.ExpectedCapacity, req.FalsePositiveRate)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		writeJSON(w, http.StatusOK, common.CreateFilterResponse{
			Success:         true,
			Message:       "Filter created successfully",
			Name:          req.Name,
			BitArraySize: bf.M(),
			HashCount:    bf.K(),
			ExpectedCapacity: req.ExpectedCapacity,
		})
	}
}

func handleAddElement(fm *FilterManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req common.AddElementRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
			return
		}

		bf, ok := fm.Get(req.Name)
		if !ok {
			writeError(w, http.StatusNotFound, fmt.Errorf("filter '%s' not found", req.Name))
			return
		}

		if err := bf.Add(req.Element); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		writeJSON(w, http.StatusOK, common.AddElementResponse{
			Success: true,
			Message: "Element added successfully",
		})
	}
}

func handleContainsElement(fm *FilterManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req common.ContainsElementRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
			return
		}

		bf, ok := fm.Get(req.Name)
		if !ok {
			writeError(w, http.StatusNotFound, fmt.Errorf("filter '%s' not found", req.Name))
			return
		}

		contains, err := bf.Contains(req.Element)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		writeJSON(w, http.StatusOK, common.ContainsElementResponse{
			Success:  true,
			Contains: contains,
		})
	}
}

func handleGetFilterInfo(fm *FilterManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req common.GetFilterInfoRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
			return
		}

		bf, ok := fm.Get(req.Name)
		if !ok {
			writeError(w, http.StatusNotFound, fmt.Errorf("filter '%s' not found", req.Name))
			return
		}

		remaining, _ := bf.EstimateRemainingCapacity()

		writeJSON(w, http.StatusOK, common.GetFilterInfoResponse{
			Success:               true,
			Name:                  req.Name,
			ElementsAdded:         bf.ElementsAdded(),
			FillRate:             bf.FillRate(),
			CurrentFalsePositive: bf.CurrentFalsePositiveRate(),
			RemainingCapacity:    remaining,
			ExpectedCapacity:      bf.Capacity(),
			TargetFalsePositive:  bf.FalsePositiveRate(),
			BitArraySize:        bf.M(),
			HashCount:           bf.K(),
		})
	}
}

func handleClearFilter(fm *FilterManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req common.ClearFilterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
			return
		}

		bf, ok := fm.Get(req.Name)
		if !ok {
			writeError(w, http.StatusNotFound, fmt.Errorf("filter '%s' not found", req.Name))
			return
		}

		bf.Clear()

		writeJSON(w, http.StatusOK, common.ClearFilterResponse{
			Success: true,
			Message: "Filter cleared successfully",
		})
	}
}

func handleListFilters(fm *FilterManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filters := fm.List()
		writeJSON(w, http.StatusOK, common.ListFiltersResponse{
			Success: true,
			Filters: filters,
		})
	}
}

func handleExportFilter(fm *FilterManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req common.ExportFilterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
			return
		}

		bf, ok := fm.Get(req.Name)
		if !ok {
			writeError(w, http.StatusNotFound, fmt.Errorf("filter '%s' not found", req.Name))
			return
		}

		data, err := bf.SerializeToBase64()
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Errorf("failed to serialize filter: %w", err))
			return
		}

		writeJSON(w, http.StatusOK, common.ExportFilterResponse{
			Success: true,
			Name:    req.Name,
			Data:    data,
		})
	}
}

func handleImportFilter(fm *FilterManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req common.ImportFilterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
			return
		}

		bf, err := bloomfilter.DeserializeFromBase64(req.Data)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("failed to deserialize filter: %w", err))
			return
		}

		fm.Set(req.Name, bf)

		writeJSON(w, http.StatusOK, common.ImportFilterResponse{
			Success: true,
			Message: "Filter imported successfully",
		})
	}
}

func handleDeleteFilter(fm *FilterManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req common.DeleteFilterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
			return
		}

		if !fm.Delete(req.Name) {
			writeError(w, http.StatusNotFound, fmt.Errorf("filter '%s' not found", req.Name))
			return
		}

		writeJSON(w, http.StatusOK, common.DeleteFilterResponse{
			Success: true,
			Message: "Filter deleted successfully",
		})
	}
}

func getPort() string {
	port := "8080"

	flag.StringVar(&port, "port", port, "Server port")
	flag.Parse()

	if envPort := os.Getenv("BLOOM_SERVER_PORT"); envPort != "" {
		port = envPort
	}

	return port
}

func main() {
	port := getPort()
	fm := NewFilterManager()

	mux := http.NewServeMux()

	mux.HandleFunc("/create", handleCreateFilter(fm))
	mux.HandleFunc("/add", handleAddElement(fm))
	mux.HandleFunc("/contains", handleContainsElement(fm))
	mux.HandleFunc("/info", handleGetFilterInfo(fm))
	mux.HandleFunc("/clear", handleClearFilter(fm))
	mux.HandleFunc("/list", handleListFilters(fm))
	mux.HandleFunc("/export", handleExportFilter(fm))
	mux.HandleFunc("/import", handleImportFilter(fm))
	mux.HandleFunc("/delete", handleDeleteFilter(fm))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Bloom Filter server starting on port %s...", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
