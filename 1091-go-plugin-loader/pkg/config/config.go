package config

import (
	"encoding/json"
	"os"
)

type ServerConfig struct {
	PluginDir  string `json:"plugin_dir"`
	Address    string `json:"address"`
}

func Default() *ServerConfig {
	return &ServerConfig{
		PluginDir: "./plugins",
		Address:   ":8080",
	}
}

func Load(path string) (*ServerConfig, error) {
	cfg := Default()

	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
