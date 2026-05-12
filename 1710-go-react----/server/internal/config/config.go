package config

import (
	"flag"
	"os"
)

type Config struct {
	Port       string
	DBPath     string
	ServerAddr string
}

func Load() *Config {
	port := flag.String("port", "8080", "Server port")
	dbPath := flag.String("db", "trial.db", "SQLite database path")
	flag.Parse()

	if envPort := os.Getenv("PORT"); envPort != "" {
		port = &envPort
	}
	if envDB := os.Getenv("DB_PATH"); envDB != "" {
		dbPath = &envDB
	}

	return &Config{
		Port:       *port,
		DBPath:     *dbPath,
		ServerAddr: "http://localhost:" + *port,
	}
}
