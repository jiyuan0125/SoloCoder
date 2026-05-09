package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"zstd-tool/protocol"
	"zstd-tool/zstdlib"
)

const (
	defaultPort       = ":8080"
	defaultDictDir    = "./dictionaries"
	defaultUploadDir  = "./uploads"
	maxUploadSize     = 1024 * 1024 * 1024
)

type Server struct {
	compressor     *zstdlib.Compressor
	dictManager    *zstdlib.DictionaryManager
	uploadDir      string
	dictDir        string
}

func NewServer(dictDir, uploadDir string) (*Server, error) {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, err
	}

	dictManager, err := zstdlib.NewDictionaryManager(dictDir)
	if err != nil {
		return nil, err
	}

	if err := dictManager.LoadExistingDictionaries(); err != nil {
		log.Printf("Warning: failed to load existing dictionaries: %v", err)
	}

	compressor := zstdlib.NewCompressor()

	dicts := dictManager.ListDictionaries()
	for _, dict := range dicts {
		data, ok := dictManager.GetDictionary(dict.Name)
		if ok {
			_ = compressor.LoadDictionary(dict.Name, data)
		}
	}

	return &Server{
		compressor:  compressor,
		dictManager: dictManager,
		uploadDir:   uploadDir,
		dictDir:     dictDir,
	}, nil
}

func (s *Server) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON: %v", err)
	}
}

func (s *Server) handleCompress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Failed to parse form data",
		})
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "No file provided",
		})
		return
	}
	defer file.Close()

	level := protocol.LevelDefault
	if levelStr := r.FormValue("level"); levelStr != "" {
		switch levelStr {
		case "fastest":
			level = protocol.LevelFastest
		case "default":
			level = protocol.LevelDefault
		case "best":
			level = protocol.LevelBest
		}
	}

	dictName := r.FormValue("dict")

	tempDir, err := os.MkdirTemp(s.uploadDir, "compress-")
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Failed to create temp directory",
		})
		return
	}
	defer os.RemoveAll(tempDir)

	inputPath := filepath.Join(tempDir, fileHeader.Filename)
	inputFile, err := os.Create(inputPath)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Failed to save uploaded file",
		})
		return
	}

	io.Copy(inputFile, file)
	inputFile.Close()

	outputPath := filepath.Join(tempDir, fileHeader.Filename+".zst")
	originalSize, compressedSize, err := s.compressor.CompressFile(inputPath, outputPath, level, dictName)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Compression failed: %v", err),
		})
		return
	}

	outputFile, err := os.Open(outputPath)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Failed to read compressed file",
		})
		return
	}
	defer outputFile.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+fileHeader.Filename+".zst")
	w.Header().Set("X-Original-Size", fmt.Sprintf("%d", originalSize))
	w.Header().Set("X-Compressed-Size", fmt.Sprintf("%d", compressedSize))

	io.Copy(w, outputFile)
}

func (s *Server) handleDecompress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Failed to parse form data",
		})
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "No file provided",
		})
		return
	}
	defer file.Close()

	dictName := r.FormValue("dict")

	tempDir, err := os.MkdirTemp(s.uploadDir, "decompress-")
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Failed to create temp directory",
		})
		return
	}
	defer os.RemoveAll(tempDir)

	inputPath := filepath.Join(tempDir, fileHeader.Filename)
	inputFile, err := os.Create(inputPath)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Failed to save uploaded file",
		})
		return
	}

	io.Copy(inputFile, file)
	inputFile.Close()

	originalFilename := fileHeader.Filename
	if len(originalFilename) > 4 && originalFilename[len(originalFilename)-4:] == ".zst" {
		originalFilename = originalFilename[:len(originalFilename)-4]
	}

	outputPath := filepath.Join(tempDir, originalFilename)
	originalSize, compressedSize, err := s.compressor.DecompressFile(inputPath, outputPath, dictName)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Decompression failed: %v", err),
		})
		return
	}

	outputFile, err := os.Open(outputPath)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Failed to read decompressed file",
		})
		return
	}
	defer outputFile.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+originalFilename)
	w.Header().Set("X-Original-Size", fmt.Sprintf("%d", originalSize))
	w.Header().Set("X-Compressed-Size", fmt.Sprintf("%d", compressedSize))

	io.Copy(w, outputFile)
}

func (s *Server) handleTrainDict(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.TrainDictionaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, protocol.TrainDictionaryResponse{
			Success:      false,
			ErrorMessage: "Invalid JSON request",
		})
		return
	}

	if req.DictionaryName == "" {
		s.writeJSON(w, http.StatusBadRequest, protocol.TrainDictionaryResponse{
			Success:      false,
			ErrorMessage: "Dictionary name is required",
		})
		return
	}

	if len(req.FilePaths) == 0 {
		s.writeJSON(w, http.StatusBadRequest, protocol.TrainDictionaryResponse{
			Success:      false,
			ErrorMessage: "At least one file path is required",
		})
		return
	}

	info, err := s.dictManager.TrainDictionary(req.DictionaryName, req.FilePaths)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, protocol.TrainDictionaryResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Training failed: %v", err),
		})
		return
	}

	data, ok := s.dictManager.GetDictionary(req.DictionaryName)
	if ok {
		_ = s.compressor.LoadDictionary(req.DictionaryName, data)
	}

	s.writeJSON(w, http.StatusOK, protocol.TrainDictionaryResponse{
		Success:     true,
		DictName:    info.Name,
		DictSize:    info.DictSize,
		SampleCount: info.SampleCount,
	})
}

func (s *Server) handleListDicts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dicts := s.dictManager.ListDictionaries()
	var infos []protocol.DictionaryInfo
	for _, d := range dicts {
		infos = append(infos, protocol.DictionaryInfo{
			Name:        d.Name,
			SampleCount: d.SampleCount,
			DictSize:    d.DictSize,
			CreatedAt:   time.Unix(d.CreatedAt, 0).Format(time.RFC3339),
		})
	}

	s.writeJSON(w, http.StatusOK, protocol.ListDictionaryResponse{
		Success:      true,
		Dictionaries: infos,
	})
}

func main() {
	server, err := NewServer(defaultDictDir, defaultUploadDir)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	http.HandleFunc("/api/compress", server.handleCompress)
	http.HandleFunc("/api/decompress", server.handleDecompress)
	http.HandleFunc("/api/dict/train", server.handleTrainDict)
	http.HandleFunc("/api/dict/list", server.handleListDicts)

	log.Printf("ZSTD Server started on port %s", defaultPort)
	if err := http.ListenAndServe(defaultPort, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
