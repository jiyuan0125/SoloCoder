package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Route struct {
	Path       string `json:"path"`
	Backend    string `json:"backend"`
	StripPrefix bool  `json:"strip_prefix"`
	AuthRequired bool `json:"auth_required"`
}

type JWTConfig struct {
	Secret string `json:"secret"`
}

type APIKeyConfig struct {
	Header string   `json:"header"`
	Keys   []string `json:"keys"`
}

type AuthConfig struct {
	Mode   string       `json:"mode"`
	JWT    JWTConfig    `json:"jwt"`
	APIKey APIKeyConfig `json:"api_key"`
}

type LogConfig struct {
	File    string `json:"file"`
	Enabled bool   `json:"enabled"`
}

type Config struct {
	Routes []Route    `json:"routes"`
	Auth   AuthConfig `json:"auth"`
	Log    LogConfig  `json:"log"`
}

type Manager struct {
	configPath string
	current    *Config
	mu         sync.RWMutex
}

func NewManager(configPath string) (*Manager, error) {
	m := &Manager{configPath: configPath}
	if err := m.load(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Manager) load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}

	m.mu.Lock()
	m.current = &cfg
	m.mu.Unlock()

	return nil
}

func (m *Manager) Get() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current
}

func (m *Manager) Watch() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	absPath, _ := filepath.Abs(m.configPath)
	lastStat, _ := os.Stat(absPath)

	for range ticker.C {
		currentStat, err := os.Stat(absPath)
		if err != nil {
			continue
		}

		if lastStat != nil && currentStat.ModTime().After(lastStat.ModTime()) {
			log.Printf("[config] 检测到配置变更，重新加载...")
			if err := m.load(); err != nil {
				log.Printf("[config] 加载失败: %v", err)
			} else {
				log.Printf("[config] 配置已更新")
			}
		}
		lastStat = currentStat
	}
}
