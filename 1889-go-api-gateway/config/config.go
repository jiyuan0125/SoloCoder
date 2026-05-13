package config

import (
	"os"
)

type Config struct {
	Port         string
	JWTSecret    []byte
	DefaultRoute string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8902"
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-jwt-secret-key-please-change-in-production"
	}
	return &Config{
		Port:      port,
		JWTSecret: []byte(secret),
	}
}
