package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"

	"rtree/internal/api"
	"rtree/internal/rtree"
)

const (
	Port = 8201
)

type Server struct {
	tree *rtree.Rtree
	mu   sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		tree: rtree.New(),
	}
}

func (s *Server) writeJSON(w http.ResponseWriter, statusCode int, resp *api.Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) writeError(w http.ResponseWriter, statusCode int, err error) {
	s.writeJSON(w, statusCode, &api.Response{
		Success: false,
		Error:   err.Error(),
	})
}

func (s *Server) handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	var req api.AddRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	added := 0
	for _, obj := range req.Objects {
		rObj := rtree.NewSimpleObject(obj.ID, obj.MinX, obj.MinY, obj.MaxX, obj.MaxY, obj.Metadata)
		if err := s.tree.Insert(rObj); err != nil {
			continue
		}
		added++
	}

	s.writeJSON(w, http.StatusOK, &api.Response{
		Success: true,
		Data: map[string]int{
			"added": added,
			"total": len(req.Objects),
		},
	})
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	var req api.DeleteRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.tree.Delete(req.ID); err != nil {
		s.writeError(w, http.StatusNotFound, err)
		return
	}

	s.writeJSON(w, http.StatusOK, &api.Response{
		Success: true,
	})
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	parseFloat := func(key string) (float64, error) {
		val := r.URL.Query().Get(key)
		if val == "" {
			return 0, fmt.Errorf("missing parameter: %s", key)
		}
		return strconv.ParseFloat(val, 64)
	}

	minX, err := parseFloat("x1")
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}
	minY, err := parseFloat("y1")
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}
	maxX, err := parseFloat("x2")
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}
	maxY, err := parseFloat("y2")
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	results := s.tree.Search(minX, minY, maxX, maxY)

	apiObjs := make([]api.Object, 0, len(results))
	for _, obj := range results {
		mx, my, mX, mY := obj.Bounds()
		apiObjs = append(apiObjs, api.Object{
			ID:       obj.ID(),
			MinX:     mx,
			MinY:     my,
			MaxX:     mX,
			MaxY:     mY,
			Metadata: obj.Metadata(),
		})
	}

	s.writeJSON(w, http.StatusOK, &api.Response{
		Success: true,
		Data:    apiObjs,
	})
}

func (s *Server) handleKNN(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	parseFloat := func(key string) (float64, error) {
		val := r.URL.Query().Get(key)
		if val == "" {
			return 0, fmt.Errorf("missing parameter: %s", key)
		}
		return strconv.ParseFloat(val, 64)
	}

	parseInt := func(key string) (int, error) {
		val := r.URL.Query().Get(key)
		if val == "" {
			return 0, fmt.Errorf("missing parameter: %s", key)
		}
		return strconv.Atoi(val)
	}

	x, err := parseFloat("x")
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}
	y, err := parseFloat("y")
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}
	k, err := parseInt("k")
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	results := s.tree.KNN(x, y, k)

	apiObjs := make([]api.Object, 0, len(results))
	for _, obj := range results {
		mx, my, mX, mY := obj.Bounds()
		apiObjs = append(apiObjs, api.Object{
			ID:       obj.ID(),
			MinX:     mx,
			MinY:     my,
			MaxX:     mX,
			MaxY:     mY,
			Metadata: obj.Metadata(),
		})
	}

	s.writeJSON(w, http.StatusOK, &api.Response{
		Success: true,
		Data:    apiObjs,
	})
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	allNodesMBR := s.tree.AllNodesMBR()
	mbrInfos := make([]api.MBRInfo, 0, len(allNodesMBR))
	for _, mbr := range allNodesMBR {
		mbrInfos = append(mbrInfos, api.MBRInfo{
			MinX: mbr.MinX,
			MinY: mbr.MinY,
			MaxX: mbr.MaxX,
			MaxY: mbr.MaxY,
		})
	}

	info := api.TreeInfo{
		NodeCount:   s.tree.NodeCount(),
		Height:      s.tree.Height(),
		ObjectCount: s.tree.Count(),
		AllNodesMBR: mbrInfos,
	}

	s.writeJSON(w, http.StatusOK, &api.Response{
		Success: true,
		Data:    info,
	})
}

func (s *Server) handleVisualize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	visualization := s.tree.Visualize()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(visualization))
}

func main() {
	server := NewServer()

	http.HandleFunc("/api/objects/add", server.handleAdd)
	http.HandleFunc("/api/objects/delete", server.handleDelete)
	http.HandleFunc("/api/objects/search", server.handleSearch)
	http.HandleFunc("/api/objects/knn", server.handleKNN)
	http.HandleFunc("/api/tree/info", server.handleInfo)
	http.HandleFunc("/api/tree/visualize", server.handleVisualize)

	addr := fmt.Sprintf(":%d", Port)
	fmt.Printf("R-tree server starting on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
