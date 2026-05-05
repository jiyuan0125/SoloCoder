package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"json-extractor/internal/jsonparser"
	"json-extractor/pkg/common"
)

type Server struct {
	host      string
	port      int
	extractor *jsonparser.Extractor
	clients   map[net.Conn]struct{}
	mu        sync.Mutex
}

func NewServer(host string, port int) *Server {
	return &Server{
		host:      host,
		port:      port,
		extractor: jsonparser.NewExtractor(),
		clients:   make(map[net.Conn]struct{}),
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	defer listener.Close()

	log.Printf("JSON Extractor Server started on %s:%d", s.host, s.port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down server...")
		s.Stop()
		os.Exit(0)
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-sigChan:
				return nil
			default:
				log.Printf("Failed to accept connection: %v", err)
				continue
			}
		}

		s.mu.Lock()
		s.clients[conn] = struct{}{}
		s.mu.Unlock()

		go s.handleClient(conn)
	}
}

func (s *Server) handleClient(conn net.Conn) {
	defer func() {
		conn.Close()
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
	}()

	log.Printf("New client connected: %s", conn.RemoteAddr())

	req, err := common.ReceiveRequest(conn)
	if err != nil {
		log.Printf("Failed to receive request: %v", err)
		return
	}

	var results []string
	var extractErr error

	if req.Filename == "-" || req.Filename == "" {
		reader, recvErr := common.ReceiveStdinData(conn)
		if recvErr != nil {
			resp := &common.Response{Error: recvErr.Error()}
			common.SendResponse(conn, resp)
			log.Printf("Failed to receive stdin data: %v", recvErr)
			return
		}
		results, extractErr = s.extractor.ExtractFromReader(reader, req.Fields, req.OutputJSON, req.Compact)
	} else {
		results, extractErr = s.extractor.ExtractFromFile(req.Filename, req.Fields, req.OutputJSON, req.Compact)
	}

	resp := &common.Response{}
	if extractErr != nil {
		resp.Error = extractErr.Error()
		log.Printf("Extraction error: %v", extractErr)
	} else {
		resp.Results = results
	}

	if err := common.SendResponse(conn, resp); err != nil {
		log.Printf("Failed to send response: %v", err)
	}

	log.Printf("Client disconnected: %s", conn.RemoteAddr())
}

func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for conn := range s.clients {
		conn.Close()
		delete(s.clients, conn)
	}
}
