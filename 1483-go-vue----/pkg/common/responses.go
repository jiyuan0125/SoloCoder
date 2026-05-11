package common

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type SubmitClaimResponse struct {
	Success bool       `json:"success"`
	CaseID  string     `json:"case_id"`
	CaseNo  string     `json:"case_no"`
	Error   string     `json:"error,omitempty"`
}

type GetCaseResponse struct {
	Success bool       `json:"success"`
	Case    *ClaimCase `json:"case,omitempty"`
	Error   string     `json:"error,omitempty"`
}

type ListCasesResponse struct {
	Success bool         `json:"success"`
	Cases   []*ClaimCase `json:"cases,omitempty"`
	Error   string       `json:"error,omitempty"`
}

type UpdateCaseResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type CalculatePayoutResponse struct {
	Success bool   `json:"success"`
	Payout  *Payout `json:"payout,omitempty"`
	Error   string `json:"error,omitempty"`
}
