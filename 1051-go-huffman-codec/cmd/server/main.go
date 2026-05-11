package main

import (
	"encoding/json"
	"fmt"
	"huffman-codec/common"
	"huffman-codec/huffman"
	"log"
	"net/http"
	"sync"
)

type Server struct {
	currentFreq huffman.FreqMap
	mu          sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		currentFreq: make(huffman.FreqMap),
	}
}

func (s *Server) encodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.EncodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	originalSize := len(req.Text)

	if req.Text == "" {
		s.mu.Lock()
		s.currentFreq = make(huffman.FreqMap)
		s.mu.Unlock()

		resp := common.EncodeResponse{
			Data:        "",
			PaddingBits: 0,
			TreeData:    "",
			OriginalSize: 0,
			EncodedSize:  0,
			Frequencies:  make(map[string]int),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	freq := huffman.CountFrequencies(req.Text)
	s.mu.Lock()
	s.currentFreq = freq.Copy()
	s.mu.Unlock()

	tree := huffman.BuildHuffmanTree(freq)
	codeTable := huffman.BuildCodeTable(tree)

	encodedData, paddingBits, err := huffman.Encode(req.Text, codeTable)
	if err != nil {
		http.Error(w, fmt.Sprintf("Encode error: %v", err), http.StatusInternalServerError)
		return
	}

	treeData := huffman.SerializeTree(tree)

	resp := common.EncodeResponse{
		Data:         common.Base64Encode(encodedData),
		PaddingBits:  paddingBits,
		TreeData:     common.Base64Encode(treeData),
		OriginalSize: originalSize,
		EncodedSize:  len(encodedData),
		Frequencies:  common.FrequenciesToJSON(freq),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) decodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.DecodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Data == "" {
		resp := common.DecodeResponse{
			Text:         "",
			OriginalSize: 0,
			EncodedSize:  0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	encodedData, err := common.Base64Decode(req.Data)
	if err != nil {
		http.Error(w, "Invalid base64 data", http.StatusBadRequest)
		return
	}

	treeData, err := common.Base64Decode(req.TreeData)
	if err != nil {
		http.Error(w, "Invalid tree data", http.StatusBadRequest)
		return
	}

	tree, err := huffman.DeserializeTree(treeData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Deserialize tree error: %v", err), http.StatusBadRequest)
		return
	}

	decodedText, err := huffman.Decode(encodedData, req.PaddingBits, tree)
	if err != nil {
		http.Error(w, fmt.Sprintf("Decode error: %v", err), http.StatusInternalServerError)
		return
	}

	resp := common.DecodeResponse{
		Text:         decodedText,
		OriginalSize: len(decodedText),
		EncodedSize:  len(encodedData),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) frequencyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	freq := s.currentFreq.Copy()
	s.mu.RUnlock()

	resp := common.FrequencyResponse{
		Frequencies: common.FrequenciesToJSON(freq),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	server := NewServer()

	http.HandleFunc("/encode", server.encodeHandler)
	http.HandleFunc("/decode", server.decodeHandler)
	http.HandleFunc("/frequencies", server.frequencyHandler)

	addr := ":8400"
	log.Printf("Server starting on %s", addr)
	log.Printf("Endpoints:")
	log.Printf("  POST /encode - Encode text")
	log.Printf("  POST /decode - Decode data")
	log.Printf("  GET  /frequencies - Get current frequency table")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
