package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/baru/commitlog/api"
	"github.com/baru/commitlog/commitlog"
)

type Server struct {
	cl *commitlog.CommitLog
}

func NewServer(dataDir string) (*Server, error) {
	cl, err := commitlog.New(dataDir, commitlog.DefaultConfig())
	if err != nil {
		return nil, err
	}
	return &Server{cl: cl}, nil
}

func (s *Server) Close() error {
	return s.cl.Close()
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func (s *Server) handleProduce(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	var req api.ProduceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if strings.TrimSpace(req.Topic) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("topic is required"))
		return
	}

	offset, err := s.cl.Append(req.Topic, req.Value)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, api.ProduceResponse{Offset: offset})
}

func (s *Server) handleFetch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	var req api.FetchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if strings.TrimSpace(req.Topic) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("topic is required"))
		return
	}

	if req.Max <= 0 {
		req.Max = 100
	}

	records, next, err := s.cl.Fetch(req.Topic, req.Offset, req.Max)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	messages := make([]api.Message, 0, len(records))
	for _, rec := range records {
		messages = append(messages, api.Message{
			Offset:    rec.Offset(),
			Topic:     req.Topic,
			Value:     rec.Value(),
			Timestamp: rec.Timestamp(),
		})
	}

	writeJSON(w, http.StatusOK, api.FetchResponse{
		Messages: messages,
		Next:     next,
	})
}

func (s *Server) handleCreateGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	var req api.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if strings.TrimSpace(req.GroupID) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("group_id is required"))
		return
	}
	if strings.TrimSpace(req.Topic) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("topic is required"))
		return
	}

	if err := s.cl.CreateGroup(req.GroupID, req.Topic); err != nil {
		writeJSON(w, http.StatusOK, api.CreateGroupResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, api.CreateGroupResponse{Success: true})
}

func (s *Server) handleJoinGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	var req api.JoinGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if strings.TrimSpace(req.GroupID) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("group_id is required"))
		return
	}
	if strings.TrimSpace(req.ConsumerID) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("consumer_id is required"))
		return
	}

	offset, err := s.cl.JoinGroup(req.GroupID, req.ConsumerID)
	if err != nil {
		writeJSON(w, http.StatusOK, api.JoinGroupResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, api.JoinGroupResponse{
		AssignedOffset: offset,
		Success:        true,
	})
}

func (s *Server) handleCommitOffset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	var req api.CommitOffsetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if strings.TrimSpace(req.GroupID) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("group_id is required"))
		return
	}
	if strings.TrimSpace(req.ConsumerID) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("consumer_id is required"))
		return
	}
	if req.Offset < 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("offset must be >= 0"))
		return
	}

	if err := s.cl.CommitOffset(req.GroupID, req.ConsumerID, req.Offset); err != nil {
		writeJSON(w, http.StatusOK, api.CommitOffsetResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, api.CommitOffsetResponse{Success: true})
}

func (s *Server) handleGetOffset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	var req api.GetOffsetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if strings.TrimSpace(req.GroupID) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("group_id is required"))
		return
	}
	if strings.TrimSpace(req.ConsumerID) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("consumer_id is required"))
		return
	}

	offset, err := s.cl.GetOffset(req.GroupID, req.ConsumerID)
	if err != nil {
		writeJSON(w, http.StatusOK, api.GetOffsetResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, api.GetOffsetResponse{
		Offset:  offset,
		Success: true,
	})
}

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dataDir := flag.String("data", "./data", "data directory")
	flag.Parse()

	server, err := NewServer(*dataDir)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}
	defer server.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/produce", server.handleProduce)
	mux.HandleFunc("/fetch", server.handleFetch)
	mux.HandleFunc("/groups/create", server.handleCreateGroup)
	mux.HandleFunc("/groups/join", server.handleJoinGroup)
	mux.HandleFunc("/groups/commit", server.handleCommitOffset)
	mux.HandleFunc("/groups/offset", server.handleGetOffset)

	log.Printf("server listening on %s, data dir: %s", *addr, *dataDir)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
