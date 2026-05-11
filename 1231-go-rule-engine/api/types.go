package api

type Rule struct {
	ID          string   `json:"id"`
	Priority    int      `json:"priority"`
	Condition   string   `json:"condition"`
	Actions     []string `json:"actions"`
	Description string   `json:"description,omitempty"`
}

type AddRuleRequest struct {
	Rule *Rule `json:"rule"`
}

type AddRuleResponse struct {
	ID    string `json:"id"`
	Error string `json:"error,omitempty"`
}

type GetRuleResponse struct {
	Rule  *Rule  `json:"rule,omitempty"`
	Error string `json:"error,omitempty"`
}

type ListRulesResponse struct {
	Rules []*Rule `json:"rules"`
}

type UpdateRuleRequest struct {
	Rule *Rule `json:"rule"`
}

type UpdateRuleResponse struct {
	Error string `json:"error,omitempty"`
}

type DeleteRuleResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type EvaluateRequest struct {
	Environment map[string]interface{} `json:"environment"`
}

type ActionResult struct {
	RuleID  string `json:"rule_id"`
	RulePri int    `json:"rule_priority"`
	Action  string `json:"action"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}

type EvaluateResponse struct {
	MatchedRules []Rule         `json:"matched_rules"`
	Results      []ActionResult `json:"action_results"`
	Error        string         `json:"error,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
