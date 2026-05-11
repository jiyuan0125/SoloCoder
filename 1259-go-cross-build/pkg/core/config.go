package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"go-cross-build/pkg/api"
)

func LoadConfig(path string) (*api.BuildConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config api.BuildConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if len(config.Targets) == 0 {
		return nil, fmt.Errorf("no build targets specified")
	}

	if config.MainPackage == "" {
		config.MainPackage = "."
	}

	if config.OutputDir == "" {
		config.OutputDir = "dist"
	}

	if err := ValidateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func ValidateConfig(config *api.BuildConfig) error {
	if len(config.Targets) == 0 {
		return fmt.Errorf("no build targets specified")
	}

	for i, target := range config.Targets {
		if !IsValidCombination(target.Platform.GOOS, target.Platform.GOARCH) {
			return fmt.Errorf("target %d: invalid platform combination %s/%s",
				i+1, target.Platform.GOOS, target.Platform.GOARCH)
		}
	}

	return nil
}

func SaveConfig(config *api.BuildConfig, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func CreateDefaultConfig() *api.BuildConfig {
	targets := make([]api.BuildTarget, 0, 6)
	for _, p := range GetDefaultTargets() {
		targets = append(targets, api.BuildTarget{
			Platform: api.Platform{
				GOOS:   p.GOOS,
				GOARCH: p.GOARCH,
			},
		})
	}

	return &api.BuildConfig{
		Name:        "default-build",
		Description: "Default cross-platform build configuration",
		Targets:     targets,
		ReleaseMode: false,
		OutputDir:   "dist",
		MainPackage: ".",
	}
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
