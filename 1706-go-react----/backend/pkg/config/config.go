package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port int
}

func Load() *Config {
	port := 8210
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}
	return &Config{Port: port}
}
