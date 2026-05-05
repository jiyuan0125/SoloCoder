package server

import (
	"fmt"
	"log"
	"net"
	"sync"

	"slowquery/internal/aggregator"
	"slowquery/internal/parser"
	"slowquery/internal/templater"
	"slowquery/protocol"
)

type Server struct {
	addr       string
	listener   net.Listener
	quit       chan struct{}
	wg         sync.WaitGroup
	aggregator *aggregator.Aggregator
	templater  *templater.SQLTemplater
	mu         sync.RWMutex
	currentReq *protocol.Request
	isRunning  bool
}

func NewServer(addr string) *Server {
	return &Server{
		addr:      addr,
		quit:      make(chan struct{}),
		templater: templater.NewSQLTemplater(),
	}
}

func (s *Server) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", s.addr, err)
	}

	log.Printf("Server listening on %s", s.addr)

	s.wg.Add(1)
	go s.acceptLoop()

	return nil
}

func (s *Server) acceptLoop() {
	defer s.wg.Done()

	for {
		select {
		case <-s.quit:
			return
		default:
		}

		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				log.Printf("Accept error: %v", err)
				continue
			}
		}

		s.wg.Add(1)
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	pconn := protocol.NewConn(conn)

	for {
		select {
		case <-s.quit:
			return
		default:
		}

		req, err := pconn.ReceiveRequest()
		if err != nil {
			return
		}

		resp := s.handleRequest(req)

		if err := pconn.SendResponse(resp); err != nil {
			log.Printf("Failed to send response: %v", err)
			return
		}
	}
}

func (s *Server) handleRequest(req *protocol.Request) *protocol.Response {
	switch req.Command {
	case protocol.CommandTypeStart:
		return s.handleStart(req)
	case protocol.CommandTypeStop:
		return s.handleStop()
	case protocol.CommandTypeStatus:
		return s.handleStatus()
	case protocol.CommandTypeGetStats:
		return s.handleGetStats(req)
	default:
		return &protocol.Response{
			Success: false,
			Error:   fmt.Sprintf("unknown command: %s", req.Command),
		}
	}
}

func (s *Server) handleStart(req *protocol.Request) *protocol.Response {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return &protocol.Response{
			Success: false,
			Error:   "analysis already running",
		}
	}

	if req.LogFilePath == "" {
		return &protocol.Response{
			Success: false,
			Error:   "log file path is required",
		}
	}

	s.currentReq = req
	s.aggregator = aggregator.NewAggregator(req.MinExecTimeMs)
	s.isRunning = true

	go s.runAnalysis(req)

	return &protocol.Response{
		Success: true,
		Status:  "started",
	}
}

func (s *Server) runAnalysis(req *protocol.Request) {
	defer func() {
		s.mu.Lock()
		s.isRunning = false
		s.mu.Unlock()
	}()

	p := parser.NewParser(req.LogFormat)
	results, err := p.ParseFile(req.LogFilePath)
	if err != nil {
		log.Printf("Failed to start parsing: %v", err)
		return
	}

	for result := range results {
		select {
		case <-s.quit:
			return
		default:
		}

		if result.Error != nil {
			log.Printf("Warning: line %d: %v", result.LineNum, result.Error)
			s.aggregator.AddSkipped()
			continue
		}

		if result.Entry == nil {
			continue
		}

		sqlTemplate := s.templater.TemplateSQL(result.Entry.SQL)
		s.aggregator.Add(result.Entry, sqlTemplate)
	}
}

func (s *Server) handleStop() *protocol.Response {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return &protocol.Response{
			Success: false,
			Error:   "no analysis running",
		}
	}

	select {
	case <-s.quit:
	default:
		close(s.quit)
	}

	s.isRunning = false

	return &protocol.Response{
		Success: true,
		Status:  "stopped",
	}
}

func (s *Server) handleStatus() *protocol.Response {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := "idle"
	if s.isRunning {
		status = "running"
	}

	return &protocol.Response{
		Success: true,
		Status:  status,
	}
}

func (s *Server) handleGetStats(req *protocol.Request) *protocol.Response {
	s.mu.RLock()
	agg := s.aggregator
	s.mu.RUnlock()

	if agg == nil {
		return &protocol.Response{
			Success: false,
			Error:   "no analysis has been run",
		}
	}

	sortBy := protocol.SortByAvgExecTime
	if req.SortBy == protocol.SortByTotalExecTime {
		sortBy = protocol.SortByTotalExecTime
	}

	topN := 20
	if req.TopN > 0 {
		topN = req.TopN
	}

	stats := agg.GetSortedStats(sortBy, topN)

	return &protocol.Response{
		Success: true,
		Stats:   stats,
	}
}

func (s *Server) Stop() {
	select {
	case <-s.quit:
	default:
		close(s.quit)
	}

	if s.listener != nil {
		s.listener.Close()
	}

	s.wg.Wait()
	log.Println("Server stopped")
}
