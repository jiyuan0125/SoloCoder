package api

type Operation string

const (
	OpCombination         Operation = "C"
	OpCombinationMod      Operation = "Cmod"
	OpPermutation         Operation = "P"
	OpPermutationMod      Operation = "Pmod"
	OpPermutationDup      Operation = "Pdup"
	OpPermutationDupMod   Operation = "PdupMod"
)

type CombinatoricsRequest struct {
	Operation Operation  `json:"operation"`
	N         int64      `json:"n,omitempty"`
	K         int64      `json:"k,omitempty"`
	Mod       int64      `json:"mod,omitempty"`
	Counts    []int64    `json:"counts,omitempty"`
}

type CombinatoricsResponse struct {
	Success bool   `json:"success"`
	Result  string `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}
