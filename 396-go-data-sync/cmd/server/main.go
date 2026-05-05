package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go-data-sync/pkg/protocol"
)

var (
	port     = flag.String("port", "8888", "Server port")
	addr     = flag.String("addr", "127.0.0.1", "Server address")
)

func main() {
	flag.Parse()
	
	listenAddr := fmt.Sprintf("%s:%s", *addr, *port)
	
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()
	
	log.Printf("Server started on %s", listenAddr)
	
	go handleSignals(listener)
	
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		
		go handleConnection(conn)
	}
}

func handleSignals(listener net.Listener) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	<-sigChan
	log.Println("Shutting down server...")
	listener.Close()
	os.Exit(0)
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	
	var req protocol.Request
	if err := protocol.Decode(reader, &req); err != nil {
		log.Printf("Failed to decode request: %v", err)
		sendErrorResponse(writer, fmt.Sprintf("Failed to decode request: %v", err))
		return
	}
	
	log.Printf("Received command: %s", req.Command)
	
	var resp *protocol.Response
	var err error
	
	switch req.Command {
	case protocol.CmdSync:
		resp, err = handleSync(&req)
	case protocol.CmdStatus:
		resp, err = handleStatus(&req)
	case protocol.CmdHistory:
		resp, err = handleHistory(&req)
	default:
		resp = &protocol.Response{
			Success: false,
			Message: fmt.Sprintf("Unknown command: %s", req.Command),
		}
	}
	
	if err != nil {
		sendErrorResponse(writer, err.Error())
		return
	}
	
	if err := protocol.Encode(writer, resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
		return
	}
	
	if err := writer.Flush(); err != nil {
		log.Printf("Failed to flush response: %v", err)
	}
}

func sendErrorResponse(writer *bufio.Writer, message string) {
	resp := &protocol.Response{
		Success: false,
		Message: message,
	}
	protocol.Encode(writer, resp)
	writer.Flush()
}
