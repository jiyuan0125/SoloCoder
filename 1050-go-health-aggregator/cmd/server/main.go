package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"gopkg.in/yaml.v3"

	"github.com/health-aggregator/pkg/common"
	"github.com/health-aggregator/pkg/core"
)

type Config struct {
	ServerAddr   string                  `yaml:"server_addr"`
	Global       *core.GlobalConfig      `yaml:"global,omitempty"`
	Services     []*common.ServiceConfig `yaml:"services,omitempty"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		ServerAddr: ":8080",
		Global:     core.DefaultGlobalConfig(),
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.Global == nil {
		cfg.Global = core.DefaultGlobalConfig()
	}
	return cfg, nil
}

func main() {
	configPath := "server.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	var cfg *Config
	var err error
	if _, statErr := os.Stat(configPath); statErr == nil {
		cfg, err = LoadConfig(configPath)
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}
	} else {
		cfg = &Config{
			ServerAddr: ":8080",
			Global:     core.DefaultGlobalConfig(),
		}
	}

	store := core.NewStateStore(cfg.Global)
	for _, svc := range cfg.Services {
		store.AddService(svc)
	}

	scheduler := core.NewScheduler(store)
	scheduler.Start()
	defer scheduler.Stop()

	apiServer := NewAPIServer(store, scheduler)

	httpServer := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: apiServer,
	}

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		fmt.Printf("Health Aggregator Server starting on %s...\n", cfg.ServerAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-stopCh
	fmt.Println("\nShutting down...")
}
