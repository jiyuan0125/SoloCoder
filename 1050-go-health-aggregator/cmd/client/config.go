package main

import (
	"os"

	"gopkg.in/yaml.v3"

	"github.com/health-aggregator/pkg/common"
	"github.com/health-aggregator/pkg/core"
)

type ClientConfig struct {
	ServerURL string                  `yaml:"server_url"`
	Services  []*common.ServiceConfig `yaml:"services,omitempty"`
}

func LoadClientConfig(path string) (*ClientConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &ClientConfig{
		ServerURL: "http://localhost:8080",
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	global := core.DefaultGlobalConfig()
	for _, svc := range cfg.Services {
		core.ApplyDefaults(svc, global)
	}

	return cfg, nil
}
