package config

import (
	"flag"
	"os"
)

type Config struct {
	Port string
}

func Load() *Config {
	port := flag.String("port", "", "Server port")
	flag.Parse()

	if *port == "" {
		if envPort := os.Getenv("PORT"); envPort != "" {
			*port = envPort
		} else {
			*port = "8503"
		}
	}

	return &Config{
		Port: *port,
	}
}
