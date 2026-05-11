package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/spatial-index/pkg/common"
	"github.com/spatial-index/pkg/geometry"
	"github.com/spatial-index/pkg/geojson"
	"github.com/spatial-index/pkg/rtree"
)

type Server struct {
	index *rtree.RTree
	mu    sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		index: rtree.NewRTree(),
	}
}

func (s *Server) handleInsert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		GeoJSON interface{} `json:"geojson"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.InsertResponse{
			Success: false,
			Error:   "Invalid JSON: " + err.Error(),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	data, err := json.Marshal(req.GeoJSON)
	if err != nil {
		resp := common.InsertResponse{
			Success: false,
			Error:   "Invalid GeoJSON: " + err.Error(),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	fc, err := geojson.Parse(data)
	if err != nil {
		resp := common.InsertResponse{
			Success: false,
			Error:   "Parse error: " + err.Error(),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	countBefore := s.index.Len()
	s.index.InsertFeatureCollection(fc)
	countAfter := s.index.Len()

	resp := common.InsertResponse{
		Success: true,
		Count:   countAfter - countBefore,
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleRangeQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req common.RangeQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.QueryResponse{
			Success: false,
			Error:   "Invalid JSON: " + err.Error(),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	rect := geometry.NewMBR(req.MinLon, req.MinLat, req.MaxLon, req.MaxLat)

	s.mu.RLock()
	candidates := s.index.Search(rect)
	s.mu.RUnlock()

	results := make([]geometry.Feature, 0, len(candidates))
	for _, f := range candidates {
		if f != nil {
			results = append(results, *f)
		}
	}

	fc := geometry.FeatureCollection{Features: results}
	geoJSON := geojson.FeatureCollectionToGeoJSON(fc)

	resp := common.QueryResponse{
		Success:  true,
		Count:    len(results),
		Features: geoJSON,
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handlePointQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req common.PointQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.QueryResponse{
			Success: false,
			Error:   "Invalid JSON: " + err.Error(),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	p := geometry.NewPoint(req.Lon, req.Lat)

	s.mu.RLock()
	features := s.index.PointQuery(p)
	s.mu.RUnlock()

	results := make([]geometry.Feature, 0, len(features))
	for _, f := range features {
		if f != nil {
			results = append(results, *f)
		}
	}

	fc := geometry.FeatureCollection{Features: results}
	geoJSON := geojson.FeatureCollectionToGeoJSON(fc)

	resp := common.QueryResponse{
		Success:  true,
		Count:    len(results),
		Features: geoJSON,
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleNearestNeighbor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req common.NearestNeighborRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.QueryResponse{
			Success: false,
			Error:   "Invalid JSON: " + err.Error(),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	if req.K <= 0 {
		req.K = 1
	}

	p := geometry.NewPoint(req.Lon, req.Lat)

	s.mu.RLock()
	features := s.index.NearestNeighbors(p, req.K)
	s.mu.RUnlock()

	results := make([]geometry.Feature, 0, len(features))
	for _, f := range features {
		if f != nil {
			results = append(results, *f)
		}
	}

	fc := geometry.FeatureCollection{Features: results}
	geoJSON := geojson.FeatureCollectionToGeoJSON(fc)

	resp := common.QueryResponse{
		Success:  true,
		Count:    len(results),
		Features: geoJSON,
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	s.mu.RLock()
	count := s.index.Len()
	s.mu.RUnlock()

	resp := common.StatsResponse{
		Success: true,
		Count:   count,
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	s.mu.Lock()
	s.index = rtree.NewRTree()
	s.mu.Unlock()

	resp := common.ClearResponse{Success: true}
	json.NewEncoder(w).Encode(resp)
}

func getPort() string {
	port := flag.String("port", "", "Port to listen on (e.g., 8080)")
	flag.Parse()

	if *port != "" {
		return ":" + *port
	}

	if envPort := os.Getenv("PORT"); envPort != "" {
		return ":" + envPort
	}

	return ":8504"
}

func main() {
	server := NewServer()

	http.HandleFunc("/insert", server.handleInsert)
	http.HandleFunc("/query/range", server.handleRangeQuery)
	http.HandleFunc("/query/point", server.handlePointQuery)
	http.HandleFunc("/query/nearest", server.handleNearestNeighbor)
	http.HandleFunc("/stats", server.handleStats)
	http.HandleFunc("/clear", server.handleClear)

	port := getPort()
	fmt.Printf("Spatial Index Server starting on %s\n", port)
	fmt.Println("Endpoints:")
	fmt.Println("  POST /insert         - Insert GeoJSON data")
	fmt.Println("  POST /query/range    - Range query (min_lon, min_lat, max_lon, max_lat)")
	fmt.Println("  POST /query/point    - Point query (lon, lat)")
	fmt.Println("  POST /query/nearest  - Nearest neighbor query (lon, lat, k)")
	fmt.Println("  GET  /stats          - Get index statistics")
	fmt.Println("  POST /clear          - Clear all data")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
