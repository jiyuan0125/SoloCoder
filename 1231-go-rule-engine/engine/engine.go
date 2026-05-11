package engine

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type ParsedAction struct {
	Operation string
	Resource  string
	Raw       string
}

func ParseAction(action string) ParsedAction {
	action = strings.TrimSpace(action)
	parts := strings.SplitN(action, ":", 2)
	if len(parts) == 2 {
		return ParsedAction{
			Operation: strings.TrimSpace(parts[0]),
			Resource:  strings.TrimSpace(parts[1]),
			Raw:       action,
		}
	}
	return ParsedAction{
		Operation: action,
		Resource:  "",
		Raw:       action,
	}
}

func ActionsConflict(a1, a2 ParsedAction) bool {
	if a1.Resource != a2.Resource {
		return false
	}
	if a1.Resource == "" {
		return false
	}
	if a1.Operation == a2.Operation {
		return false
	}
	return true
}

type Engine struct {
	store  *Store
	execFn func(ParsedAction) error
}

type EngineOption func(*Engine)

func WithExecutor(fn func(ParsedAction) error) EngineOption {
	return func(e *Engine) {
		e.execFn = fn
	}
}

func NewEngine(store *Store, opts ...EngineOption) *Engine {
	e := &Engine{store: store}
	e.execFn = func(_ ParsedAction) error {
		return nil
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *Engine) AddRule(rule *Rule) {
	e.store.Add(rule)
}

func (e *Engine) GetRule(id string) (*Rule, bool) {
	return e.store.Get(id)
}

func (e *Engine) UpdateRule(rule *Rule) {
	e.store.Update(rule)
}

func (e *Engine) DeleteRule(id string) bool {
	return e.store.Delete(id)
}

func (e *Engine) ListRules() []*Rule {
	return e.store.List()
}

type MatchedRule struct {
	Rule   *Rule
	Parsed []ParsedAction
}

func (e *Engine) Evaluate(env map[string]string) (*EvaluateResult, error) {
	rules := e.store.List()

	var matched []*Rule
	for _, r := range rules {
		ok, err := EvaluateCondition(r.Condition, env)
		if err != nil {
			return nil, err
		}
		if ok {
			matched = append(matched, r)
		}
	}

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].Priority < matched[j].Priority
	})

	matchedCopies := make([]Rule, len(matched))
	for i, r := range matched {
		matchedCopies[i] = *r
	}

	parsedRules := make([]MatchedRule, len(matched))
	for i, r := range matched {
		parsed := make([]ParsedAction, len(r.Actions))
		for j, a := range r.Actions {
			parsed[j] = ParseAction(a)
		}
		parsedRules[i] = MatchedRule{Rule: r, Parsed: parsed}
	}

	usedResources := make(map[string]bool)
	results := make([]ActionResult, 0)

	for _, mr := range parsedRules {
		for _, pa := range mr.Parsed {
			if pa.Resource != "" && usedResources[pa.Resource] {
				results = append(results, ActionResult{
					RuleID:  mr.Rule.ID,
					RulePri: mr.Rule.Priority,
					Action:  pa.Raw,
					Status:  "skipped by conflict",
				})
				continue
			}

			err := e.execFn(pa)
			if err != nil {
				results = append(results, ActionResult{
					RuleID:  mr.Rule.ID,
					RulePri: mr.Rule.Priority,
					Action:  pa.Raw,
					Status:  "failed",
					Error:   err.Error(),
				})
			} else {
				results = append(results, ActionResult{
					RuleID:  mr.Rule.ID,
					RulePri: mr.Rule.Priority,
					Action:  pa.Raw,
					Status:  "executed",
				})
			}

			if pa.Resource != "" {
				usedResources[pa.Resource] = true
			}
		}
	}

	return &EvaluateResult{
		MatchedRules: matchedCopies,
		Results:      results,
	}, nil
}

func ValidateRule(rule *Rule) error {
	if rule.ID == "" {
		return errors.New("rule id is required")
	}
	if rule.Condition == "" {
		return errors.New("condition is required")
	}
	_, err := ParseExpression(rule.Condition)
	if err != nil {
		return fmt.Errorf("invalid condition: %w", err)
	}
	if len(rule.Actions) == 0 {
		return errors.New("at least one action is required")
	}
	return nil
}
