package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if os.Args != nil && len(os.Args) > 1 {
		for i, arg := range os.Args {
			if arg == "--port" || arg == "-p" {
				if i+1 < len(os.Args) {
					if _, err := strconv.Atoi(os.Args[i+1]); err == nil {
						port = os.Args[i+1]
					}
				}
			}
		}
	}
	return &Config{Port: port}
}
