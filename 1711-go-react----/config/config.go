package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
}

func Load() *Config {
	godotenv.Load()

	port := os.Getenv("PORT")
	if p, err := strconv.Atoi(os.Getenv("SERVER_PORT")); err == nil && p > 0 {
		port = strconv.Itoa(p)
	}

	return &Config{
		Port: port,
	}
}
