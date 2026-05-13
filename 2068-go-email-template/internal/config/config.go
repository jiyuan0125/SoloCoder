package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	SMTPFrom     string `json:"smtp_from"`
	DBPath       string `json:"db_path"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config file not found: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config failed: %w", err)
	}

	if cfg.SMTPHost == "" || cfg.SMTPPort == 0 || cfg.SMTPUsername == "" || cfg.SMTPPassword == "" {
		return nil, fmt.Errorf("incomplete SMTP configuration")
	}

	if cfg.DBPath == "" {
		cfg.DBPath = "email_template.db"
	}

	return &cfg, nil
}
