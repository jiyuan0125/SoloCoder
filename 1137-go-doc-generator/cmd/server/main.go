package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"godoc-generator/internal/api"
	"godoc-generator/internal/parser"
)

type Server struct {
	docStore *DocumentStore
	mux      *http.ServeMux
}

type DocumentStore struct {
	mu    sync.RWMutex
	docs  map[string]*parser.PackageDoc
	opts  parser.ParseOptions
}

func NewDocumentStore() *DocumentStore {
	return &DocumentStore{
		docs: make(map[string]*parser.PackageDoc),
		opts: parser.ParseOptions{},
	}
}

func (s *DocumentStore) Store(pkgName string, doc *parser.PackageDoc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.docs[pkgName] = doc
}

func (s *DocumentStore) Get(pkgName string) (*parser.PackageDoc, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	doc, ok := s.docs[pkgName]
	return doc, ok
}

func (s *DocumentStore) GetAll() []*parser.PackageDoc {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*parser.PackageDoc, 0, len(s.docs))
	for _, doc := range s.docs {
		result = append(result, doc)
	}
	return result
}

func NewServer() *Server {
	s := &Server{
		docStore: NewDocumentStore(),
		mux:      http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/upload", s.handleUpload)
	s.mux.HandleFunc("/api/package", s.handleGetPackage)
	s.mux.HandleFunc("/api/type", s.handleGetType)
	s.mux.HandleFunc("/api/function", s.handleGetFunction)
	s.mux.HandleFunc("/api/search", s.handleSearch)
	s.mux.HandleFunc("/api/export", s.handleExport)
	s.mux.HandleFunc("/api/health", s.handleHealth)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	s.mux.ServeHTTP(w, r)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, api.ErrorResponse{
		Success: false,
		Error:   message,
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	contentType := r.Header.Get("Content-Type")

	var uploadReq api.UploadRequest

	if strings.Contains(contentType, "multipart/form-data") {
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse form: %v", err))
			return
		}

		uploadReq.IncludeUnexported = r.FormValue("include_unexported") == "true"

		files := r.MultipartForm.File["files"]
		for _, fh := range files {
			file, err := fh.Open()
			if err != nil {
				writeError(w, http.StatusBadRequest, fmt.Sprintf("failed to open file: %v", err))
				return
			}
			content, err := io.ReadAll(file)
			file.Close()
			if err != nil {
				writeError(w, http.StatusBadRequest, fmt.Sprintf("failed to read file: %v", err))
				return
			}
			uploadReq.Files = append(uploadReq.Files, api.FileUpload{
				Filename: fh.Filename,
				Content:  string(content),
			})
		}
	} else {
		err := json.NewDecoder(r.Body).Decode(&uploadReq)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("failed to decode request: %v", err))
			return
		}
	}

	if len(uploadReq.Files) == 0 {
		writeError(w, http.StatusBadRequest, "no files provided")
		return
	}

	opts := parser.ParseOptions{
		IncludeUnexported: uploadReq.IncludeUnexported,
	}
	p := parser.NewParser(opts)

	tmpDir, err := os.MkdirTemp("", "godoc-upload-*")
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create temp dir: %v", err))
		return
	}
	defer os.RemoveAll(tmpDir)

	for _, f := range uploadReq.Files {
		filePath := filepath.Join(tmpDir, f.Filename)
		err := os.WriteFile(filePath, []byte(f.Content), 0644)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to write file: %v", err))
			return
		}
	}

	doc, err := p.ParsePackage(tmpDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to parse package: %v", err))
		return
	}

	s.docStore.Store(doc.Name, doc)

	writeJSON(w, http.StatusOK, api.UploadResponse{
		Success: true,
		Message: "package uploaded and parsed successfully",
		Package: doc.Name,
	})
}

func (s *Server) handleGetPackage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	pkgName := r.URL.Query().Get("name")
	if pkgName == "" {
		allDocs := s.docStore.GetAll()
		if len(allDocs) == 0 {
			writeJSON(w, http.StatusOK, api.PackageResponse{
				Success: true,
				Message: "no packages available",
				Data:    []interface{}{},
			})
			return
		}
		writeJSON(w, http.StatusOK, api.PackageResponse{
			Success: true,
			Message: fmt.Sprintf("found %d packages", len(allDocs)),
			Data:    allDocs,
		})
		return
	}

	doc, ok := s.docStore.Get(pkgName)
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("package %s not found", pkgName))
		return
	}

	writeJSON(w, http.StatusOK, api.PackageResponse{
		Success: true,
		Message: "package found",
		Data:    doc,
	})
}

