package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"

	"geofence/common"
	"geofence/geofence"

	"github.com/google/uuid"
)

type Server struct {
	fences map[string]*geofence.Fence
	mu     sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		fences: make(map[string]*geofence.Fence),
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, common.ErrorResponse{Error: err.Error()})
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req common.FenceCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if len(req.Points) < 3 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("polygon must have at least 3 vertices"))
		return
	}

	points := make([]geofence.Point, len(req.Points))
	for i, p := range req.Points {
		points[i] = geofence.NewPoint(p.Lat, p.Lng)
	}

	polygon := geofence.Polygon{Points: points}
	if err := polygon.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	area := polygon.Area()

	fence := &geofence.Fence{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Polygon:     polygon,
		Area:        area,
	}

	s.mu.Lock()
	s.fences[fence.ID] = fence
	s.mu.Unlock()

	resp := common.FenceCreateResponse{
		ID:   fence.ID,
		Name: fence.Name,
		Area: area,
	}
	if area < geofence.MinArea {
		resp.Warning = fmt.Sprintf("fence area is %.2f m², below minimum recommended %d m²", area, int(geofence.MinArea))
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fences := make([]common.FenceInfo, 0, len(s.fences))
	for _, f := range s.fences {
		fences = append(fences, common.FenceInfo{
			ID:          f.ID,
			Name:        f.Name,
			Description: f.Description,
			Area:        f.Area,
			PointCount:  len(f.Polygon.Points),
		})
	}

	writeJSON(w, http.StatusOK, common.FenceListResponse{Fences: fences})
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/fences/"):]
	if id == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("fence id required"))
		return
	}

	s.mu.Lock()
	if _, exists := s.fences[id]; !exists {
		s.mu.Unlock()
		writeError(w, http.StatusNotFound, fmt.Errorf("fence not found"))
		return
	}
	delete(s.fences, id)
	s.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleJudge(w http.ResponseWriter, r *http.Request) {
	var req common.JudgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	point := geofence.NewPoint(req.Point.Lat, req.Point.Lng)

	s.mu.RLock()
	defer s.mu.RUnlock()

	var matchedIDs []string
	var matchedNames []string
	minDist := math.Inf(1)

	checkFence := func(f *geofence.Fence) {
		dist := f.Polygon.MinDistance(point)
		if dist < minDist {
			minDist = dist
		}
		if f.Polygon.Contains(point) {
			matchedIDs = append(matchedIDs, f.ID)
			matchedNames = append(matchedNames, f.Name)
		}
	}

	if req.FenceID != "" {
		f, exists := s.fences[req.FenceID]
		if !exists {
			writeError(w, http.StatusNotFound, fmt.Errorf("fence not found"))
			return
		}
		checkFence(f)
	} else {
		for _, f := range s.fences {
			checkFence(f)
		}
	}

	resp := common.JudgeResponse{
		Point: req.Point,
		Results: []common.JudgeResult{{
			In:         len(matchedIDs) > 0,
			FenceIDs:   matchedIDs,
			FenceNames: matchedNames,
		}},
		MinDistance: minDist,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleBatchJudge(w http.ResponseWriter, r *http.Request) {
	var req common.BatchJudgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]common.PointResult, len(req.Points))

	for idx, p := range req.Points {
		point := geofence.NewPoint(p.Lat, p.Lng)

		var matchedIDs []string
		var matchedNames []string
		minDist := math.Inf(1)

		for _, f := range s.fences {
			dist := f.Polygon.MinDistance(point)
			if dist < minDist {
				minDist = dist
			}
			if f.Polygon.Contains(point) {
				matchedIDs = append(matchedIDs, f.ID)
				matchedNames = append(matchedNames, f.Name)
			}
		}

		results[idx] = common.PointResult{
			Point:       p,
			In:          len(matchedIDs) > 0,
			FenceIDs:    matchedIDs,
			FenceNames:  matchedNames,
			MinDistance: minDist,
		}
	}

	writeJSON(w, http.StatusOK, common.BatchJudgeResponse{Results: results})
}

func main() {
	s := NewServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/fences", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			s.handleCreate(w, r)
		case http.MethodGet:
			s.handleList(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		}
	})
	mux.HandleFunc("/fences/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			s.handleDelete(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		}
	})
	mux.HandleFunc("/judge", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.handleJudge(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		}
	})
	mux.HandleFunc("/batch-judge", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.handleBatchJudge(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		}
	})

	fmt.Println("Geofence server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
