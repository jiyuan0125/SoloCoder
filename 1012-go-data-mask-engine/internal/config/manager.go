package config

import (
	"encoding/json"
	"os"
	"sync"

	"mask-engine/internal/mask"
)

type Manager struct {
	mu    sync.RWMutex
	rules *mask.DefaultRules
	path  string
}

func NewManager(configPath string) (*Manager, error) {
	m := &Manager{
		path:  configPath,
		rules: mask.NewDefaultRules(),
	}
	if configPath != "" {
		if err := m.load(); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}
	return m, nil
}

func (m *Manager) GetRules() *mask.DefaultRules {
	m.mu.RLock()
	defer m.mu.RUnlock()
	rulesCopy := *m.rules
	rulesCopy.Phone = copyRule(m.rules.Phone)
	rulesCopy.IDCard = copyRule(m.rules.IDCard)
	rulesCopy.BankCard = copyRule(m.rules.BankCard)
	rulesCopy.Email = copyRule(m.rules.Email)
	rulesCopy.Name = copyRule(m.rules.Name)
	return &rulesCopy
}

func (m *Manager) UpdateRules(newRules *mask.DefaultRules) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if newRules.Phone != nil {
		m.rules.Phone = copyRule(newRules.Phone)
	}
	if newRules.IDCard != nil {
		m.rules.IDCard = copyRule(newRules.IDCard)
	}
	if newRules.BankCard != nil {
		m.rules.BankCard = copyRule(newRules.BankCard)
	}
	if newRules.Email != nil {
		m.rules.Email = copyRule(newRules.Email)
	}
	if newRules.Name != nil {
		m.rules.Name = copyRule(newRules.Name)
	}
	return m.save()
}

func copyRule(r *mask.MaskRule) *mask.MaskRule {
	if r == nil {
		return nil
	}
	copied := *r
	return &copied
}

func (m *Manager) load() error {
	data, err := os.ReadFile(m.path)
	if err != nil {
		return err
	}
	var rules mask.DefaultRules
	if err := json.Unmarshal(data, &rules); err != nil {
		return err
	}
	m.rules = &rules
	return nil
}

func (m *Manager) save() error {
	if m.path == "" {
		return nil
	}
	data, err := json.MarshalIndent(m.rules, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.path, data, 0644)
}
