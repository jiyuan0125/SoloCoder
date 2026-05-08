package server

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"file-watch-debounce/common"
	"file-watch-debounce/watcher"
)

type Server struct {
	manager *watcher.Manager
}

func NewServer() *Server {
	return &Server{
		manager: watcher.NewManager(),
	}
}

func (s *Server) generateID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return time.Now().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(buf)
}

func (s *Server) AddWatchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONResponse(w, http.StatusBadRequest, common.AddWatchResponse{
			Success: false,
			Error:   "Failed to read request body: " + err.Error(),
		})
		return
	}
	defer r.Body.Close()

	var req common.AddWatchRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, common.AddWatchResponse{
			Success: false,
			Error:   "Invalid JSON: " + err.Error(),
		})
		return
	}

	id := s.generateID()
	config := watcher.WatchConfig{
		Path:     req.Path,
		Debounce: req.Debounce,
		MaxWait:  req.MaxWait,
		Callback: s.manager.CreateCallbackWithCustom(func(event interface{}) {
			if req.CallbackURL != "" {
				s.sendCallback(req.CallbackURL, event)
			}
		}),
	}

	if err := s.manager.AddWatch(id, config); err != nil {
		writeJSONResponse(w, http.StatusInternalServerError, common.AddWatchResponse{
			Success: false,
			Error:   "Failed to add watch: " + err.Error(),
		})
		return
	}

	writeJSONResponse(w, http.StatusOK, common.AddWatchResponse{
		Success: true,
		ID:      id,
	})
}

func (s *Server) RemoveWatchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONResponse(w, http.StatusBadRequest, common.RemoveWatchResponse{
			Success: false,
			Error:   "Failed to read request body: " + err.Error(),
		})
		return
	}
	defer r.Body.Close()

	var req common.RemoveWatchRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, common.RemoveWatchResponse{
			Success: false,
			Error:   "Invalid JSON: " + err.Error(),
		})
		return
	}

	if err := s.manager.RemoveWatch(req.ID); err != nil {
		writeJSONResponse(w, http.StatusInternalServerError, common.RemoveWatchResponse{
			Success: false,
			Error:   "Failed to remove watch: " + err.Error(),
		})
		return
	}

	writeJSONResponse(w, http.StatusOK, common.RemoveWatchResponse{
		Success: true,
	})
}

func (s *Server) ListWatchesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	watches := s.manager.ListWatches()
	infoList := make([]common.WatchInformation, 0, len(watches))

	for id, w := range watches {
		config := w.Config()
		infoList = append(infoList, common.WatchInformation{
			ID:          id,
			Path:        config.Path,
			Debounce:    config.Debounce,
			MaxWait:     config.MaxWait,
			Active:      w.IsActive(),
		})
	}

	writeJSONResponse(w, http.StatusOK, common.ListWatchesResponse{
		Success: true,
		Watches: infoList,
	})
}

func (s *Server) ListEventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	recentEvents := s.manager.GetRecentEvents()
	events := make([]common.Event, 0, len(recentEvents))

	for _, e := range recentEvents {
		events = append(events, common.Event{
			Path:      e.Path,
			OldPath:   e.OldPath,
			Type:      common.EventType(e.Type),
			Timestamp: e.Timestamp,
		})
	}

	writeJSONResponse(w, http.StatusOK, common.ListEventsResponse{
		Success: true,
		Events:  events,
	})
}

func (s *Server) sendCallback(url string, event interface{}) {
	we, ok := event.(*watcher.Event)
	if !ok {
		return
	}

	ce := common.Event{
		Path:      we.Path,
		OldPath:   we.OldPath,
		Type:      common.EventType(we.Type),
		Timestamp: we.Timestamp,
	}

	body, err := json.Marshal(ce)
	if err != nil {
		return
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	_, _ = client.Do(req)
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}
