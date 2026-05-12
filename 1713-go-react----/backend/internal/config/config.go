package config

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	Port   int
	DBPath string
}

func Load() *Config {
	cfg := &Config{
		Port:   8300,
		DBPath: "ohims.db",
	}

	if envPort := os.Getenv("OHIMS_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			cfg.Port = p
		}
	}

	if envDB := os.Getenv("OHIMS_DB_PATH"); envDB != "" {
		cfg.DBPath = envDB
	}

	flag.IntVar(&cfg.Port, "port", cfg.Port, "Server port")
	flag.StringVar(&cfg.DBPath, "db", cfg.DBPath, "Database path")
	flag.Parse()

	return cfg
}
