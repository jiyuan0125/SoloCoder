package config

import (
	"encoding/json"
	"os"
	"sync"
)

type FieldRule struct {
	External     string      `json:"external"`
	Internal     string      `json:"internal"`
	Type         string      `json:"type"`
	Default      interface{} `json:"default"`
	IsDate       bool        `json:"is_date"`
	IsArray      bool        `json:"is_array"`
	ArrayFilter  string      `json:"array_filter"`
	ArraySortBy  string      `json:"array_sort_by"`
	ArraySortAsc bool        `json:"array_sort_asc"`
	NestedRules  []FieldRule `json:"nested_rules"`
}

type PathRule struct {
	Path    string      `json:"path"`
	Method  string      `json:"method"`
	Request []FieldRule `json:"request"`
	Response []FieldRule `json:"response"`
}

type Config struct {
	Rules []PathRule `json:"rules"`
}

var (
	rulesConfig Config
	configLock  sync.RWMutex
)

func LoadRules(path string) error {
	configLock.Lock()
	defer configLock.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	rulesConfig = config
	return nil
}

func GetRules() Config {
	configLock.RLock()
	defer configLock.RUnlock()
	return rulesConfig
}

func GetPathRule(path, method string) *PathRule {
	configLock.RLock()
	defer configLock.RUnlock()

	for i := range rulesConfig.Rules {
		if rulesConfig.Rules[i].Path == path && rulesConfig.Rules[i].Method == method {
			return &rulesConfig.Rules[i]
		}
	}
	return nil
}

func GetWriteLock() *sync.RWMutex {
	return &configLock
}
