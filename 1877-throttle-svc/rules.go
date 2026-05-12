package main

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

func NewRuleStore() *RuleStore {
	return &RuleStore{
		rules: make(map[string]Rule),
	}
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (r *RuleStore) Create(req RuleRequest) Rule {
	r.mu.Lock()
	defer r.mu.Unlock()
	rule := Rule{
		ID:            generateID(),
		Dimension:     req.Dimension,
		Mode:          req.Mode,
		WindowSeconds: req.WindowSeconds,
		MaxRequests:   req.MaxRequests,
		CreatedAt:     time.Now(),
	}
	r.rules[rule.ID] = rule
	return rule
}

func (r *RuleStore) GetAll() []Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Rule, 0, len(r.rules))
	for _, rule := range r.rules {
		result = append(result, rule)
	}
	return result
}

func (r *RuleStore) GetByDimension(dim Dimension) []Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Rule, 0)
	for _, rule := range r.rules {
		if rule.Dimension == dim {
			result = append(result, rule)
		}
	}
	return result
}
