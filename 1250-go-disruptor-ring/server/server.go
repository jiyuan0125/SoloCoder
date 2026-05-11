package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"disruptor/common"
	"disruptor/ringbuffer"
)

type Server struct {
	manager *ringbuffer.Manager
	mux     *http.ServeMux
}

func NewServer() *Server {
	s := &Server{
		manager: ringbuffer.NewManager(),
		mux:     http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/buffer/create", s.handleCreateBuffer)
	s.mux.HandleFunc("/buffer/destroy", s.handleDestroyBuffer)
	s.mux.HandleFunc("/buffer/publish", s.handlePublish)
	s.mux.HandleFunc("/buffer/batch-publish", s.handleBatchPublish)
	s.mux.HandleFunc("/buffer/consume", s.handleConsume)
	s.mux.HandleFunc("/buffer/batch-consume", s.handleBatchConsume)
	s.mux.HandleFunc("/buffer/info", s.handleBufferInfo)
	s.mux.HandleFunc("/buffer/list", s.handleListBuffers)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, err string) {
	writeJSON(w, status, common.ErrorResponse{Error: err})
}

func (s *Server) handleCreateBuffer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.CreateBufferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	capacity, err := s.manager.CreateBuffer(req.Name, req.Size)
	if err != nil {
		switch err.Error() {
		case common.ErrBufferExists:
			writeError(w, http.StatusConflict, err.Error())
		case common.ErrInvalidSize:
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, common.CreateBufferResponse{
		Name:     req.Name,
		Capacity: capacity,
	})
}

func (s *Server) handleDestroyBuffer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.DestroyBufferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.manager.DestroyBuffer(req.Name); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, nil)
}

func (s *Server) handlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	seq, err := s.manager.Publish(req.BufferName, req.Message)
	if err != nil {
		switch err.Error() {
		case common.ErrBufferNotFound:
			writeError(w, http.StatusNotFound, err.Error())
		case common.ErrBufferFull:
			writeError(w, http.StatusTooManyRequests, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, common.PublishResponse{Sequence: seq})
}

func (s *Server) handleBatchPublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.BatchPublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	startSeq, count, err := s.manager.BatchPublish(req.BufferName, req.Messages)
	if err != nil {
		switch err.Error() {
		case common.ErrBufferNotFound:
			writeError(w, http.StatusNotFound, err.Error())
		case common.ErrBufferFull:
			writeError(w, http.StatusTooManyRequests, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, common.BatchPublishResponse{
		StartSequence: startSeq,
		Count:         count,
	})
}

func (s *Server) handleConsume(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.ConsumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	seq, message, err := s.manager.Consume(req.BufferName, req.ConsumerID)
	if err != nil {
		switch err.Error() {
		case common.ErrBufferNotFound:
			writeError(w, http.StatusNotFound, err.Error())
		case common.ErrNoData:
			writeError(w, http.StatusNoContent, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, common.ConsumeResponse{
		Sequence: seq,
		Message:  message,
	})
}

func (s *Server) handleBatchConsume(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.BatchConsumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	messages, err := s.manager.BatchConsume(req.BufferName, req.ConsumerID, req.MaxCount)
	if err != nil {
		switch err.Error() {
		case common.ErrBufferNotFound:
			writeError(w, http.StatusNotFound, err.Error())
		case common.ErrNoData:
			writeError(w, http.StatusNoContent, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, common.BatchConsumeResponse{Messages: messages})
}

func (s *Server) handleBufferInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	name := r.URL.Query().Get("name")
	if strings.TrimSpace(name) == "" {
		writeError(w, http.StatusBadRequest, "name parameter is required")
		return
	}

	info, err := s.manager.BufferInfo(name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, common.BufferInfoResponse{
		Name:           info.Name,
		Capacity:       info.Capacity,
		Used:           info.Used,
		ProducerCursor: info.ProducerCursor,
		Consumers:      info.Consumers,
	})
}

func (s *Server) handleListBuffers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	buffers := s.manager.ListBuffers()
	writeJSON(w, http.StatusOK, common.ListBufferResponse{Buffers: buffers})
}
