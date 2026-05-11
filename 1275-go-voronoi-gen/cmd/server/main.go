package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"

	"voronoi/internal/geom"
	"voronoi/internal/types"
	"voronoi/internal/voronoi"
)

type Server struct {
	diagram *voronoi.Diagram
	bounds  geom.Rectangle
	mu      sync.RWMutex
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) setDiagram(diagram *voronoi.Diagram, bounds geom.Rectangle) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.diagram = diagram
	s.bounds = bounds
}

func (s *Server) getDiagram() (*voronoi.Diagram, geom.Rectangle) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.diagram, s.bounds
}

func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req types.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	seeds := make([]geom.Point, len(req.Seeds))
	for i, seed := range req.Seeds {
		seeds[i] = geom.Point{X: seed.X, Y: seed.Y}
	}

	generator := voronoi.NewGenerator()
	diagram := generator.Generate(seeds)

	var bounds geom.Rectangle
	if req.Bounds.MinX == 0 && req.Bounds.MinY == 0 && req.Bounds.MaxX == 0 && req.Bounds.MaxY == 0 {
		bounds = s.calculateDefaultBounds(seeds)
	} else {
		bounds = geom.Rectangle{
			Min: geom.Point{X: req.Bounds.MinX, Y: req.Bounds.MinY},
			Max: geom.Point{X: req.Bounds.MaxX, Y: req.Bounds.MaxY},
		}
	}

	clipped := voronoi.ClipToBounds(diagram, bounds)
	s.setDiagram(clipped, bounds)

	cells := voronoi.BuildCells(clipped, bounds)
	areas := voronoi.CalculateCellAreas(clipped, bounds)

	response := types.GenerateResponse{
		Diagram: s.convertToTypesDiagram(clipped, cells, areas, bounds),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleNearestNeighbor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	diagram, _ := s.getDiagram()
	if diagram == nil {
		http.Error(w, "No diagram generated yet", http.StatusBadRequest)
		return
	}

	var req types.NearestNeighborRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := geom.Point{X: req.Point.X, Y: req.Point.Y}
	seed, distance, index := voronoi.FindNearestNeighbor(diagram, query)

	response := types.NearestNeighborResponse{
		Seed:      types.Point{X: seed.X, Y: seed.Y},
		Distance:  distance,
		CellIndex: index,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleClip(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	diagram, _ := s.getDiagram()
	if diagram == nil {
		http.Error(w, "No diagram generated yet", http.StatusBadRequest)
		return
	}

	var req types.ClipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	region := geom.Rectangle{
		Min: geom.Point{X: req.Region.MinX, Y: req.Region.MinY},
		Max: geom.Point{X: req.Region.MaxX, Y: req.Region.MaxY},
	}

	clippedEdges := voronoi.ClipEdgesToRegion(diagram, region)

	response := types.ClipResponse{
		Edges: make([]types.Edge, len(clippedEdges)),
	}

	for i, edge := range clippedEdges {
		response.Edges[i] = types.Edge{
			Start: types.Point{X: edge.Start.X, Y: edge.Start.Y},
			End:   types.Point{X: edge.End.X, Y: edge.End.Y},
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleArea(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	diagram, bounds := s.getDiagram()
	if diagram == nil {
		http.Error(w, "No diagram generated yet", http.StatusBadRequest)
		return
	}

	areas := voronoi.CalculateCellAreas(diagram, bounds)

	response := types.AreaResponse{
		Areas: areas,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) calculateDefaultBounds(seeds []geom.Point) geom.Rectangle {
	if len(seeds) == 0 {
		return geom.Rectangle{
			Min: geom.Point{X: -10, Y: -10},
			Max: geom.Point{X: 10, Y: 10},
		}
	}

	minX, maxX := seeds[0].X, seeds[0].X
	minY, maxY := seeds[0].Y, seeds[0].Y

	for _, seed := range seeds {
		if seed.X < minX {
			minX = seed.X
		}
		if seed.X > maxX {
			maxX = seed.X
		}
		if seed.Y < minY {
			minY = seed.Y
		}
		if seed.Y > maxY {
			maxY = seed.Y
		}
	}

	margin := (maxX - minX + maxY - minY) * 0.5
	if margin < 1 {
		margin = 1
	}

	return geom.Rectangle{
		Min: geom.Point{X: minX - margin, Y: minY - margin},
		Max: geom.Point{X: maxX + margin, Y: maxY + margin},
	}
}

func (s *Server) convertToTypesDiagram(diagram *voronoi.Diagram, cells []*voronoi.VoronoiCell, areas []float64, bounds geom.Rectangle) types.Diagram {
	result := types.Diagram{
		Seeds: make([]types.Point, len(diagram.Seeds)),
		Edges: make([]types.Edge, len(diagram.Edges)),
		Bounds: types.Bounds{
			MinX: bounds.Min.X,
			MinY: bounds.Min.Y,
			MaxX: bounds.Max.X,
			MaxY: bounds.Max.Y,
		},
	}

	for i, seed := range diagram.Seeds {
		result.Seeds[i] = types.Point{X: seed.X, Y: seed.Y}
	}

	for i, edge := range diagram.Edges {
		result.Edges[i] = types.Edge{
			Start: types.Point{X: edge.Start.X, Y: edge.Start.Y},
			End:   types.Point{X: edge.End.X, Y: edge.End.Y},
		}
	}

	if cells != nil {
		result.Cells = make([]types.VoronoiCell, len(cells))
		for i, cell := range cells {
			result.Cells[i].Seed = types.Point{X: cell.Seed.X, Y: cell.Seed.Y}
			result.Cells[i].Points = make([]types.Point, len(cell.Points))
			for j, p := range cell.Points {
				result.Cells[i].Points[j] = types.Point{X: p.X, Y: p.Y}
			}
			result.Cells[i].Edges = make([]types.Edge, len(cell.Edges))
			for j, e := range cell.Edges {
				result.Cells[i].Edges[j] = types.Edge{
					Start: types.Point{X: e.Start.X, Y: e.Start.Y},
					End:   types.Point{X: e.End.X, Y: e.End.Y},
				}
			}
			if i < len(areas) {
				result.Cells[i].Area = areas[i]
			}
		}
	}

	return result
}

func main() {
	port := flag.String("port", "8503", "Server port")
	flag.Parse()

	if envPort := os.Getenv("VORONOI_PORT"); envPort != "" {
		*port = envPort
	}

	server := NewServer()

	http.HandleFunc("/generate", server.handleGenerate)
	http.HandleFunc("/nearest", server.handleNearestNeighbor)
	http.HandleFunc("/clip", server.handleClip)
	http.HandleFunc("/area", server.handleArea)

	fmt.Printf("Voronoi server starting on port %s...\n", *port)
	fmt.Printf("Endpoints:\n")
	fmt.Printf("  POST /generate - Generate Voronoi diagram\n")
	fmt.Printf("  POST /nearest  - Find nearest neighbor\n")
	fmt.Printf("  POST /clip     - Clip edges to region\n")
	fmt.Printf("  POST /area     - Calculate cell areas\n")

	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}
