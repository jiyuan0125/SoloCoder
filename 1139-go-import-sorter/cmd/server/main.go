package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"sync"

	"gofmt-tool/api"
	"gofmt-tool/formatter"
)

var (
	defaultConfig = api.DefaultConfig()
	configMutex   sync.RWMutex
	currentConfig = defaultConfig
)

func main() {
	port := flag.String("port", "8080", "HTTP server port")
	flag.Parse()

	http.HandleFunc("/format", handleFormat)
	http.HandleFunc("/config", handleConfig)
	http.HandleFunc("/batch", handleBatch)

	fmt.Printf("Server starting on port %s...\n", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func handleFormat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.FormatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cfg := getEffectiveConfig(req.Config)
	source, stats, err := formatter.Format(req.Source, toFormatterConfig(cfg))
	if err != nil {
		http.Error(w, fmt.Sprintf("Format error: %v", err), http.StatusBadRequest)
		return
	}

	resp := api.FormatResponse{
		Source: source,
		Stats: api.Stats{
			LinesModified: stats.LinesModified,
			ImportsMoved:  stats.ImportsMoved,
			LinesSplit:    stats.LinesSplit,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		configMutex.RLock()
		cfg := currentConfig
		configMutex.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cfg)
		return
	}

	if r.Method == http.MethodPost {
		var req api.ConfigRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		configMutex.Lock()
		currentConfig = req.Config
		if currentConfig.LineWidth <= 0 {
			currentConfig.LineWidth = 120
		}
		configMutex.Unlock()

		resp := api.ConfigResponse{Success: true}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.BatchFormatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cfg := getEffectiveConfig(req.Config)
	fmtCfg := toFormatterConfig(cfg)

	results := make([]api.FileResult, 0, len(req.Files))
	for _, file := range req.Files {
		source, stats, err := formatter.Format(file.Source, fmtCfg)
		if err != nil {
			continue
		}
		results = append(results, api.FileResult{
			Name:   file.Name,
			Source: source,
			Stats: api.Stats{
				LinesModified: stats.LinesModified,
				ImportsMoved:  stats.ImportsMoved,
				LinesSplit:    stats.LinesSplit,
			},
		})
	}

	resp := api.BatchFormatResponse{Files: results}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getEffectiveConfig(reqCfg api.Config) api.Config {
	configMutex.RLock()
	current := currentConfig
	configMutex.RUnlock()

	if reqCfg.LineWidth <= 0 {
		reqCfg.LineWidth = current.LineWidth
	}
	if reqCfg.ModuleName == "" {
		reqCfg.ModuleName = current.ModuleName
	}

	return reqCfg
}

func toFormatterConfig(cfg api.Config) formatter.Config {
	return formatter.Config{
		LineWidth:        cfg.LineWidth,
		RemoveEmptyLines: cfg.RemoveEmptyLines,
		ModuleName:       cfg.ModuleName,
	}
}
