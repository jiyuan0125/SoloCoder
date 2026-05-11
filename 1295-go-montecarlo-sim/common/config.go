package common

import (
	"encoding/json"
	"os"
)

type BatchTask struct {
	Name        string            `json:"name"`
	Type        SimulationType    `json:"type"`
	Pi          *PiRequest        `json:"pi,omitempty"`
	Option      *OptionRequest    `json:"option,omitempty"`
	Probability *ProbabilityRequest `json:"probability,omitempty"`
}

type BatchConfig struct {
	ServerURL string      `json:"server_url"`
	Tasks     []BatchTask `json:"tasks"`
}

func LoadBatchConfig(path string) (*BatchConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config BatchConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	if config.ServerURL == "" {
		config.ServerURL = "http://localhost:8104"
	}
	return &config, nil
}
