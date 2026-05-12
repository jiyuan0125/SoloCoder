package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port   string
	DBPath string
}

func LoadConfig(cmdPort string) *Config {
	port := "8300"
	if cmdPort != "" {
		port = cmdPort
	} else if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./health_supervision.db"
	}

	return &Config{
		Port:   port,
		DBPath: dbPath,
	}
}

func (c *Config) GetAddress() string {
	return fmt.Sprintf(":%s", c.Port)
}

func ParsePort(portStr string) int {
	p, err := strconv.Atoi(portStr)
	if err != nil {
		return 8300
	}
	return p
}
