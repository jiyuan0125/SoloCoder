package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"socks5-proxy/common"
	"socks5-proxy/socks5"
)

type Server struct {
	socks5Server *socks5.Server
	mu             sync.Mutex
}

func NewServer() *Server {
	return &Server{
		socks5Server: socks5.NewServer(),
	}
}

func (s *Server) handleStartProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.StartProxyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: "Invalid request body"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.socks5Server.IsRunning() {
		writeJSON(w, http.StatusConflict, common.ErrorResponse{Error: "Proxy already running"})
		return
	}

	s.socks5Server = socks5.NewServer()
	s.socks5Server.RequireAuth = req.RequireAuth

	for _, u := range req.Users {
		if err := s.socks5Server.AuthStore.AddUser(u.Username, u.Password); err != nil {
			writeJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: err.Error()})
			return
		}
	}

	addr := fmt.Sprintf(":%d", req.ListenPort)

	go func() {
		if err := s.socks5Server.ListenAndServe(addr); err != nil {
			log.Printf("SOCKS5 server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	if !s.socks5Server.IsRunning() {
		writeJSON(w, http.StatusInternalServerError, common.ErrorResponse{Error: "Failed to start proxy"})
		return
	}

	writeJSON(w, http.StatusOK, common.StartProxyResponse{
		Success: true,
		Message: "Proxy started successfully",
		Address: s.socks5Server.Addr().String(),
	})
}

func (s *Server) handleProxyStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	running := s.socks5Server.IsRunning()
	addr := ""
	if running {
		addr = s.socks5Server.Addr().String()
	}
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, common.ProxyStatusResponse{
		Running: running,
		Address: addr,
	})
}

func (s *Server) handleActiveConnections(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.socks5Server.IsRunning() {
		writeJSON(w, http.StatusServiceUnavailable, common.ErrorResponse{Error: "Proxy not running"})
		return
	}

	conns := s.socks5Server.GetConnections()
	assocs := s.socks5Server.GetAssociations()

	tcpConns := make([]common.ConnectionInfo, 0, len(conns))
	for _, c := range conns {
		tcpConns = append(tcpConns, common.ConnectionInfo{
			ClientAddr: c.ClientAddr.String(),
			TargetAddr: c.TargetAddr,
			StartTime:  c.StartTime.Format(time.RFC3339),
		})
	}

	udpAssocs := make([]common.UDPAssociationInfo, 0, len(assocs))
	for _, a := range assocs {
		udpAssocs = append(udpAssocs, common.UDPAssociationInfo{
			ClientAddr: a.ClientAddr.String(),
			RelayAddr:  a.RelayAddr.String(),
			StartTime:  a.StartTime.Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, common.ActiveConnectionsResponse{
		TCPConnections:  tcpConns,
		UDPAssociations: udpAssocs,
	})
}

func (s *Server) handleAddUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.AddUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: "Invalid request body"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.socks5Server.AuthStore.AddUser(req.Username, req.Password); err != nil {
		writeJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, common.UserResponse{
		Success: true,
		Message: fmt.Sprintf("User %s added successfully", req.Username),
	})
}

func (s *Server) handleRemoveUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.RemoveUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: "Invalid request body"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.socks5Server.AuthStore.RemoveUser(req.Username); err != nil {
		writeJSON(w, http.StatusNotFound, common.ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, common.UserResponse{
		Success: true,
		Message: fmt.Sprintf("User %s removed successfully", req.Username),
	})
}

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	users := s.socks5Server.AuthStore.ListUsers()
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, common.ListUsersResponse{
		Users: users,
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func main() {
	server := NewServer()

	http.HandleFunc("/api/proxy/start", server.handleStartProxy)
	http.HandleFunc("/api/proxy/status", server.handleProxyStatus)
	http.HandleFunc("/api/proxy/connections", server.handleActiveConnections)
	http.HandleFunc("/api/users/add", server.handleAddUser)
	http.HandleFunc("/api/users/remove", server.handleRemoveUser)
	http.HandleFunc("/api/users/list", server.handleListUsers)

	port := getEnvOrDefault("HTTP_PORT", "8210")
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	_, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("Invalid port: %v", err)
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
