package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"ssh-tunnel/pkg/common"
	"ssh-tunnel/pkg/tunnel"
)

type Server struct {
	manager *tunnel.Manager
	logger  *log.Logger
}

func NewServer() *Server {
	return &Server{
		manager: tunnel.NewManager(),
		logger:  log.New(os.Stdout, "[server] ", log.LstdFlags),
	}
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (s *Server) writeError(w http.ResponseWriter, status int, err string) {
	s.writeJSON(w, status, &common.ErrorResponse{
		Success: false,
		Error:   err,
	})
}

func (s *Server) handleCreateTunnel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.CreateTunnelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ID == "" {
		s.writeError(w, http.StatusBadRequest, "tunnel ID is required")
		return
	}

	tun, err := s.manager.CreateTunnel(&req.TunnelConfig)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	state := tun.GetState()
	s.writeJSON(w, http.StatusCreated, &common.CreateTunnelResponse{
		Success: true,
		Tunnel:  &state,
	})
}

func (s *Server) handleStopTunnel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.StopTunnelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.manager.StopTunnel(req.ID); err != nil {
		s.writeError(w, http.StatusNotFound, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, &common.StopTunnelResponse{
		Success: true,
		Message: "tunnel stopped",
	})
}

func (s *Server) handleListTunnels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	tunnels := s.manager.ListTunnels()
	states := make([]*common.TunnelState, len(tunnels))
	for i := range tunnels {
		states[i] = &tunnels[i]
	}

	s.writeJSON(w, http.StatusOK, &common.ListTunnelsResponse{
		Success: true,
		Tunnels: states,
	})
}

func (s *Server) handleGetTunnel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		s.writeError(w, http.StatusBadRequest, "tunnel ID is required")
		return
	}

	tun, err := s.manager.GetTunnel(id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err.Error())
		return
	}

	state := tun.GetState()
	s.writeJSON(w, http.StatusOK, &common.GetTunnelResponse{
		Success: true,
		Tunnel:  &state,
	})
}

func (s *Server) run(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/tunnels/create", s.handleCreateTunnel)
	mux.HandleFunc("/tunnels/stop", s.handleStopTunnel)
	mux.HandleFunc("/tunnels/list", s.handleListTunnels)
	mux.HandleFunc("/tunnels/get", s.handleGetTunnel)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		s.logger.Println("收到停止信号，正在关闭服务...")
		s.manager.StopAll()
		_ = server.Shutdown(nil)
	}()

	s.logger.Printf("服务端监听: %s", addr)
	return server.ListenAndServe()
}

func main() {
	addr := flag.String("addr", ":8080", "服务端监听地址")
	flag.Parse()

	server := NewServer()
	if err := server.run(*addr); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务端运行失败: %v", err)
	}
}