func (s *Server) handleGetType(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	pkgName := r.URL.Query().Get("package")
	typeName := r.URL.Query().Get("name")

	if typeName == "" {
		writeError(w, http.StatusBadRequest, "type name is required")
		return
	}

	var targetType *parser.Type

	if pkgName != "" {
		doc, ok := s.docStore.Get(pkgName)
		if !ok {
			writeError(w, http.StatusNotFound, fmt.Sprintf("package %s not found", pkgName))
			return
		}
		for _, t := range doc.Types {
			if t.Name == typeName {
				targetType = &t
				break
			}
		}
	} else {
		allDocs := s.docStore.GetAll()
		for _, doc := range allDocs {
			for _, t := range doc.Types {
				if t.Name == typeName {
					targetType = &t
					break
				}
			}
			if targetType != nil {
				break
			}
		}
	}

	if targetType == nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("type %s not found", typeName))
		return
	}

	writeJSON(w, http.StatusOK, api.PackageResponse{
		Success: true,
		Message: "type found",
		Data:    targetType,
	})
}

func (s *Server) handleGetFunction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	pkgName := r.URL.Query().Get("package")
	funcName := r.URL.Query().Get("name")

	if funcName == "" {
		writeError(w, http.StatusBadRequest, "function name is required")
		return
	}

	var targetFunc *parser.Function

	if pkgName != "" {
		doc, ok := s.docStore.Get(pkgName)
		if !ok {
			writeError(w, http.StatusNotFound, fmt.Sprintf("package %s not found", pkgName))
			return
		}
		for _, f := range doc.Functions {
			if f.Name == funcName {
				targetFunc = &f
				break
			}
		}
	} else {
		allDocs := s.docStore.GetAll()
		for _, doc := range allDocs {
			for _, f := range doc.Functions {
				if f.Name == funcName {
					targetFunc = &f
					break
				}
			}
			if targetFunc != nil {
				break
			}
		}
	}

	if targetFunc == nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("function %s not found", funcName))
		return
	}

	writeJSON(w, http.StatusOK, api.PackageResponse{
		Success: true,
		Message: "function found",
		Data:    targetFunc,
	})
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "search query is required")
		return
	}

	includeUnexported := r.URL.Query().Get("include_unexported") == "true"

	allDocs := s.docStore.GetAll()
	var allResults []api.SearchResult

	for _, doc := range allDocs {
		results := parser.Search(doc, query, includeUnexported)
		for _, r := range results {
			allResults = append(allResults, api.SearchResult{
				Kind:     r.Kind,
				Name:     r.Name,
				Comment:  r.Comment,
				Summary:  r.Summary,
				Exported: r.Exported,
			})
		}
	}

	writeJSON(w, http.StatusOK, api.SearchResponse{
		Success: true,
		Message: fmt.Sprintf("found %d results", len(allResults)),
		Results: allResults,
	})
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	pkgName := r.URL.Query().Get("package")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	var doc *parser.PackageDoc
	if pkgName != "" {
		var ok bool
		doc, ok = s.docStore.Get(pkgName)
		if !ok {
			writeError(w, http.StatusNotFound, fmt.Sprintf("package %s not found", pkgName))
			return
		}
	} else {
		allDocs := s.docStore.GetAll()
		if len(allDocs) == 0 {
			writeError(w, http.StatusNotFound, "no packages available")
			return
		}
		doc = allDocs[0]
	}

	var content []byte
	var contentType string
	var err error

	switch strings.ToLower(format) {
	case "json":
		content, err = parser.ToJSON(doc)
		contentType = "application/json"
	case "markdown", "md":
		content, err = parser.ToMarkdown(doc)
		contentType = "text/markdown"
	default:
		writeError(w, http.StatusBadRequest, fmt.Sprintf("unsupported format: %s", format))
		return
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to export: %v", err))
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.%s", doc.Name, format))
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := NewServer()

	log.Printf("Go Doc Generator Server starting on port %s...", port)
	log.Printf("Endpoints:")
	log.Printf("  POST /api/upload - Upload Go source files")
	log.Printf("  GET  /api/package - Get package documentation")
	log.Printf("  GET  /api/type - Get type documentation")
	log.Printf("  GET  /api/function - Get function documentation")
	log.Printf("  GET  /api/search - Search documentation")
	log.Printf("  GET  /api/export - Export documentation")
	log.Printf("  GET  /api/health - Health check")

	if err := http.ListenAndServe(":"+port, server); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
