package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"insurance-claim/pkg/common"
)

type APIClient struct {
	BaseURL string
	Client  *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
}

func (c *APIClient) doRequest(method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var errResp common.ErrorResponse
		json.Unmarshal(respBody, &errResp)
		if errResp.Error != "" {
			return fmt.Errorf("server error: %s", errResp.Error)
		}
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	if result != nil {
		return json.Unmarshal(respBody, result)
	}
	return nil
}

func (c *APIClient) CreatePolicy(policy *common.Policy) error {
	return c.doRequest("POST", "/api/policies", policy, nil)
}

func (c *APIClient) SubmitClaim(req *common.SubmitClaimRequest) (*common.SubmitClaimResponse, error) {
	var result common.SubmitClaimResponse
	err := c.doRequest("POST", "/api/claims", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *APIClient) ListClaims() (*common.ListCasesResponse, error) {
	var result common.ListCasesResponse
	err := c.doRequest("GET", "/api/claims", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *APIClient) GetCase(caseID string) (*common.GetCaseResponse, error) {
	var result common.GetCaseResponse
	err := c.doRequest("GET", "/api/claims/"+caseID, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *APIClient) AssignInvestigator(req *common.AssignInvestigatorRequest) error {
	return c.doRequest("POST", "/api/claims/"+req.CaseID+"/assign-investigator", req, nil)
}

func (c *APIClient) SubmitInvestigation(req *common.SubmitInvestigationRequest) error {
	return c.doRequest("POST", "/api/claims/"+req.CaseID+"/submit-investigation", req, nil)
}

func (c *APIClient) AssignAssessor(req *common.AssignAssessorRequest) error {
	return c.doRequest("POST", "/api/claims/"+req.CaseID+"/assign-assessor", req, nil)
}

func (c *APIClient) SubmitAssessment(req *common.SubmitAssessmentRequest) error {
	return c.doRequest("POST", "/api/claims/"+req.CaseID+"/submit-assessment", req, nil)
}

func (c *APIClient) CalculatePayout(caseID string) (*common.CalculatePayoutResponse, error) {
	var result common.CalculatePayoutResponse
	err := c.doRequest("POST", "/api/claims/"+caseID+"/calculate-payout", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *APIClient) ApprovePayout(caseID, operator string) error {
	body := map[string]string{"operator": operator}
	return c.doRequest("POST", "/api/claims/"+caseID+"/approve-payout", body, nil)
}

func (c *APIClient) MarkAsPaid(caseID, operator string) error {
	body := map[string]string{"operator": operator}
	return c.doRequest("POST", "/api/claims/"+caseID+"/mark-paid", body, nil)
}

func (c *APIClient) FlagForReview(req *common.FlagForReviewRequest) error {
	return c.doRequest("POST", "/api/claims/"+req.CaseID+"/flag-review", req, nil)
}

func (c *APIClient) ResolveReview(req *common.ResolveReviewRequest) error {
	return c.doRequest("POST", "/api/claims/"+req.CaseID+"/resolve-review", req, nil)
}
