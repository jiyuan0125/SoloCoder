package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type BackendConfig struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Path    string `json:"path"`
}

type Config struct {
	Backends     []BackendConfig `json:"backends"`
	HealthCheck  struct {
		Interval int `json:"interval"`
		Path     string `json:"path"`
		Timeout  int `json:"timeout"`
	} `json:"health_check"`
	ManagementPath string `json:"management_path"`
}

func LoadConfig(path string) *Config {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Cannot open config file: %v", err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("Cannot read config file: %v", err)
	}

	var cfg Config
	if err := json.Unmarshal(bytes, &cfg); err != nil {
		log.Fatalf("Cannot parse config file: %v", err)
	}

	if cfg.HealthCheck.Interval == 0 {
		cfg.HealthCheck.Interval = 10
	}
	if cfg.HealthCheck.Path == "" {
		cfg.HealthCheck.Path = "/health"
	}
	if cfg.HealthCheck.Timeout == 0 {
		cfg.HealthCheck.Timeout = 5
	}
	if cfg.ManagementPath == "" {
		cfg.ManagementPath = "/_admin"
	}

	return &cfg
}
