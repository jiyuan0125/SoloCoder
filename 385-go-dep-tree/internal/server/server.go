package server

import (
	"log"
	"net"
	"sync"
	
	"dep-tree/internal/analyzer"
	"dep-tree/protocol"
)

type Server struct {
	listener net.Listener
	stop     chan struct{}
	wg       sync.WaitGroup
}

func New() *Server {
	return &Server{
		stop: make(chan struct{}),
	}
}

func (s *Server) Start(addr string) error {
	var err error
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	
	log.Printf("Server listening on %s", addr)
	
	for {
		select {
		case <-s.stop:
			return nil
		default:
		}
		
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.stop:
				return nil
			default:
				log.Printf("Accept error: %v", err)
				continue
			}
		}
		
		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

func (s *Server) Stop() {
	close(s.stop)
	if s.listener != nil {
		s.listener.Close()
	}
	s.wg.Wait()
}

func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()
	
	requestData, err := protocol.ReadMessage(conn)
	if err != nil {
		log.Printf("Failed to read request: %v", err)
		return
	}
	
	req, err := protocol.DecodeRequest(requestData)
	if err != nil {
		log.Printf("Failed to decode request: %v", err)
		sendErrorResponse(conn, err.Error())
		return
	}
	
	log.Printf("Processing request for: %s", req.ProjectPath)
	
	project, err := analyzer.Analyze(req.ProjectPath, req.FilterOpts)
	if err != nil {
		log.Printf("Analysis failed: %v", err)
		sendErrorResponse(conn, err.Error())
		return
	}
	
	resp := &protocol.Response{
		Success: true,
		Project: project,
	}
	
	responseData, err := protocol.EncodeResponse(resp)
	if err != nil {
		log.Printf("Failed to encode response: %v", err)
		return
	}
	
	err = protocol.WriteMessage(conn, responseData)
	if err != nil {
		log.Printf("Failed to send response: %v", err)
	}
}

func sendErrorResponse(conn net.Conn, errMsg string) {
	resp := &protocol.Response{
		Success:  false,
		ErrorMsg: errMsg,
	}
	
	data, _ := protocol.EncodeResponse(resp)
	protocol.WriteMessage(conn, data)
}
