package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"env-switcher/protocol"
)

const (
	defaultPort = "45678"
	portEnvVar  = "ENV_SWITCHER_PORT"
)

func getPort() string {
	if port := os.Getenv(portEnvVar); port != "" {
		return port
	}
	return defaultPort
}

func handleConnection(conn net.Conn, em *EnvManager) {
	defer conn.Close()
	
	reader := bufio.NewReader(conn)
	data, err := reader.ReadBytes('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request: %v\n", err)
		return
	}
	
	req, err := protocol.DecodeRequest(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding request: %v\n", err)
		resp := &protocol.Response{
			Success: false,
			Message: "Invalid request format",
		}
		respData, _ := protocol.EncodeResponse(resp)
		conn.Write(respData)
		return
	}
	
	resp := em.HandleRequest(req)
	
	respData, err := protocol.EncodeResponse(resp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding response: %v\n", err)
		return
	}
	
	conn.Write(respData)
}

func main() {
	em, err := NewEnvManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize environment manager: %v\n", err)
		os.Exit(1)
	}
	
	port := getPort()
	listener, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()
	
	fmt.Printf("env-switcher server listening on localhost:%s\n", port)
	
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			go handleConnection(conn, em)
		}
	}()
	
	<-sigChan
	fmt.Println("\nServer shutting down...")
}
