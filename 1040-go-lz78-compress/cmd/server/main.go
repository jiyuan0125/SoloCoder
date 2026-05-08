package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"lz78-tool/pkg/api"
	"lz78-tool/pkg/lz78"
)

type Server struct {
	compressor  *lz78.LZ78
	currentDict map[int]string
	dictMutex   sync.RWMutex
}

func NewServer(maxDictSize ...int) *Server {
	return &Server{
		compressor: lz78.New(maxDictSize...),
		currentDict: map[int]string{
			0: "",
		},
	}
}

func (s *Server) handleCompress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.CompressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	indexes, dict, err := s.compressor.CompressWithDict(req.Text)
	if err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.dictMutex.Lock()
	s.currentDict = dict
	s.dictMutex.Unlock()

	resp := api.CompressResponse{
		Success: true,
		Indexes: indexes,
	}
	sendJSON(w, resp, http.StatusOK)
}

func (s *Server) handleDecompress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.DecompressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	text, dict, err := s.compressor.DecompressWithDict(req.Indexes)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.dictMutex.Lock()
	s.currentDict = dict
	s.dictMutex.Unlock()

	resp := api.DecompressResponse{
		Success: true,
		Text:    text,
	}
	sendJSON(w, resp, http.StatusOK)
}

func (s *Server) handleDictStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.dictMutex.RLock()
	dict := make(map[int]string, len(s.currentDict))
	for k, v := range s.currentDict {
		dict[k] = v
	}
	s.dictMutex.RUnlock()

	resp := api.DictStatusResponse{
		Success: true,
		Count:   len(dict),
		Entries: dict,
	}
	sendJSON(w, resp, http.StatusOK)
}

func (s *Server) handleResetDict(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.dictMutex.Lock()
	s.currentDict = map[int]string{
		0: "",
	}
	s.dictMutex.Unlock()

	resp := api.DictStatusResponse{
		Success: true,
		Count:   1,
		Entries: map[int]string{0: ""},
	}
	sendJSON(w, resp, http.StatusOK)
}

func sendJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, errMsg string, status int) {
	resp := api.ErrorResponse{
		Success: false,
		Error:   errMsg,
	}
	sendJSON(w, resp, status)
}

func main() {
	server := NewServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/compress", server.handleCompress)
	mux.HandleFunc("/api/decompress", server.handleDecompress)
	mux.HandleFunc("/api/dict", server.handleDictStatus)
	mux.HandleFunc("/api/dict/reset", server.handleResetDict)

	log.Println("LZ78 compression server starting on :8080...")
	log.Println("Endpoints:")
	log.Println("  POST /api/compress   - Compress text")
	log.Println("  POST /api/decompress - Decompress indexes")
	log.Println("  GET  /api/dict       - Get dictionary status")
	log.Println("  POST /api/dict/reset - Reset dictionary")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
