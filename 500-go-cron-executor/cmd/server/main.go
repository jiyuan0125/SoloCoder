package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cron-executor/pkg/server"
	"cron-executor/pkg/tcpserver"
	"cron-executor/pkg/webapi"
)

func main() {
	configPath := flag.String("config", server.DefaultConfigPath(), "Path to config file")
	flag.Parse()

	cfg, err := server.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	srv := server.NewServer(cfg)

	if err := srv.LoadTasksFromConfig(); err != nil {
		log.Fatalf("Failed to load tasks: %v", err)
	}

	if err := srv.LoadPersistedData(); err != nil {
		log.Printf("Warning: failed to load persisted data: %v", err)
	}

	tcpSrv := tcpserver.NewTCPServer(cfg.TCPPort, srv)
	if err := tcpSrv.Start(); err != nil {
		log.Fatalf("Failed to start TCP server: %v", err)
	}

	webAPI := webapi.NewWebAPI(cfg.WebAPIPort, srv)
	if err := webAPI.Start(); err != nil {
		log.Fatalf("Failed to start Web API: %v", err)
	}

	srv.Start()

	fmt.Printf("Cron Executor Server started\n")
	fmt.Printf("TCP Port: %d\n", cfg.TCPPort)
	fmt.Printf("Web API Port: %d\n", cfg.WebAPIPort)
	fmt.Printf("Log Directory: %s\n", cfg.LogDir)
	fmt.Printf("Data Directory: %s\n", cfg.DataDir)
	fmt.Printf("Loaded %d tasks\n", len(srv.ListTasks()))

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	fmt.Println("\nShutting down...")

	webAPI.Stop()
	tcpSrv.Stop()
	srv.Stop()

	fmt.Println("Server stopped")
}
