package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port    int
	DBPath  string
	BaseURL string
}

func Load() *Config {
	_ = godotenv.Load()

	port := 8300
	if p, err := strconv.Atoi(os.Getenv("PORT")); err == nil && p > 0 {
		port = p
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./confman.db"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:" + strconv.Itoa(port)
	}

	log.Printf("Configuration loaded: Port=%d, DB=%s", port, dbPath)

	return &Config{
		Port:    port,
		DBPath:  dbPath,
		BaseURL: baseURL,
	}
}
