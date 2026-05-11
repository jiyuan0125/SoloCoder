package common

type ImportRequest struct {
	Transactions [][]string `json:"transactions"`
}

type ImportResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	TransactionsCount int `json:"transactions_count"`
}

type SetWeightsRequest struct {
	Weights map[string]float64 `json:"weights"`
}

type SetWeightsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type MineRequest struct {
	MinSup  float64 `json:"min_sup"`
	MinConf float64 `json:"min_conf"`
}

type MineResponse struct {
	Success        bool              `json:"success"`
	Message        string            `json:"message,omitempty"`
	FrequentCount  int               `json:"frequent_count"`
	RulesCount     int               `json:"rules_count"`
	Stats          *MiningStatsData  `json:"stats"`
	TopRules       []*RuleData       `json:"top_rules,omitempty"`
}

type FrequentItemSetResponse struct {
	Success         bool              `json:"success"`
	Message         string            `json:"message,omitempty"`
	FrequentItemSets []*FrequentItemSetData `json:"frequent_item_sets,omitempty"`
}

type RulesResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Rules   []*RuleData `json:"rules,omitempty"`
}

type StatsResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message,omitempty"`
	Stats   *MiningStatsData  `json:"stats,omitempty"`
}

type FrequentItemSetData struct {
	Items           []string `json:"items"`
	WeightedSupport float64  `json:"weighted_support"`
	RawCount        int      `json:"raw_count"`
}

type RuleData struct {
	Antecedent      []string `json:"antecedent"`
	Consequent      []string `json:"consequent"`
	Confidence      float64  `json:"confidence"`
	WeightedSupport float64  `json:"weighted_support"`
}

type MiningStatsData struct {
	ScansCount      int `json:"scans_count"`
	CandidatesCount int `json:"candidates_count"`
	FrequentCount   int `json:"frequent_count"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
