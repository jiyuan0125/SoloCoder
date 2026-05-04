package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"insurance-claim/common"
)

type APIClient struct {
	baseURL string
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{baseURL: baseURL}
}

func (c *APIClient) CreatePolicy(req common.CreatePolicyRequest) (*common.Policy, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(fmt.Sprintf("%s/policies/create", c.baseURL), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp common.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if !apiResp.Success {
		return nil, fmt.Errorf(apiResp.Message)
	}

	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return nil, err
	}

	var policy common.Policy
	if err := json.Unmarshal(dataBytes, &policy); err != nil {
		return nil, err
	}

	return &policy, nil
}

func (c *APIClient) SubmitClaim(req common.SubmitClaimRequest) (*common.Claim, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(fmt.Sprintf("%s/claims/submit", c.baseURL), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp common.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if !apiResp.Success {
		return nil, fmt.Errorf(apiResp.Message)
	}

	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return nil, err
	}

	var claim common.Claim
	if err := json.Unmarshal(dataBytes, &claim); err != nil {
		return nil, err
	}

	return &claim, nil
}

func (c *APIClient) ReviewClaim(req common.ReviewClaimRequest) (*common.Claim, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(fmt.Sprintf("%s/claims/review", c.baseURL), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp common.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if !apiResp.Success {
		return nil, fmt.Errorf(apiResp.Message)
	}

	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return nil, err
	}

	var claim common.Claim
	if err := json.Unmarshal(dataBytes, &claim); err != nil {
		return nil, err
	}

	return &claim, nil
}

func (c *APIClient) ConfirmPayment(req common.ConfirmPaymentRequest) (*common.Claim, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(fmt.Sprintf("%s/claims/pay", c.baseURL), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp common.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if !apiResp.Success {
		return nil, fmt.Errorf(apiResp.Message)
	}

	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return nil, err
	}

	var claim common.Claim
	if err := json.Unmarshal(dataBytes, &claim); err != nil {
		return nil, err
	}

	return &claim, nil
}

func (c *APIClient) GetClaim(claimID string) (*common.Claim, error) {
	resp, err := http.Get(fmt.Sprintf("%s/claims/%s", c.baseURL, claimID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp common.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if !apiResp.Success {
		return nil, fmt.Errorf(apiResp.Message)
	}

	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return nil, err
	}

	var claim common.Claim
	if err := json.Unmarshal(dataBytes, &claim); err != nil {
		return nil, err
	}

	return &claim, nil
}

func (c *APIClient) GetPolicyClaims(policyNumber string) ([]*common.Claim, error) {
	resp, err := http.Get(fmt.Sprintf("%s/policies/%s/claims", c.baseURL, policyNumber))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp common.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if !apiResp.Success {
		return nil, fmt.Errorf(apiResp.Message)
	}

	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return nil, err
	}

	var claimsResp common.ClaimsListResponse
	if err := json.Unmarshal(dataBytes, &claimsResp); err != nil {
		return nil, err
	}

	return claimsResp.Claims, nil
}

func (c *APIClient) GetPendingClaims() ([]*common.Claim, error) {
	resp, err := http.Get(fmt.Sprintf("%s/claims/pending", c.baseURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp common.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if !apiResp.Success {
		return nil, fmt.Errorf(apiResp.Message)
	}

	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return nil, err
	}

	var claimsResp common.ClaimsListResponse
	if err := json.Unmarshal(dataBytes, &claimsResp); err != nil {
		return nil, err
	}

	return claimsResp.Claims, nil
}

func (c *APIClient) GetApprovedClaims() ([]*common.Claim, error) {
	resp, err := http.Get(fmt.Sprintf("%s/claims/approved", c.baseURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp common.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if !apiResp.Success {
		return nil, fmt.Errorf(apiResp.Message)
	}

	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return nil, err
	}

	var claimsResp common.ClaimsListResponse
	if err := json.Unmarshal(dataBytes, &claimsResp); err != nil {
		return nil, err
	}

	return claimsResp.Claims, nil
}

func (c *APIClient) ListAllPolicies() ([]*common.Policy, error) {
	resp, err := http.Get(fmt.Sprintf("%s/policies", c.baseURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp common.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if !apiResp.Success {
		return nil, fmt.Errorf(apiResp.Message)
	}

	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return nil, err
	}

	var policiesResp common.PoliciesListResponse
	if err := json.Unmarshal(dataBytes, &policiesResp); err != nil {
		return nil, err
	}

	return policiesResp.Policies, nil
}

func (c *APIClient) ListAllClaims() ([]*common.Claim, error) {
	resp, err := http.Get(fmt.Sprintf("%s/claims", c.baseURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp common.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if !apiResp.Success {
		return nil, fmt.Errorf(apiResp.Message)
	}

	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return nil, err
	}

	var claimsResp common.ClaimsListResponse
	if err := json.Unmarshal(dataBytes, &claimsResp); err != nil {
		return nil, err
	}

	return claimsResp.Claims, nil
}
