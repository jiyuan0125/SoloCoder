package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"

	"filescanner/internal/config"
	"filescanner/internal/protocol"
	"filescanner/internal/scanner"
)

type Server struct {
	address  string
	config   *config.Config
	listener net.Listener
	wg       sync.WaitGroup
	stopChan chan struct{}
}

func NewServer(address string, cfg *config.Config) *Server {
	return &Server{
		address:  address,
		config:   cfg,
		stopChan: make(chan struct{}),
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	s.listener = listener
	log.Printf("Server started on %s", s.address)

	go s.acceptConnections()
	return nil
}

func (s *Server) Stop() {
	close(s.stopChan)
	if s.listener != nil {
		s.listener.Close()
	}
	s.wg.Wait()
	log.Println("Server stopped")
}

func (s *Server) acceptConnections() {
	for {
		select {
		case <-s.stopChan:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-s.stopChan:
					return
				default:
					log.Printf("Failed to accept connection: %v", err)
					continue
				}
			}
			s.wg.Add(1)
			go s.handleConnection(conn)
		}
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	log.Printf("New connection from %s", conn.RemoteAddr())

	for {
		select {
		case <-s.stopChan:
			return
		default:
			msg, err := protocol.DecodeMessage(conn)
			if err != nil {
				if err.Error() != "EOF" {
					log.Printf("Failed to decode message: %v", err)
				}
				return
			}

			response, err := s.processMessage(msg)
			if err != nil {
				errorMsg := &protocol.Message{
					Type:    protocol.TypeError,
					Payload: err.Error(),
				}
				if encodeErr := protocol.EncodeMessage(conn, errorMsg); encodeErr != nil {
					log.Printf("Failed to send error response: %v", encodeErr)
				}
				continue
			}

			if response != nil {
				if encodeErr := protocol.EncodeMessage(conn, response); encodeErr != nil {
					log.Printf("Failed to send response: %v", encodeErr)
				}
			}
		}
	}
}

func (s *Server) processMessage(msg *protocol.Message) (*protocol.Message, error) {
	switch msg.Type {
	case protocol.TypeScanRequest:
		return s.handleScanRequest(msg.Payload)
	default:
		return nil, fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

func (s *Server) handleScanRequest(payload string) (*protocol.Message, error) {
	var req protocol.ScanRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return nil, fmt.Errorf("failed to parse scan request: %w", err)
	}

	log.Printf("Received scan request for directory: %s", req.Directory)

	cfg := s.config
	if req.ConfigFile != "" {
		var err error
		cfg, err = config.LoadConfig(req.ConfigFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	scan := scanner.NewScanner(cfg)
	resp, err := scan.Scan(req.Directory, req.ExcludePaths, req.Severity)
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	responsePayload, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal scan response: %w", err)
	}

	return &protocol.Message{
		Type:    protocol.TypeScanResponse,
		Payload: string(responsePayload),
	}, nil
}
