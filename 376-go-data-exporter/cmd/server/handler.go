package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	"data-exporter/pkg/exporter"
	"data-exporter/pkg/proto"
)

func handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req proto.ExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	config := exporter.ExportConfig{
		Fields:       req.Fields,
		FieldMapping: req.FieldMapping,
	}

	var buf bytes.Buffer
	var err error

	switch req.Format {
	case proto.FormatCSV:
		err = exporter.ExportCSV(&buf, req.Data, config)
	case proto.FormatJSON:
		err = exporter.ExportJSON(&buf, req.Data, config)
	case proto.FormatJSONLines:
		err = exporter.ExportJSONLines(&buf, req.Data, config)
	default:
		http.Error(w, "Invalid export format", http.StatusBadRequest)
		return
	}

	if err != nil {
		resp := proto.ExportResponse{
			Success: false,
			Message: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := proto.ExportResponse{
		Success: true,
		Data:    buf.String(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
