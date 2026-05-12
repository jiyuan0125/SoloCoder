package config

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	Port int
}

func Load() *Config {
	cfg := &Config{
		Port: 8300,
	}

	if portStr := os.Getenv("PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Port = port
		}
	}

	flag.IntVar(&cfg.Port, "port", cfg.Port, "Server port")
	flag.Parse()

	return cfg
}
