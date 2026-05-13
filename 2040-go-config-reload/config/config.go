package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"reflect"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type ServiceConfig struct {
	Name        string      `yaml:"name"`
	Host        string      `yaml:"host"`
	Port        int         `yaml:"port"`
	Enabled     bool        `yaml:"enabled"`
	Timeout     int         `yaml:"timeout"`
	MaxRetries  int         `yaml:"maxRetries"`
	LogLevel    string      `yaml:"logLevel"`
	Environment string      `yaml:"environment"`
	Features    []string    `yaml:"features"`
	Settings    interface{} `yaml:"settings"`
}

type AppConfig struct {
	Services []ServiceConfig `yaml:"services"`
}

type ConfigChange struct {
	Timestamp   time.Time
	Hash        string
	OldHash     string
	Added       []string
	Removed     []string
	Modified    []string
}

type ConfigChangeNotification struct {
	Change *ConfigChange
	Config *AppConfig
}

type ConfigManager struct {
	mu               sync.RWMutex
	configPath       string
	currentConfig    *AppConfig
	currentHash      string
	lastModifiedTime time.Time
	notifiers        []chan<- ConfigChangeNotification
	stopTicker       chan struct{}
}

func NewConfigManager(configPath string) *ConfigManager {
	return &ConfigManager{
		configPath: configPath,
		stopTicker: make(chan struct{}),
	}
}

func (cm *ConfigManager) GetConfig() *AppConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.currentConfig
}

func (cm *ConfigManager) GetCurrentHash() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.currentHash
}

func (cm *ConfigManager) RegisterNotifier(ch chan<- ConfigChangeNotification) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.notifiers = append(cm.notifiers, ch)
}

func (cm *ConfigManager) StartReloadTicker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				cm.Reload()
			case <-cm.stopTicker:
				ticker.Stop()
				return
			}
		}
	}()
}

func (cm *ConfigManager) StopReloadTicker() {
	close(cm.stopTicker)
}

func (cm *ConfigManager) LoadInitial() error {
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", cm.configPath)
	}

	_, err := cm.forceReload(false)
	return err
}

func (cm *ConfigManager) Reload() (*ConfigChange, error) {
	return cm.forceReload(true)
}

func (cm *ConfigManager) forceReload(notify bool) (*ConfigChange, error) {
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	hash := computeHash(data)

	cm.mu.RLock()
	currentHash := cm.currentHash
	cm.mu.RUnlock()

	if hash == currentHash {
		return nil, nil
	}

	var newConfig AppConfig
	if err := yaml.Unmarshal(data, &newConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := validateConfig(&newConfig); err != nil {
		return nil, err
	}

	change := &ConfigChange{
		Timestamp:   time.Now(),
		Hash:        hash,
		OldHash:     currentHash,
	}

	cm.mu.RLock()
	oldConfig := cm.currentConfig
	cm.mu.RUnlock()

	if oldConfig != nil {
		change.Added, change.Removed, change.Modified = compareConfigs(oldConfig, &newConfig)
	}

	cm.mu.Lock()
	cm.currentConfig = &newConfig
	cm.currentHash = hash
	cm.lastModifiedTime = time.Now()
	cm.mu.Unlock()

	if notify {
		cm.notifyAll(change)
	}

	return change, nil
}

func (cm *ConfigManager) notifyAll(change *ConfigChange) {
	cm.mu.RLock()
	config := cm.currentConfig
	notifiers := make([]chan<- ConfigChangeNotification, len(cm.notifiers))
	copy(notifiers, cm.notifiers)
	cm.mu.RUnlock()

	notification := ConfigChangeNotification{
		Change: change,
		Config: config,
	}

	for _, ch := range notifiers {
		select {
		case ch <- notification:
		default:
			fmt.Printf("WARNING: notifier channel full, skipping notification\n")
		}
	}
}

func computeHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func validateConfig(config *AppConfig) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	for i, svc := range config.Services {
		if svc.Name == "" {
			return fmt.Errorf("service[%d]: name is required", i)
		}
		if svc.Port < 0 || svc.Port > 65535 {
			return fmt.Errorf("service[%s]: port must be between 0 and 65535", svc.Name)
		}
		if svc.Timeout < 0 {
			return fmt.Errorf("service[%s]: timeout must be non-negative", svc.Name)
		}
		if svc.MaxRetries < 0 {
			return fmt.Errorf("service[%s]: maxRetries must be non-negative", svc.Name)
		}
	}
	return nil
}

func compareConfigs(old, new *AppConfig) (added, removed, modified []string) {
	oldMap := make(map[string]ServiceConfig)
	for _, svc := range old.Services {
		oldMap[svc.Name] = svc
	}

	newMap := make(map[string]ServiceConfig)
	for _, svc := range new.Services {
		newMap[svc.Name] = svc
	}

	for name := range newMap {
		if _, exists := oldMap[name]; !exists {
			added = append(added, name)
		}
	}

	for name := range oldMap {
		if _, exists := newMap[name]; !exists {
			removed = append(removed, name)
		}
	}

	for name, newSvc := range newMap {
		if oldSvc, exists := oldMap[name]; exists {
			if !reflect.DeepEqual(oldSvc, newSvc) {
				modified = append(modified, name)
			}
		}
	}

	return added, removed, modified
}
