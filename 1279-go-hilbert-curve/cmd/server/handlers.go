package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/example/hilbert-curve-service/pkg/api"
	"github.com/example/hilbert-curve-service/pkg/hilbert"
)

func pointToIndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.PointToIndexRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	index, err := hilbert.PointToIndex(req.Order, hilbert.Point{X: req.X, Y: req.Y})
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := api.PointToIndexResponse{Index: index}
	sendJSON(w, resp, http.StatusOK)
}

func indexToPointHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.IndexToPointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	point, err := hilbert.IndexToPoint(req.Order, req.Index)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := api.IndexToPointResponse{X: point.X, Y: point.Y}
	sendJSON(w, resp, http.StatusOK)
}

func rangeQueryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.RangeQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ranges, err := hilbert.RangeQuery(req.Order, req.MinX, req.MaxX, req.MinY, req.MaxY)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	apiRanges := make([]api.Range, len(ranges))
	for i, r := range ranges {
		apiRanges[i] = api.Range{Min: r.Min, Max: r.Max}
	}

	resp := api.RangeQueryResponse{Ranges: apiRanges}
	sendJSON(w, resp, http.StatusOK)
}

func estimateDistanceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.EstimateDistanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	distance, err := hilbert.EstimateDistance(req.Order, req.Index1, req.Index2)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := api.EstimateDistanceResponse{Distance: distance}
	sendJSON(w, resp, http.StatusOK)
}

func batchPointToIndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.BatchPointToIndexRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	hilbertPoints := make([]hilbert.Point, len(req.Points))
	for i, p := range req.Points {
		hilbertPoints[i] = hilbert.Point{X: p.X, Y: p.Y}
	}

	indices, err := hilbert.BatchPointToIndex(req.Order, hilbertPoints, req.Ascending)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := api.BatchPointToIndexResponse{Indices: indices}
	sendJSON(w, resp, http.StatusOK)
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(api.ErrorResponse{Error: message})
}

func sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func getPort() string {
	port := os.Getenv("HILBERT_PORT")
	if port == "" {
		return "8506"
	}
	if _, err := strconv.Atoi(port); err != nil {
		fmt.Printf("Invalid port number in environment: %s. Using default 8080.\n", port)
		return "8506"
	}
	return port
}
