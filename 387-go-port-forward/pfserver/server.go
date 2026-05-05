package main

import (
	"fmt"
	"net"
	"sync"
	"time"
	
	"port-forward/protocol"
)

type Server struct {
	controlPort int
	verbose     bool
	listeners   map[int]*net.Listener
	forwards    map[int]*Forwarder
	stats       map[int]*ForwardStats
	connID      uint64
	mu          sync.RWMutex
	shutdownCh  chan struct{}
	wg          sync.WaitGroup
}

type ForwardStats struct {
	rule        protocol.ForwardRule
	activeConns int
	totalConns  int64
	totalBytes  int64
	connections map[uint64]*ConnStats
	mu          sync.RWMutex
}

type ConnStats struct {
	id           uint64
	clientAddr   string
	connectedAt  time.Time
	bytesSent    int64
	bytesReceived int64
	isActive     bool
	mu           sync.RWMutex
}

func NewServer(controlPort int, verbose bool) *Server {
	return &Server{
		controlPort: controlPort,
		verbose:     verbose,
		listeners:   make(map[int]*net.Listener),
		forwards:    make(map[int]*Forwarder),
		stats:       make(map[int]*ForwardStats),
		shutdownCh:  make(chan struct{}),
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("127.0.0.1:%d", s.controlPort)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on control port %d: %v", s.controlPort, err)
	}
	
	logf("Server listening on control port %d", s.controlPort)
	
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.shutdownCh:
				return
			default:
			}
			
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-s.shutdownCh:
					return
				default:
					logf("Accept error: %v", err)
					continue
				}
			}
			
			s.wg.Add(1)
			go s.handleClient(conn)
		}
	}()
	
	return nil
}

func (s *Server) handleClient(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()
	
	for {
		select {
		case <-s.shutdownCh:
			return
		default:
		}
		
		msg, err := protocol.DecodeMessage(conn)
		if err != nil {
			return
		}
		
		s.handleMessage(conn, msg)
	}
}

func (s *Server) handleMessage(conn net.Conn, msg protocol.Message) {
	switch msg.Type {
	case protocol.MsgTypeAddForward:
		s.handleAddForward(conn, msg.Payload)
	case protocol.MsgTypeRemoveForward:
		s.handleRemoveForward(conn, msg.Payload)
	case protocol.MsgTypeListForwards:
		s.handleListForwards(conn)
	case protocol.MsgTypeGetStats:
		s.handleGetStats(conn, msg.Payload)
	case protocol.MsgTypeShutdown:
		s.handleShutdown(conn)
	default:
		protocol.SendResponse(conn, false, "Unknown message type", nil)
	}
}

func (s *Server) handleAddForward(conn net.Conn, payload []byte) {
	var req protocol.AddForwardRequest
	if err := protocol.DecodeJSON(payload, &req); err != nil {
		protocol.SendResponse(conn, false, "Invalid request: "+err.Error(), nil)
		return
	}
	
	if req.LocalPort <= 0 || req.LocalPort > 65535 {
		protocol.SendResponse(conn, false, "Invalid local port", nil)
		return
	}
	
	if req.RemoteAddr == "" {
		protocol.SendResponse(conn, false, "Remote address required", nil)
		return
	}
	
	if req.Timeout <= 0 {
		req.Timeout = protocol.DefaultTimeout
	}
	
	if req.BufferSize <= 0 {
		req.BufferSize = protocol.DefaultBufferSize
	}
	
	s.mu.Lock()
	if _, exists := s.forwards[req.LocalPort]; exists {
		s.mu.Unlock()
		protocol.SendResponse(conn, false, fmt.Sprintf("Port %d already forwarded", req.LocalPort), nil)
		return
	}
	
	forwarder := NewForwarder(
		req.LocalPort,
		req.RemoteAddr,
		time.Duration(req.Timeout)*time.Second,
		req.BufferSize,
		req.Verbose || s.verbose,
	)
	
	if err := forwarder.Start(); err != nil {
		s.mu.Unlock()
		protocol.SendResponse(conn, false, fmt.Sprintf("Failed to start forwarder: %v", err), nil)
		return
	}
	
	stats := &ForwardStats{
		rule: protocol.ForwardRule{
			LocalPort:  req.LocalPort,
			RemoteAddr: req.RemoteAddr,
		},
		connections: make(map[uint64]*ConnStats),
	}
	
	s.forwards[req.LocalPort] = forwarder
	s.stats[req.LocalPort] = stats
	s.mu.Unlock()
	
	go s.monitorForwarder(req.LocalPort, forwarder)
	
	logf("Added forward: local port %d -> %s", req.LocalPort, req.RemoteAddr)
	protocol.SendResponse(conn, true, fmt.Sprintf("Forward added: %d -> %s", req.LocalPort, req.RemoteAddr), nil)
}

func (s *Server) monitorForwarder(localPort int, forwarder *Forwarder) {
	for {
		select {
		case connInfo, ok := <-forwarder.NewConnCh:
			if !ok {
				return
			}
			s.handleNewConnection(localPort, connInfo)
		case connStats, ok := <-forwarder.ConnCloseCh:
			if !ok {
				return
			}
			s.handleConnectionClose(localPort, connStats)
		case <-s.shutdownCh:
			return
		}
	}
}

