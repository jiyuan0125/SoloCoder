package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"http-replay/api"
	"http-replay/replay"
)

type Handler struct {
	storage  *replay.Storage
	player   *replay.Player
	recorders map[string]*replay.Recorder
	mu        sync.RWMutex
}

func NewHandler(sessionsDir string) *Handler {
	storage := replay.NewStorage(sessionsDir)
	return &Handler{
		storage:   storage,
		player:    replay.NewPlayer(storage),
		recorders: make(map[string]*replay.Recorder),
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/replay/record", h.createRecord)
	mux.HandleFunc("/replay/record/", h.handleRecordWithID)
	mux.HandleFunc("/replay/play", h.playRecord)
	mux.HandleFunc("/replay/sessions", h.listSessions)
}

func (h *Handler) createRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	var req api.CreateRecordRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.TargetURL == "" {
		http.Error(w, "target_url is required", http.StatusBadRequest)
		return
	}

	port, err := getAvailablePort()
	if err != nil {
		http.Error(w, "failed to get available port", http.StatusInternalServerError)
		return
	}
	listenAddr := fmt.Sprintf(":%d", port)

	recorder, err := replay.NewRecorder(req.TargetURL, req.VariableRules, listenAddr, func(data *api.RecordedSessionData) {
		_, _ = h.storage.SaveSession(data)
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := recorder.Start(); err != nil {
		http.Error(w, "failed to start recorder", http.StatusInternalServerError)
		return
	}

	h.mu.Lock()
	h.recorders[recorder.SessionID()] = recorder
	h.mu.Unlock()

	resp := api.CreateRecordResponse{
		SessionID: recorder.SessionID(),
		ProxyAddr: listenAddr,
		TargetURL: req.TargetURL,
		StartTime: time.Now(),
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) handleRecordWithID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/replay/record/")

	if strings.HasSuffix(path, "/stop") {
		sessionID := strings.TrimSuffix(path, "/stop")
		h.stopRecord(w, r, sessionID)
		return
	}

	h.getRecord(w, r, path)
}

func (h *Handler) stopRecord(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.mu.Lock()
	recorder, exists := h.recorders[sessionID]
	delete(h.recorders, sessionID)
	h.mu.Unlock()

	if !exists {
		http.Error(w, "session not found or already stopped", http.StatusNotFound)
		return
	}

	data, err := recorder.Stop()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if data == nil {
		http.Error(w, "session already stopped", http.StatusBadRequest)
		return
	}

	filePath, err := h.storage.SaveSession(data)
	if err != nil {
		http.Error(w, "failed to save session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := api.StopRecordResponse{
		SessionID:    sessionID,
		EndTime:      time.Now(),
		RequestCount: data.Meta.RequestCount,
		FilePath:     filePath,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) getRecord(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := h.storage.LoadSession(sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, session)
}

func (h *Handler) playRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	var req api.PlayRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "session_id is required", http.StatusBadRequest)
		return
	}

	h.mu.RLock()
	_, isActive := h.recorders[req.SessionID]
	h.mu.RUnlock()

	if isActive {
		http.Error(w, "session is still recording", http.StatusBadRequest)
		return
	}

	opts := replay.PlayOptions{
		TargetURL:      req.TargetURL,
		Mode:           req.Mode,
		Concurrency:    req.Concurrency,
		ReplaceRules:   req.ReplaceRules,
		VariableValues: req.VariableValues,
	}

	result, err := h.player.Play(req.SessionID, opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) listSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessions, err := h.storage.ListSessions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.mu.RLock()
	activeSessionIDs := make(map[string]bool)
	for id := range h.recorders {
		activeSessionIDs[id] = true
	}
	h.mu.RUnlock()

	for i := range sessions {
		if activeSessionIDs[sessions[i].SessionID] {
			sessions[i].Active = true
			sessions[i].FilePath = ""
		}
	}

	resp := api.RecordSessionList{
		Sessions: sessions,
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func getAvailablePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}
