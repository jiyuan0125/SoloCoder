package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"cron-tester/internal/cron"
	"cron-tester/internal/protocol"
)

type Server struct {
	port  int
	cache *cron.Cache
}

func NewServer(port int) *Server {
	return &Server{
		port:  port,
		cache: cron.NewCache(),
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Printf("Cron Tester Server listening on %s\n", addr)

	go s.handleSignals()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Accept error: %v\n", err)
			continue
		}

		go handleConnection(conn, s.cache)
	}
}

func (s *Server) handleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		sig := <-sigChan
		fmt.Printf("\nReceived signal: %s\n", sig)
		fmt.Println("Shutting down server...")
		os.Exit(0)
	}
}

func RunServer() {
	port := flag.Int("port", protocol.DefaultPort, "Port to listen on")
	flag.Parse()

	server := NewServer(*port)
	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
