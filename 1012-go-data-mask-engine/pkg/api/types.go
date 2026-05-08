package api

type MaskRule struct {
	PrefixKeep int `json:"prefix_keep"`
	SuffixKeep int `json:"suffix_keep"`
}

type Rules struct {
	Phone    *MaskRule `json:"phone,omitempty"`
	IDCard   *MaskRule `json:"idcard,omitempty"`
	BankCard *MaskRule `json:"bankcard,omitempty"`
	Email    *MaskRule `json:"email,omitempty"`
	Name     *MaskRule `json:"name,omitempty"`
}

type MaskRequest struct {
	Text          string `json:"text"`
	OverrideRules *Rules `json:"override_rules,omitempty"`
}

type MaskResponse struct {
	Original string `json:"original"`
	Masked   string `json:"masked"`
}

type RulesResponse struct {
	Rules *Rules `json:"rules"`
}

type UpdateRulesRequest = Rules

type UpdateRulesResponse struct {
	Success bool `json:"success"`
}
