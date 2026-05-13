package matcher

import (
	"net"
	"net/http"
	"strings"
	"sync"
)

type ConditionType string

const (
	ConditionTypeHeader ConditionType = "header"
	ConditionTypeCookie ConditionType = "cookie"
	ConditionTypeIP     ConditionType = "ip"
)

type Operator string

const (
	OperatorAND Operator = "AND"
	OperatorOR  Operator = "OR"
)

type Condition struct {
	Type   ConditionType `json:"type"`
	Key    string        `json:"key"`
	Value  string        `json:"value"`
	CIDR   string        `json:"cidr,omitempty"`
}

type Rule struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Operator  Operator   `json:"operator"`
	Conditions []Condition `json:"conditions"`
	Tags      map[string]string `json:"tags"`
	Enabled   bool       `json:"enabled"`
}

type RuleStore struct {
	mu    sync.RWMutex
	rules map[string]*Rule
}

func NewRuleStore() *RuleStore {
	return &RuleStore{
		rules: make(map[string]*Rule),
	}
}

func (s *RuleStore) Add(rule *Rule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules[rule.ID] = rule
}

func (s *RuleStore) Remove(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.rules[id]; exists {
		delete(s.rules, id)
		return true
	}
	return false
}

func (s *RuleStore) Update(id string, rule *Rule) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.rules[id]; exists {
		rule.ID = id
		s.rules[id] = rule
		return true
	}
	return false
}

func (s *RuleStore) List() []*Rule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rules := make([]*Rule, 0, len(s.rules))
	for _, r := range s.rules {
		rules = append(rules, r)
	}
	return rules
}

func (s *RuleStore) Get(id string) (*Rule, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rule, exists := s.rules[id]
	return rule, exists
}

func MatchCondition(req *http.Request, cond Condition) bool {
	switch cond.Type {
	case ConditionTypeHeader:
		value := req.Header.Get(cond.Key)
		return strings.EqualFold(value, cond.Value)

	case ConditionTypeCookie:
		cookie, err := req.Cookie(cond.Key)
		if err != nil {
			return false
		}
		return cookie.Value == cond.Value

	case ConditionTypeIP:
		if cond.CIDR == "" {
			return false
		}
		_, cidr, err := net.ParseCIDR(cond.CIDR)
		if err != nil {
			return false
		}
		ipStr := getClientIP(req)
		ip := net.ParseIP(ipStr)
		if ip == nil {
			return false
		}
		return cidr.Contains(ip)

	default:
		return false
	}
}

func MatchRule(req *http.Request, rule *Rule) bool {
	if !rule.Enabled {
		return false
	}
	if len(rule.Conditions) == 0 {
		return false
	}

	switch rule.Operator {
	case OperatorAND:
		for _, cond := range rule.Conditions {
			if !MatchCondition(req, cond) {
				return false
			}
		}
		return true

	case OperatorOR:
		for _, cond := range rule.Conditions {
			if MatchCondition(req, cond) {
				return true
			}
		}
		return false

	default:
		return false
	}
}

func (s *RuleStore) MatchAll(req *http.Request) map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]string)
	for _, rule := range s.rules {
		if MatchRule(req, rule) {
			for k, v := range rule.Tags {
				result[k] = v
			}
		}
	}
	return result
}

func getClientIP(req *http.Request) string {
	if forwarded := req.Header.Get("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	if realIP := req.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return ip
}
