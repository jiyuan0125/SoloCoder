package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"s2geometry/api"
	"s2geometry/s2"
)

func main() {
	port := flag.String("port", "", "HTTP server port")
	flag.Parse()

	if *port == "" {
		if envPort := os.Getenv("S2_PORT"); envPort != "" {
			*port = envPort
		} else {
			*port = "8500"
		}
	}

	http.HandleFunc("/latlng-to-cellid", latLngToCellIDHandler)
	http.HandleFunc("/cellid-to-latlng", cellIDToLatLngHandler)
	http.HandleFunc("/contains", containsHandler)
	http.HandleFunc("/neighbors", neighborsHandler)
	http.HandleFunc("/covering", coveringHandler)

	fmt.Printf("S2 Geometry server listening on port %s\n", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(api.ErrorResponse{Error: message})
}

func latLngToCellIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.LatLngToCellIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Level < 0 || req.Level > s2.MaxLevel {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Level must be between 0 and %d", s2.MaxLevel))
		return
	}

	ll := s2.LatLng{
		Lat: s2.DegToRad(req.Lat),
		Lng: s2.DegToRad(req.Lng),
	}
	cellID := s2.CellIDFromLatLng(ll).Parent(req.Level)

	resp := api.LatLngToCellIDResponse{
		CellID: strconv.FormatUint(uint64(cellID), 10),
		Face:   cellID.Face(),
		Level:  cellID.Level(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func cellIDToLatLngHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.CellIDToLatLngRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cellIDUint, err := strconv.ParseUint(req.CellID, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid CellID")
		return
	}

	cellID := s2.CellID(cellIDUint)
	if !cellID.IsValid() {
		writeError(w, http.StatusBadRequest, "Invalid CellID")
		return
	}

	ll := cellID.LatLng()
	rect := cellID.RectBound()

	resp := api.CellIDToLatLngResponse{
		Lat:   s2.RadToDeg(ll.Lat),
		Lng:   s2.RadToDeg(ll.Lng),
		LatLo: s2.RadToDeg(rect.LatLo),
		LatHi: s2.RadToDeg(rect.LatHi),
		LngLo: s2.RadToDeg(rect.LngLo),
		LngHi: s2.RadToDeg(rect.LngHi),
		Face:  cellID.Face(),
		Level: cellID.Level(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func containsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.ContainsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	containerUint, err := strconv.ParseUint(req.Container, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid container CellID")
		return
	}

	containedUint, err := strconv.ParseUint(req.Contained, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid contained CellID")
		return
	}

	container := s2.CellID(containerUint)
	contained := s2.CellID(containedUint)

	resp := api.ContainsResponse{
		Contains: contained.Contains(container),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func neighborsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.NeighborsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cellIDUint, err := strconv.ParseUint(req.CellID, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid CellID")
		return
	}

	if req.Level < 0 || req.Level > s2.MaxLevel {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Level must be between 0 and %d", s2.MaxLevel))
		return
	}

	cellID := s2.CellID(cellIDUint)
	neighbors := cellID.AllNeighbors(req.Level)

	neighborStrings := make([]string, 0, len(neighbors))
	seen := make(map[uint64]bool)
	for _, n := range neighbors {
		if !seen[uint64(n)] {
			neighborStrings = append(neighborStrings, strconv.FormatUint(uint64(n), 10))
			seen[uint64(n)] = true
		}
	}

	resp := api.NeighborsResponse{
		Neighbors: neighborStrings,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func coveringHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.CoveringRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.MinLevel < 0 || req.MinLevel > s2.MaxLevel {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("MinLevel must be between 0 and %d", s2.MaxLevel))
		return
	}
	if req.MaxLevel < 0 || req.MaxLevel > s2.MaxLevel {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("MaxLevel must be between 0 and %d", s2.MaxLevel))
		return
	}
	if req.MinLevel > req.MaxLevel {
		writeError(w, http.StatusBadRequest, "MinLevel must not exceed MaxLevel")
		return
	}
	if req.MaxCells <= 0 {
		req.MaxCells = 100
	}

	rect := s2.RectFromDegrees(req.LatLo, req.LatHi, req.LngLo, req.LngHi)
	coverer := s2.RegionCoverer{
		MinLevel: req.MinLevel,
		MaxLevel: req.MaxLevel,
		MaxCells: req.MaxCells,
	}
	cells := coverer.Covering(rect)

	cellStrings := make([]string, 0, len(cells))
	for _, cell := range cells {
		cellStrings = append(cellStrings, strconv.FormatUint(uint64(cell), 10))
	}

	resp := api.CoveringResponse{
		CellIDs: cellStrings,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
