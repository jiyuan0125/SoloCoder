package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/qrcode/generator/pkg/common"
	"github.com/qrcode/generator/pkg/qrcode"
)

const (
	defaultPort = ":8080"
)

type Server struct {
	generator *qrcode.Generator
}

func NewServer() *Server {
	return &Server{
		generator: qrcode.NewGenerator(),
	}
}

func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.GenerateRequest
	contentType := r.Header.Get("Content-Type")

	if isMultipart(contentType) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			writeError(w, http.StatusBadRequest, "failed to parse multipart form: "+err.Error())
			return
		}

		req.Content = r.FormValue("content")
		req.ErrorLevel = common.ErrorLevel(r.FormValue("error_level"))

		if sizeStr := r.FormValue("size"); sizeStr != "" {
			if size, err := strconv.Atoi(sizeStr); err == nil {
				req.Size = size
			}
		}

		if file, _, err := r.FormFile("logo"); err == nil {
			defer file.Close()
			if logoData, err := io.ReadAll(file); err == nil {
				req.Logo = logoData
			}
		}
	} else {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to read request body")
			return
		}
		defer r.Body.Close()

		if err := json.Unmarshal(body, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
	}

	resp, err := s.generator.Generate(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, resp.Message)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(resp.Image)))
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Image)
}

func isMultipart(contentType string) bool {
	for i := 0; i+len("multipart/") <= len(contentType); i++ {
		if contentType[i:i+len("multipart/")] == "multipart/" {
			return true
		}
	}
	return false
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(common.GenerateResponse{
		Success: false,
		Message: message,
	})
}

func main() {
	server := NewServer()

	http.HandleFunc("/generate", server.handleGenerate)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	port := defaultPort
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = ":" + envPort
	}

	println("QR code server starting on", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		panic(err)
	}
}
