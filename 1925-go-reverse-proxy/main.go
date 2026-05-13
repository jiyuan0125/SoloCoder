package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
)

type Mode string

const (
	ModeOff       Mode = "off"
	ModeRecording Mode = "recording"
	ModeReplaying Mode = "replaying"
)

type RecordedRequest struct {
	Method string            `json:"method"`
	Path   string            `json:"path"`
	Header map[string]string `json:"header"`
	Body   string            `json:"body"`
}

type RecordedResponse struct {
	StatusCode int               `json:"statusCode"`
	Header     map[string]string `json:"header"`
	Body       string            `json:"body"`
}

type RecordingEntry struct {
	Request  RecordedRequest  `json:"request"`
	Response RecordedResponse `json:"response"`
}

type State struct {
	mu         sync.RWMutex
	mode       Mode
	backendURL string
	recordings []RecordingEntry
	replayIdx  int
}

func NewState() *State {
	return &State{
		mode:       ModeOff,
		backendURL: "",
		recordings: []RecordingEntry{},
		replayIdx:  0,
	}
}

func (s *State) GetMode() Mode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mode
}

func (s *State) SetMode(m Mode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mode = m
}

func (s *State) GetBackendURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.backendURL
}

func (s *State) SetBackendURL(u string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.backendURL = u
}

func (s *State) AddRecording(entry RecordingEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recordings = append(s.recordings, entry)
}

func (s *State) GetRecordingCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.recordings)
}

func (s *State) GetReplayIdx() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.replayIdx
}

func (s *State) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mode = ModeOff
	s.recordings = []RecordingEntry{}
	s.replayIdx = 0
}

func (s *State) TryTransition(next Mode) (bool, string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current := s.mode
	var allowed Mode

	switch current {
	case ModeOff:
		allowed = ModeRecording
	case ModeRecording:
		allowed = ModeReplaying
	case ModeReplaying:
		allowed = ModeOff
	default:
		allowed = ModeOff
	}

	if next == allowed {
		s.mode = next
		if next == ModeReplaying {
			s.replayIdx = 0
		}
		return true, ""
	}

	return false, fmt.Sprintf(`{"current":"%s","allowedNext":"%s"}`, current, allowed)
}

func (s *State) MatchAndAdvance(req RecordedRequest) (RecordingEntry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.replayIdx >= len(s.recordings) {
		s.mode = ModeOff
		return RecordingEntry{}, false
	}

	entry := s.recordings[s.replayIdx]

	if entry.Request.Method != req.Method ||
		entry.Request.Path != req.Path ||
		entry.Request.Body != req.Body {
		return RecordingEntry{}, false
	}

	s.replayIdx++
	if s.replayIdx >= len(s.recordings) {
		s.mode = ModeOff
	}

	return entry, true
}

func headersToMap(h http.Header) map[string]string {
	m := make(map[string]string)
	for k, v := range h {
		m[k] = strings.Join(v, ", ")
	}
	return m
}

func mapToHeaders(m map[string]string) http.Header {
	h := make(http.Header)
	for k, v := range m {
		h.Set(k, v)
	}
	return h
}

func readBodyAndRestore(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	return body, nil
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
	header     http.Header
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
		header:         make(http.Header),
	}
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	for k, v := range r.ResponseWriter.Header() {
		for _, vv := range v {
			r.ResponseWriter.Header().Set(k, vv)
		}
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func handleStatus(state *State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"mode":           state.GetMode(),
			"recordedCount":  state.GetRecordingCount(),
			"replayedCount":  state.GetReplayIdx(),
		})
	}
}

type ModeRequest struct {
	Next string `json:"next"`
}

func handleMode(state *State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ModeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		next := Mode(req.Next)
		if next != ModeOff && next != ModeRecording && next != ModeReplaying {
			http.Error(w, "Invalid mode", http.StatusBadRequest)
			return
		}

		ok, errorMsg := state.TryTransition(next)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(errorMsg))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

type BackendRequest struct {
	URL string `json:"url"`
}

func handleBackend(state *State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req BackendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.URL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		state.SetBackendURL(req.URL)
		w.WriteHeader(http.StatusNoContent)
	}
}

func createProxy(target string) (*httputil.ReverseProxy, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	return proxy, nil
}

func handleProxy(state *State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/admin/") {
			http.NotFound(w, r)
			return
		}

		mode := state.GetMode()

		switch mode {
		case ModeOff:
			http.Error(w, "Proxy is off", http.StatusServiceUnavailable)

		case ModeRecording:
			backendURL := state.GetBackendURL()
			if backendURL == "" {
				http.Error(w, "Backend URL not configured", http.StatusServiceUnavailable)
				return
			}

			proxy, err := createProxy(backendURL)
			if err != nil {
				http.Error(w, "Invalid backend URL", http.StatusInternalServerError)
				return
			}

			reqBody, err := readBodyAndRestore(r)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusInternalServerError)
				return
			}

			recorder := newResponseRecorder(w)
			proxy.ServeHTTP(recorder, r)

			entry := RecordingEntry{
				Request: RecordedRequest{
					Method: r.Method,
					Path:   r.URL.Path,
					Header: headersToMap(r.Header),
					Body:   string(reqBody),
				},
				Response: RecordedResponse{
					StatusCode: recorder.statusCode,
					Header:     headersToMap(recorder.ResponseWriter.Header()),
					Body:       recorder.body.String(),
				},
			}

			state.AddRecording(entry)

		case ModeReplaying:
			reqBody, err := readBodyAndRestore(r)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusInternalServerError)
				return
			}

			req := RecordedRequest{
				Method: r.Method,
				Path:   r.URL.Path,
				Header: headersToMap(r.Header),
				Body:   string(reqBody),
			}

			entry, matched := state.MatchAndAdvance(req)
			if !matched {
				http.NotFound(w, r)
				return
			}

			for k, v := range entry.Response.Header {
				w.Header().Set(k, v)
			}
			w.WriteHeader(entry.Response.StatusCode)
			w.Write([]byte(entry.Response.Body))
		}
	}
}

func main() {
	state := NewState()

	mux := http.NewServeMux()

	mux.HandleFunc("/admin/status", handleStatus(state))
	mux.HandleFunc("/admin/mode", handleMode(state))
	mux.HandleFunc("/admin/backend", handleBackend(state))
	mux.HandleFunc("/", handleProxy(state))

	port := os.Getenv("PORT")
	if port == "" {
		port = "9205"
	}

	addr := ":" + port
	fmt.Printf("Server starting on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
