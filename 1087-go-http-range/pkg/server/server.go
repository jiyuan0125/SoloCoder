package server

import (
	"bytes"
	"net/http"
	"time"

	"example.com/httprange/pkg/rangehandler"
)

type Server struct {
	content []byte
	etag    string
	modTime time.Time
}

func NewServer(content []byte) *Server {
	return &Server{
		content: content,
		etag:    rangehandler.GenerateETag(content),
		modTime: time.Now(),
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	info := rangehandler.ResourceInfo{
		Content:      s.content,
		ContentType:  "application/octet-stream",
		ETag:         s.etag,
		LastModified: s.modTime,
	}

	resp := rangehandler.HandleRangeRequest(r, info)

	for key, values := range resp.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	if r.Method == http.MethodGet && resp.Body != nil {
		w.Write(resp.Body)
	}
}

func (s *Server) GetETag() string {
	return s.etag
}

func (s *Server) GetLastModified() time.Time {
	return s.modTime
}

func (s *Server) GetContent() []byte {
	return bytes.Clone(s.content)
}
