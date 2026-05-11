package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"

	"tcp-framing/framing"
	"tcp-framing/protocol"
)

type Server struct {
	config *framing.FramerOptions
	mu     sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		config: framing.DefaultFramerOptions(),
	}
}

func (s *Server) getConfig() *framing.FramerOptions {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &framing.FramerOptions{
		Mode:         s.config.Mode,
		HeaderSize:   s.config.HeaderSize,
		ByteOrder:    s.config.ByteOrder,
		MaxFrameSize: s.config.MaxFrameSize,
		Delimiter:    s.config.Delimiter,
	}
}

func (s *Server) setConfig(cfg *framing.FramerOptions) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = cfg
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cfg := s.getConfig()
	resp := protocol.GetConfigResponse{
		Mode:         cfg.Mode,
		HeaderSize:   cfg.HeaderSize,
		ByteOrder:    cfg.ByteOrder,
		MaxFrameSize: cfg.MaxFrameSize,
		Delimiter:    cfg.Delimiter,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleSetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	var req protocol.ConfigRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	newCfg := &framing.FramerOptions{
		Mode:         req.Mode,
		HeaderSize:   req.HeaderSize,
		ByteOrder:    req.ByteOrder,
		MaxFrameSize: req.MaxFrameSize,
		Delimiter:    req.Delimiter,
	}
	testFramer, err := framing.NewFramer(newCfg)
	if err != nil {
		resp := protocol.ConfigResponse{Success: false, Message: err.Error()}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}
	_ = testFramer
	s.setConfig(newCfg)
	resp := protocol.ConfigResponse{Success: true, Message: "config updated"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleEcho(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	var req protocol.EchoRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cfg := s.getConfig()
	if req.Mode != "" {
		cfg.Mode = req.Mode
	}
	framer, err := framing.NewFramer(cfg)
	if err != nil {
		resp := protocol.EchoResponse{Success: false, Message: err.Error()}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}
	encoded, err := framer.EncodeBatch(req.Messages)
	if err != nil {
		resp := protocol.EchoResponse{Success: false, Message: err.Error()}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}
	framer.Feed(encoded)
	decoded, err := framer.Decode()
	if err != nil {
		resp := protocol.EchoResponse{Success: false, Message: err.Error()}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}
	resp := protocol.EchoResponse{Success: true, Messages: decoded}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleTCPConnection(conn net.Conn) {
	defer conn.Close()
	cfg := s.getConfig()
	framer, err := framing.NewFramer(cfg)
	if err != nil {
		log.Printf("failed to create framer: %v", err)
		return
	}
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("read error: %v", err)
			}
			return
		}
		framer.Feed(buf[:n])
		messages, err := framer.Decode()
		if err != nil {
			log.Printf("decode error: %v, closing connection", err)
			return
		}
		if len(messages) == 0 {
			continue
		}
		response, err := framer.EncodeBatch(messages)
		if err != nil {
			log.Printf("encode error: %v", err)
			return
		}
		if _, err := conn.Write(response); err != nil {
			log.Printf("write error: %v", err)
			return
		}
	}
}

func main() {
	srv := NewServer()
	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			srv.handleGetConfig(w, r)
		} else if r.Method == http.MethodPost {
			srv.handleSetConfig(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/echo", srv.handleEcho)
	go func() {
		listener, err := net.Listen("tcp", ":8401")
		if err != nil {
			log.Fatalf("failed to start TCP server: %v", err)
		}
		defer listener.Close()
		log.Println("TCP server listening on :8401")
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("accept error: %v", err)
				continue
			}
			go srv.handleTCPConnection(conn)
		}
	}()
	fmt.Println("HTTP server listening on :8400")
	fmt.Println("  GET  /config - get current configuration")
	fmt.Println("  POST /config - update configuration")
	fmt.Println("  POST /echo   - test framing with messages")
	fmt.Println("TCP server listening on :8401")
	log.Fatal(http.ListenAndServe(":8400", nil))
}
