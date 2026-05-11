package engine

import "sync"

type Rule struct {
	ID          string   `json:"id"`
	Priority    int      `json:"priority"`
	Condition   string   `json:"condition"`
	Actions     []string `json:"actions"`
	Description string   `json:"description,omitempty"`
}

type ActionResult struct {
	RuleID  string `json:"rule_id"`
	RulePri int    `json:"rule_priority"`
	Action  string `json:"action"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}

type EvaluateResult struct {
	MatchedRules []Rule         `json:"matched_rules"`
	Results      []ActionResult `json:"action_results"`
}

type NodeType int

const (
	NodeTypeLiteral NodeType = iota
	NodeTypeComparison
	NodeTypeLogic
	NodeTypeGroup
)

type LogicOp int

const (
	LogicOpAND LogicOp = iota
	LogicOpOR
)

type CompOp int

const (
	CompOpGT CompOp = iota
	CompOpGTE
	CompOpLT
	CompOpLTE
	CompOpEQ
	CompOpNEQ
)

type Node struct {
	Type       NodeType
	Literal    string
	CompOp     CompOp
	LeftVar    string
	RightValue string
	LogicOp    LogicOp
	Children   []*Node
}

type Store struct {
	mu    sync.RWMutex
	rules map[string]*Rule
}

func NewStore() *Store {
	return &Store{rules: make(map[string]*Rule)}
}

func (s *Store) Add(rule *Rule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules[rule.ID] = rule
}

func (s *Store) Get(id string) (*Rule, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.rules[id]
	return r, ok
}

func (s *Store) Update(rule *Rule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules[rule.ID] = rule
}

func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.rules[id]; !ok {
		return false
	}
	delete(s.rules, id)
	return true
}

func (s *Store) List() []*Rule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rules := make([]*Rule, 0, len(s.rules))
	for _, r := range s.rules {
		rules = append(rules, r)
	}
	return rules
}
