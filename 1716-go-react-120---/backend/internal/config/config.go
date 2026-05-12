package config

import (
	"flag"
	"os"
)

type Config struct {
	Port string
}

func Load() *Config {
	port := "8300"

	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	flag.StringVar(&port, "port", port, "Server port")
	flag.Parse()

	return &Config{
		Port: port,
	}
}
