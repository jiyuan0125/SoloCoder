package api

import "generic-checker/pkg/types"

type CheckRequest struct {
	Function  types.FunctionSignature `json:"function"`
	CallSites []types.CallSite        `json:"call_sites"`
}

type CheckResponse struct {
	Results []types.CheckResult `json:"results"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
