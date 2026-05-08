package main

import (
	"encoding/json"
	"net/http"
	"sync"

	"go-bitmap-index/api"
	"go-bitmap-index/pkg/bitmap"
)

type Server struct {
	mu      sync.RWMutex
	indices map[string]*bitmap.Bitmap
}

func NewServer() *Server {
	return &Server{
		indices: make(map[string]*bitmap.Bitmap),
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, api.ErrorResponse{Error: err.Error()})
}

func (s *Server) CreateIndex(w http.ResponseWriter, r *http.Request) {
	var req api.CreateIndexRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, &customErr{"index name is required"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.indices[req.Name]; exists {
		writeError(w, http.StatusConflict, &customErr{"index already exists"})
		return
	}

	bm, err := bitmap.New(req.Size)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	s.indices[req.Name] = bm
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) DeleteIndex(w http.ResponseWriter, r *http.Request) {
	var req api.DeleteIndexRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.indices[req.Name]; !exists {
		writeError(w, http.StatusNotFound, &customErr{"index not found"})
		return
	}

	delete(s.indices, req.Name)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) ListIndices(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	indices := make([]api.IndexInfo, 0, len(s.indices))
	for name, bm := range s.indices {
		indices = append(indices, api.IndexInfo{
			Name:  name,
			Size:  bm.Size(),
			Count: bm.Count(),
		})
	}

	writeJSON(w, http.StatusOK, api.ListIndicesResponse{Indices: indices})
}

func (s *Server) SetBit(w http.ResponseWriter, r *http.Request) {
	var req api.SetBitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	s.mu.RLock()
	bm, exists := s.indices[req.Index]
	s.mu.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, &customErr{"index not found"})
		return
	}

	if err := bm.Set(req.Bit); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) ClearBit(w http.ResponseWriter, r *http.Request) {
	var req api.ClearBitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	s.mu.RLock()
	bm, exists := s.indices[req.Index]
	s.mu.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, &customErr{"index not found"})
		return
	}

	if err := bm.Clear(req.Bit); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) GetBit(w http.ResponseWriter, r *http.Request) {
	var req api.GetBitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	s.mu.RLock()
	bm, exists := s.indices[req.Index]
	s.mu.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, &customErr{"index not found"})
		return
	}

	value, err := bm.Get(req.Bit)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, api.GetBitResponse{Value: value})
}

func (s *Server) Count(w http.ResponseWriter, r *http.Request) {
	var req api.CountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	s.mu.RLock()
	bm, exists := s.indices[req.Index]
	s.mu.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, &customErr{"index not found"})
		return
	}

	writeJSON(w, http.StatusOK, api.CountResponse{Count: bm.Count()})
}

func (s *Server) Indices(w http.ResponseWriter, r *http.Request) {
	var req api.IndicesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	s.mu.RLock()
	bm, exists := s.indices[req.Index]
	s.mu.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, &customErr{"index not found"})
		return
	}

	writeJSON(w, http.StatusOK, api.IndicesResponse{Indices: bm.Indices()})
}

func (s *Server) getBitmaps(names []string) ([]*bitmap.Bitmap, error) {
	bitmaps := make([]*bitmap.Bitmap, len(names))
	for i, name := range names {
		bm, exists := s.indices[name]
		if !exists {
			return nil, &customErr{"index not found: " + name}
		}
		bitmaps[i] = bm
	}
	return bitmaps, nil
}

func (s *Server) Operation(w http.ResponseWriter, r *http.Request) {
	var req api.OperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	bitmaps, err := s.getBitmaps(req.Indices)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	var result *bitmap.Bitmap

	switch req.Operation {
	case api.OpAnd:
		if len(bitmaps) < 1 {
			writeError(w, http.StatusBadRequest, &customErr{"AND operation requires at least one index"})
			return
		}
		result, err = bitmap.And(bitmaps[0], bitmaps[1:]...)
	case api.OpOr:
		if len(bitmaps) < 1 {
			writeError(w, http.StatusBadRequest, &customErr{"OR operation requires at least one index"})
			return
		}
		result, err = bitmap.Or(bitmaps[0], bitmaps[1:]...)
	case api.OpXor:
		if len(bitmaps) < 2 {
			writeError(w, http.StatusBadRequest, &customErr{"XOR operation requires at least two indices"})
			return
		}
		result, err = bitmap.Xor(bitmaps[0], bitmaps[1:]...)
	case api.OpNot:
		if len(bitmaps) != 1 {
			writeError(w, http.StatusBadRequest, &customErr{"NOT operation requires exactly one index"})
			return
		}
		result = bitmap.Not(bitmaps[0])
	case api.OpAndNot:
		if len(bitmaps) != 2 {
			writeError(w, http.StatusBadRequest, &customErr{"ANDNOT operation requires exactly two indices"})
			return
		}
		result, err = bitmap.AndNot(bitmaps[0], bitmaps[1])
	default:
		writeError(w, http.StatusBadRequest, &customErr{"unknown operation: " + string(req.Operation)})
		return
	}

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.Result != "" {
		s.indices[req.Result] = result
	}

	writeJSON(w, http.StatusOK, api.OperationResponse{Count: result.Count()})
}

func (s *Server) Serialize(w http.ResponseWriter, r *http.Request) {
	var req api.SerializeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	s.mu.RLock()
	bm, exists := s.indices[req.Index]
	s.mu.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, &customErr{"index not found"})
		return
	}

	data, err := bm.Serialize()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, api.SerializeResponse{Data: data})
}

func (s *Server) Deserialize(w http.ResponseWriter, r *http.Request) {
	var req api.DeserializeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, &customErr{"index name is required"})
		return
	}

	bm, err := bitmap.Deserialize(req.Data)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	s.mu.Lock()
	s.indices[req.Name] = bm
	s.mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

type customErr struct {
	msg string
}

func (e *customErr) Error() string {
	return e.msg
}
