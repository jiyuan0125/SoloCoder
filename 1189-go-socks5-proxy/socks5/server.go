package socks5

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	AuthStore AuthStore
	RequireAuth bool

	listener   net.Listener
	running    bool
	mu         sync.Mutex

	connections  map[string]*Connection
	associations map[string]*UDPAssociation
	connMu       sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		AuthStore:    NewUserAuth(),
		RequireAuth:  false,
		connections:  make(map[string]*Connection),
		associations: make(map[string]*UDPAssociation),
	}
}

func (s *Server) ListenAndServe(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.listener = listener
	s.running = true
	s.mu.Unlock()

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.mu.Lock()
			running := s.running
			s.mu.Unlock()
			if !running {
				return nil
			}
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) Close() error {
	s.mu.Lock()
	s.running = false
	if s.listener != nil {
		err := s.listener.Close()
		s.listener = nil
		s.mu.Unlock()
		return err
	}
	s.mu.Unlock()
	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	authenticated, err := handleGreeting(conn, s.AuthStore, s.RequireAuth)
	if err != nil {
		conn.Close()
		return
	}
	if !authenticated {
		conn.Close()
		return
	}

	req, err := readRequest(conn)
	if err != nil {
		conn.Close()
		return
	}

	switch req.Cmd {
	case CmdConnect:
		connObj, err := handleConnect(conn, req.Addr)
		if err != nil {
			sendErrorReply(conn, err)
			conn.Close()
			return
		}
		s.addConnection(connObj)

	case CmdUDPAssociate:
		assoc, err := handleUDPAssociate(conn, req.Addr)
		if err != nil {
			sendErrorReply(conn, err)
			conn.Close()
			return
		}
		s.addAssociation(assoc)

	default:
		sendCmdNotSupported(conn)
		conn.Close()
	}
}

type Request struct {
	Cmd  byte
	Addr *Address
}

func readRequest(conn net.Conn) (*Request, error) {
	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}

	if header[0] != Version {
		return nil, fmt.Errorf("unsupported SOCKS version: %d", header[0])
	}

	cmd := header[1]
	addrType := header[3]

	addr, err := ReadAddressWithType(conn, addrType)
	if err != nil {
		return nil, err
	}

	return &Request{
		Cmd:  cmd,
		Addr: addr,
	}, nil
}

func sendErrorReply(conn net.Conn, err error) {
	reply := buildReply(ReplyFailure, &Address{Type: AddrTypeIPv4, IP: net.IPv4zero, Port: 0})
	conn.Write(reply)
}

func sendCmdNotSupported(conn net.Conn) {
	reply := buildReply(ReplyCmdNotSupported, &Address{Type: AddrTypeIPv4, IP: net.IPv4zero, Port: 0})
	conn.Write(reply)
}

func (s *Server) addConnection(c *Connection) {
	s.connMu.Lock()
	defer s.connMu.Unlock()
	s.connections[c.conn.RemoteAddr().String()] = c
}

func (s *Server) addAssociation(a *UDPAssociation) {
	s.connMu.Lock()
	defer s.connMu.Unlock()
	s.associations[a.ClientAddr.String()] = a
}

func (s *Server) GetConnections() []*Connection {
	s.connMu.RLock()
	defer s.connMu.RUnlock()
	conns := make([]*Connection, 0, len(s.connections))
	for _, c := range s.connections {
		conns = append(conns, c)
	}
	return conns
}

func (s *Server) GetAssociations() []*UDPAssociation {
	s.connMu.RLock()
	defer s.connMu.RUnlock()
	assocs := make([]*UDPAssociation, 0, len(s.associations))
	for _, a := range s.associations {
		assocs = append(assocs, a)
	}
	return assocs
}

func (s *Server) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running && s.listener != nil
}

func (s *Server) Addr() net.Addr {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener != nil {
		return s.listener.Addr()
	}
	return nil
}
