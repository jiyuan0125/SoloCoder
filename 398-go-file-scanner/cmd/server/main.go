package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"filescanner/internal/config"
	"filescanner/internal/server"
)

func main() {
	var (
		address    string
		configFile string
	)

	flag.StringVar(&address, "address", "localhost:8080", "Server address to listen on")
	flag.StringVar(&configFile, "config", "", "Path to configuration file")
	flag.Parse()

	fmt.Println("Starting file scanner server...")
	fmt.Printf("Listening on: %s\n", address)

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Printf("Loaded %d scan rules\n", len(cfg.Rules))

	srv := server.NewServer(address, cfg)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// 处理系统信号，实现优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Server is running. Press Ctrl+C to stop.")

	// 等待信号
	<-sigChan
	fmt.Println("\nReceived shutdown signal. Stopping server...")

	// 优雅停止服务器
	srv.Stop()

	fmt.Println("Server stopped successfully.")
}
