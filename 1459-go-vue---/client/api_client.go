package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"contract-management/common"
)

type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

func (c *APIClient) CreateContract(req common.CreateContractRequest) (*common.ContractResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Post(c.BaseURL+"/contracts", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, readError(resp)
	}

	var contract common.ContractResponse
	if err := json.NewDecoder(resp.Body).Decode(&contract); err != nil {
		return nil, err
	}
	return &contract, nil
}

func (c *APIClient) GetContract(id string) (*common.ContractResponse, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/contracts/" + id)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, readError(resp)
	}

	var contract common.ContractResponse
	if err := json.NewDecoder(resp.Body).Decode(&contract); err != nil {
		return nil, err
	}
	return &contract, nil
}

func (c *APIClient) ListContracts() (*common.ListContractsResponse, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/contracts")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, readError(resp)
	}

	var list common.ListContractsResponse
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}
	return &list, nil
}

func (c *APIClient) UpdateContractAmount(contractID string, req common.UpdateContractAmountRequest) (*common.ContractResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPut, c.BaseURL+"/contracts/"+contractID, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, readError(resp)
	}

	var contract common.ContractResponse
	if err := json.NewDecoder(resp.Body).Decode(&contract); err != nil {
		return nil, err
	}
	return &contract, nil
}

func (c *APIClient) CompleteMilestone(contractID, milestoneID string, req common.CompleteMilestoneRequest) (*common.MilestoneResp, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/contracts/%s/milestones/%s", c.BaseURL, contractID, milestoneID)
	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, readError(resp)
	}

	var milestone common.MilestoneResp
	if err := json.NewDecoder(resp.Body).Decode(&milestone); err != nil {
		return nil, err
	}
	return &milestone, nil
}

func (c *APIClient) PayMilestone(contractID, milestoneID string, req common.PayMilestoneRequest) (*common.PaymentResp, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/contracts/%s/milestones/%s/pay", c.BaseURL, contractID, milestoneID)
	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, readError(resp)
	}

	var payment common.PaymentResp
	if err := json.NewDecoder(resp.Body).Decode(&payment); err != nil {
		return nil, err
	}
	return &payment, nil
}

func (c *APIClient) GetProgress(contractID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/contracts/%s/progress", c.BaseURL, contractID)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, readError(resp)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *APIClient) GetAuditLogs(contractID string) (*common.ListAuditLogsResponse, error) {
	url := fmt.Sprintf("%s/contracts/%s/audit-logs", c.BaseURL, contractID)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, readError(resp)
	}

	var result common.ListAuditLogsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func readError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	var errResp common.ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return fmt.Errorf("server error: %s", string(body))
	}
	return fmt.Errorf("%s", errResp.Error)
}
