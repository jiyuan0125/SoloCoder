package config

import (
	"os"
	"time"
)

type Config struct {
	ServerAddr        string
	DBPath            string
	OrderTimeout      time.Duration
	CleanupInterval   time.Duration
	ReportInterval    time.Duration
}

func Load() *Config {
	return &Config{
		ServerAddr:      getEnv("SERVER_ADDR", ":8080"),
		DBPath:        getEnv("DB_PATH", "./flashsale.db"),
		OrderTimeout:  10 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		ReportInterval:  1 * time.Minute,
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
