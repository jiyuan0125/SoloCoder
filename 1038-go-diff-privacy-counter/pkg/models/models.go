package models

type CountQueryRequest struct {
	Category string  `json:"category"`
	Epsilon  float64 `json:"epsilon"`
}

type CountQueryResponse struct {
	Count   int     `json:"count"`
	Message string  `json:"message,omitempty"`
	Epsilon float64 `json:"epsilon_used,omitempty"`
}

type BudgetResponse struct {
	RemainingBudget float64 `json:"remaining_budget"`
	TotalBudget     float64 `json:"total_budget"`
	Message         string  `json:"message,omitempty"`
}

type RechargeRequest struct {
	Amount float64 `json:"amount"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
