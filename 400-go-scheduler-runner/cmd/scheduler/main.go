package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"scheduler/internal/httpapi"
	"scheduler/internal/protocol"
	"scheduler/internal/scheduler"
	"scheduler/internal/tcp"
)

const (
	defaultTCPAddr      = "localhost:8081"
	defaultHTTPAddr     = "localhost:8080"
	defaultConfigFile   = "tasks.json"
	defaultMaxRecords   = 100
	defaultShutdownWait = 30
)

func main() {
	configFile := flag.String("config", "", "Server config file path")
	flag.Parse()

	serverConfig := loadServerConfig(*configFile)

	log.Printf("Starting scheduler service...")
	log.Printf("TCP address: %s", serverConfig.TCPAddr)
	log.Printf("HTTP address: %s", serverConfig.HTTPAddr)
	log.Printf("Tasks config file: %s", serverConfig.ConfigFile)

	sched := scheduler.NewScheduler()
	sched.SetConfigFile(serverConfig.ConfigFile)
	sched.SetMaxRecords(serverConfig.MaxRecords)

	if err := sched.LoadConfig(); err != nil {
		log.Printf("Warning: failed to load tasks config: %v", err)
	}

	tasks := sched.ListTasks()
	log.Printf("Loaded %d tasks", len(tasks))

	for _, task := range tasks {
		if task.Config.Disabled {
			log.Printf("Task '%s' is disabled (invalid cron expression)", task.Config.Name)
		}
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(1)
	go func() {
		defer wg.Done()
		sched.Start(ctx)
	}()

	httpServer := httpapi.NewHTTPServer(serverConfig.HTTPAddr, sched)
	if err := httpServer.Start(); err != nil {
		log.Printf("Failed to start HTTP server: %v", err)
		cancel()
		os.Exit(1)
	}
	log.Printf("HTTP server listening on %s", serverConfig.HTTPAddr)

	tcpServer := tcp.NewTCPServer(serverConfig.TCPAddr, sched)
	if err := tcpServer.Start(); err != nil {
		log.Printf("Failed to start TCP server: %v", err)
		cancel()
		os.Exit(1)
	}
	log.Printf("TCP server listening on %s", serverConfig.TCPAddr)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal: %s, shutting down...", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		time.Duration(serverConfig.ShutdownWait)*time.Second,
	)
	defer shutdownCancel()

	if err := httpServer.Stop(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	tcpStopCtx, tcpStopCancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer tcpStopCancel()
	if err := tcpServer.Stop(tcpStopCtx); err != nil {
		log.Printf("TCP server shutdown error: %v", err)
	}

	cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-shutdownCtx.Done():
		log.Printf("Shutdown timeout reached, forcing exit")
	case <-done:
		log.Printf("All tasks completed, exiting gracefully")
	}
}

func loadServerConfig(path string) protocol.ServerConfig {
	config := protocol.ServerConfig{
		TCPAddr:      defaultTCPAddr,
		HTTPAddr:     defaultHTTPAddr,
		ConfigFile:   defaultConfigFile,
		MaxRecords:   defaultMaxRecords,
		ShutdownWait: defaultShutdownWait,
	}

	if path == "" {
		return config
	}

	file, err := os.Open(path)
	if err != nil {
		log.Printf("Warning: failed to open server config file %s: %v", path, err)
		return config
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Warning: failed to read server config file: %v", err)
		return config
	}

	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf("Warning: failed to parse server config file: %v", err)
		return config
	}

	return config
}