func (s *Server) handleNewConnection(localPort int, connInfo *ConnInfo) {
	s.mu.Lock()
	connID := s.connID
	s.connID++
	s.mu.Unlock()
	
	s.mu.RLock()
	stats := s.stats[localPort]
	s.mu.RUnlock()
	
	if stats == nil {
		return
	}
	
	stats.mu.Lock()
	connStats := &ConnStats{
		id:          connID,
		clientAddr:  connInfo.ClientAddr,
		connectedAt: time.Now(),
		isActive:    true,
	}
	stats.connections[connID] = connStats
	stats.activeConns++
	stats.totalConns++
	stats.mu.Unlock()
	
	connInfo.ID = connID
	close(connInfo.IDReady)
	
	logf("New connection on port %d: %s -> %s", localPort, connInfo.ClientAddr, connInfo.RemoteAddr)
}

func (s *Server) handleConnectionClose(localPort int, connStats *ConnCloseInfo) {
	s.mu.RLock()
	stats := s.stats[localPort]
	s.mu.RUnlock()
	
	if stats == nil {
		return
	}
	
	stats.mu.Lock()
	if c, exists := stats.connections[connStats.ID]; exists {
		c.mu.Lock()
		c.bytesSent = connStats.BytesSent
		c.bytesReceived = connStats.BytesReceived
		c.isActive = false
		c.mu.Unlock()
	}
	stats.activeConns--
	stats.totalBytes += connStats.BytesSent + connStats.BytesReceived
	stats.mu.Unlock()
	
	logf("Connection closed on port %d: sent %d bytes, received %d bytes, duration: %v",
		localPort, connStats.BytesSent, connStats.BytesReceived, connStats.Duration)
}

func (s *Server) handleRemoveForward(conn net.Conn, payload []byte) {
	var req protocol.RemoveForwardRequest
	if err := protocol.DecodeJSON(payload, &req); err != nil {
		protocol.SendResponse(conn, false, "Invalid request: "+err.Error(), nil)
		return
	}
	
	s.mu.Lock()
	forwarder, exists := s.forwards[req.LocalPort]
	if exists {
		delete(s.forwards, req.LocalPort)
		delete(s.stats, req.LocalPort)
	}
	s.mu.Unlock()
	
	if !exists {
		protocol.SendResponse(conn, false, fmt.Sprintf("No forward on port %d", req.LocalPort), nil)
		return
	}
	
	forwarder.Stop()
	logf("Removed forward on port %d", req.LocalPort)
	protocol.SendResponse(conn, true, fmt.Sprintf("Forward removed: port %d", req.LocalPort), nil)
}

func (s *Server) handleListForwards(conn net.Conn) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	rules := make([]protocol.ForwardRule, 0, len(s.forwards))
	for port := range s.forwards {
		if stats, ok := s.stats[port]; ok {
			rules = append(rules, stats.rule)
		}
	}
	
	protocol.SendResponse(conn, true, "", rules)
}

func (s *Server) handleGetStats(conn net.Conn, payload []byte) {
	var localPort int
	if len(payload) > 0 {
		protocol.DecodeJSON(payload, &localPort)
	}
	
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var statsList []protocol.ForwardStats
	
	if localPort > 0 {
		if stats, ok := s.stats[localPort]; ok {
			statsList = append(statsList, s.buildForwardStats(stats))
		}
	} else {
		for _, stats := range s.stats {
			statsList = append(statsList, s.buildForwardStats(stats))
		}
	}
	
	protocol.SendResponse(conn, true, "", statsList)
}

func (s *Server) buildForwardStats(stats *ForwardStats) protocol.ForwardStats {
	stats.mu.RLock()
	defer stats.mu.RUnlock()
	
	conns := make([]protocol.ConnectionStats, 0, len(stats.connections))
	for _, c := range stats.connections {
		c.mu.RLock()
		conns = append(conns, protocol.ConnectionStats{
			ID:            c.id,
			LocalPort:     stats.rule.LocalPort,
			RemoteAddr:    stats.rule.RemoteAddr,
			ClientAddr:    c.clientAddr,
			ConnectedAt:   c.connectedAt.Unix(),
			BytesSent:     c.bytesSent,
			BytesReceived: c.bytesReceived,
			IsActive:      c.isActive,
		})
		c.mu.RUnlock()
	}
	
	return protocol.ForwardStats{
		Rule:        stats.rule,
		ActiveConns: stats.activeConns,
		TotalConns:  stats.totalConns,
		TotalBytes:  stats.totalBytes,
		Connections: conns,
	}
}

func (s *Server) handleShutdown(conn net.Conn) {
	protocol.SendResponse(conn, true, "Shutting down server...", nil)
	go s.Shutdown()
}

func (s *Server) Shutdown() {
	select {
	case <-s.shutdownCh:
		return
	default:
		close(s.shutdownCh)
	}
	
	s.mu.Lock()
	for port, forwarder := range s.forwards {
		forwarder.Stop()
		delete(s.forwards, port)
	}
	s.mu.Unlock()
	
	s.wg.Wait()
	logf("Server shutdown complete")
}
