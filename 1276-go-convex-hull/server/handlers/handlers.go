package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"convex-hull/api"
	"convex-hull/convexhull"
)

type Server struct {
	sessions map[string]*convexhull.DynamicConvexHull
	mu       sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		sessions: make(map[string]*convexhull.DynamicConvexHull),
	}
}

func (s *Server) HandleCompute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req api.ComputeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}
	points := apiToCorePoints(req.Points)
	hull := convexhull.Compute(points)
	writeComputeResponse(w, hull)
}

func (s *Server) HandlePerimeterArea(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req api.PerimeterAreaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}
	points := apiToCorePoints(req.Points)
	hull := convexhull.Compute(points)
	perimeter := convexhull.Perimeter(hull)
	area := convexhull.Area(hull)
	writePerimeterAreaResponse(w, hull, perimeter, area)
}

func (s *Server) HandlePointInHull(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req api.PointInHullRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}
	p := convexhull.Point{X: req.Point.X, Y: req.Point.Y}
	hullPoints := apiToCorePoints(req.HullPoints)
	hull := convexhull.Compute(hullPoints)
	inside := convexhull.PointInConvexHull(p, hull)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.PointInHullResponse{
		Success: true,
		Inside:  inside,
	})
}

func (s *Server) HandleIntersection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req api.IntersectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}
	hull1 := convexhull.Compute(apiToCorePoints(req.Hull1))
	hull2 := convexhull.Compute(apiToCorePoints(req.Hull2))
	intersection := convexhull.Intersection(hull1, hull2)
	area := convexhull.Area(intersection)
	writeIntersectionResponse(w, intersection, area)
}

func (s *Server) HandleFilterBoundary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req api.FilterBoundaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}
	points := apiToCorePoints(req.Points)
	boundary := convexhull.FilterBoundaryPoints(points)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.FilterBoundaryResponse{
		Success:  true,
		Boundary: coreToAPIPoints(boundary),
	})
}

func (s *Server) HandleWeighted(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req api.WeightedComputeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}
	weighted := make([]convexhull.WeightedPoint, len(req.Points))
	for i, p := range req.Points {
		weighted[i] = convexhull.WeightedPoint{
			Point:  convexhull.Point{X: p.Point.X, Y: p.Point.Y},
			Weight: p.Weight,
		}
	}
	hull := convexhull.ComputeWeighted(weighted)
	writeComputeResponse(w, hull)
}

func (s *Server) HandleDynamicAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req api.DynamicAddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	var dh *convexhull.DynamicConvexHull
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = generateSessionID()
		dh = convexhull.NewDynamicConvexHull()
		s.sessions[sessionID] = dh
	} else {
		var ok bool
		dh, ok = s.sessions[sessionID]
		if !ok {
			dh = convexhull.NewDynamicConvexHull()
			s.sessions[sessionID] = dh
		}
	}
	points := apiToCorePoints(req.Points)
	dh.AddPoints(points)
	hull := dh.Hull()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.DynamicAddResponse{
		Success:   true,
		SessionID: sessionID,
		Hull:      coreToAPIHull(hull),
	})
}

func (s *Server) HandleDynamicGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req api.DynamicGetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	dh, ok := s.sessions[req.SessionID]
	if !ok {
		writeError(w, "session not found", http.StatusNotFound)
		return
	}
	hull := dh.Hull()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.DynamicGetResponse{
		Success: true,
		Hull:    coreToAPIHull(hull),
	})
}

func apiToCorePoints(pts []api.Point) []convexhull.Point {
	result := make([]convexhull.Point, len(pts))
	for i, p := range pts {
		result[i] = convexhull.Point{X: p.X, Y: p.Y}
	}
	return result
}

func coreToAPIPoints(pts []convexhull.Point) []api.Point {
	if pts == nil {
		return nil
	}
	result := make([]api.Point, len(pts))
	for i, p := range pts {
		result[i] = api.Point{X: p.X, Y: p.Y}
	}
	return result
}

func coreToAPIHull(hull *convexhull.ConvexHull) api.ConvexHullResponse {
	var t api.HullType
	switch hull.Type {
	case convexhull.HullEmpty:
		t = api.HullTypeEmpty
	case convexhull.HullPoint:
		t = api.HullTypePoint
	case convexhull.HullSegment:
		t = api.HullTypeSegment
	case convexhull.HullPolygon:
		t = api.HullTypePolygon
	}
	return api.ConvexHullResponse{
		Type:   t,
		Points: coreToAPIPoints(hull.Points),
	}
}

func writeError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   msg,
	})
}

func writeComputeResponse(w http.ResponseWriter, hull *convexhull.ConvexHull) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.ComputeResponse{
		Success: true,
		Hull:    coreToAPIHull(hull),
	})
}

func writePerimeterAreaResponse(w http.ResponseWriter, hull *convexhull.ConvexHull, perimeter, area float64) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.PerimeterAreaResponse{
		Success:   true,
		Perimeter: perimeter,
		Area:      area,
		Hull:      coreToAPIHull(hull),
	})
}

func writeIntersectionResponse(w http.ResponseWriter, intersection *convexhull.ConvexHull, area float64) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.IntersectionResponse{
		Success:          true,
		Intersection:     coreToAPIHull(intersection),
		IntersectionArea: area,
	})
}

var counter int

func generateSessionID() string {
	counter++
	return "session-" + string(rune(counter))
}
