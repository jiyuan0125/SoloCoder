package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type EnvConfig struct {
	Vars map[string]string `json:"vars"`
}

type EnvsFile struct {
	Envs map[string]EnvConfig `json:"envs"`
}

type StateFile struct {
	Current string `json:"current"`
}

func getConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	configDir := filepath.Join(home, ".env-switcher")
	os.MkdirAll(configDir, 0755)
	return configDir
}

func getEnvsFilePath() string {
	return filepath.Join(getConfigDir(), "envs.json")
}

func getStateFilePath() string {
	return filepath.Join(getConfigDir(), "state.json")
}

func createDefaultEnvsFile() error {
	defaultEnvs := &EnvsFile{
		Envs: map[string]EnvConfig{
			"dev": {
				Vars: map[string]string{
					"DB_HOST":     "localhost",
					"DB_PORT":     "3306",
					"REDIS_HOST":  "localhost",
					"REDIS_PORT":  "6379",
					"API_DOMAIN":  "dev.example.com",
					"LOG_LEVEL":   "debug",
				},
			},
		},
	}
	
	data, err := json.MarshalIndent(defaultEnvs, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(getEnvsFilePath(), data, 0644)
}

func loadEnvsFile() (*EnvsFile, error) {
	envsPath := getEnvsFilePath()
	
	if _, err := os.Stat(envsPath); os.IsNotExist(err) {
		if err := createDefaultEnvsFile(); err != nil {
			return nil, err
		}
	}
	
	data, err := os.ReadFile(envsPath)
	if err != nil {
		return nil, err
	}
	
	var envs EnvsFile
	if err := json.Unmarshal(data, &envs); err != nil {
		return nil, err
	}
	
	if envs.Envs == nil {
		envs.Envs = make(map[string]EnvConfig)
	}
	
	return &envs, nil
}

func saveEnvsFile(envs *EnvsFile) error {
	data, err := json.MarshalIndent(envs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getEnvsFilePath(), data, 0644)
}

func loadStateFile() (*StateFile, error) {
	statePath := getStateFilePath()
	
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return &StateFile{Current: "dev"}, nil
	}
	
	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, err
	}
	
	var state StateFile
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	
	return &state, nil
}

func saveStateFile(state *StateFile) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getStateFilePath(), data, 0644)
}
