package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

type BackendConfig struct {
	Addr   string `json:"addr"`
	Weight int    `json:"weight"`
}

type RouteConfig struct {
	PathPrefix string          `json:"path_prefix"`
	Backends    []BackendConfig `json:"backends"`
}

type Config struct {
	ListenAddr string         `json:"listen_addr"`
	Timeout     int            `json:"timeout"`
	Routes      []RouteConfig `json:"routes"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config.Timeout <= 0 {
		config.Timeout = 30
	}
	if config.ListenAddr == "" {
		config.ListenAddr = ":8080"
	}

	return &config, nil
}

func main() {
	configPath := flag.String("config", "config.json", "Path to config file")
	flag.Parse()

	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	proxy, err := NewProxy(config)
	if err != nil {
		log.Fatalf("Failed to create proxy: %v", err)
	}

	log.Printf("Starting proxy on %s", config.ListenAddr)
	if err := proxy.Start(); err != nil {
		log.Fatalf("Failed to start proxy: %v", err)
	}
}
