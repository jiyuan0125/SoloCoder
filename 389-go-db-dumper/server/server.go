package server

import (
	"log"
	"net"
	"sync"

	"go-db-dumper/proto"
)

type Server struct {
	config    *proto.ConverterConfig
	converter *Converter
	listener  net.Listener
	running   bool
	mu        sync.Mutex
	wg        sync.WaitGroup
}

func NewServer(config *proto.ConverterConfig) *Server {
	if config == nil {
		config = proto.DefaultConverterConfig()
	}

	return &Server{
		config:    config,
		converter: NewConverter(config),
	}
}

func (s *Server) Start(addr string) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.listener = listener
	s.running = true
	s.mu.Unlock()

	log.Printf("Server started on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.mu.Lock()
			running := s.running
			s.mu.Unlock()

			if !running {
				break
			}

			log.Printf("Error accepting connection: %v", err)
			continue
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}

	return nil
}

func (s *Server) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}

	s.running = false
	if s.listener != nil {
		s.listener.Close()
	}
	s.mu.Unlock()

	s.wg.Wait()
	log.Println("Server stopped")
}

func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	log.Printf("New connection from %s", conn.RemoteAddr())

	msg, err := proto.ReadMessage(conn)
	if err != nil {
		log.Printf("Error reading message: %v", err)
		s.sendErrorResponse(conn, "Failed to read request")
		return
	}

	if msg.Type != proto.MessageTypeRequest || msg.Request == nil {
		log.Printf("Invalid message type or missing request")
		s.sendErrorResponse(conn, "Invalid request")
		return
	}

	req := msg.Request
	log.Printf("Received convert request: input=%s, output=%s, table=%s",
		req.InputFile, req.OutputFile, req.TableName)

	result, err := s.converter.Convert(req.InputFile, req.OutputFile, req.TableName)
	if err != nil {
		log.Printf("Conversion error: %v", err)
		s.sendErrorResponse(conn, err.Error())
		return
	}

	log.Printf("Conversion successful: records=%d, sql=%d", result.RecordsRead, result.SQLCount)

	response := &proto.Message{
		Type: proto.MessageTypeResponse,
		Response: &proto.ConvertResponse{
			Status:      proto.ResponseStatusSuccess,
			Message:     "Conversion completed successfully",
			RecordsRead: result.RecordsRead,
			SQLCount:    result.SQLCount,
		},
	}

	err = proto.WriteMessage(conn, response)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func (s *Server) sendErrorResponse(conn net.Conn, message string) {
	response := &proto.Message{
		Type: proto.MessageTypeResponse,
		Response: &proto.ConvertResponse{
			Status:  proto.ResponseStatusError,
			Message: message,
		},
	}

	proto.WriteMessage(conn, response)
}
