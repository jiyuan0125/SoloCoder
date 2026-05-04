package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"school-enrollment/common"
)

type APIClient struct {
	baseURL string
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{baseURL: baseURL}
}

func (c *APIClient) request(method, path string, body interface{}) (*common.APIResponse, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp common.APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("响应解析失败: %s", string(respBody))
	}

	return &apiResp, nil
}

func (c *APIClient) CreatePlan(name, grade string, classes, maxPerClass int) (*common.APIResponse, error) {
	req := common.CreatePlanRequest{
		Name:        name,
		Grade:       grade,
		Classes:     classes,
		MaxPerClass: maxPerClass,
	}
	return c.request("POST", "/plans", req)
}

func (c *APIClient) ListPlans() (*common.APIResponse, error) {
	return c.request("GET", "/plans", nil)
}

func (c *APIClient) UpdatePlan(planID string, classes, maxPerClass int) (*common.APIResponse, error) {
	req := common.UpdatePlanRequest{
		Classes:     classes,
		MaxPerClass: maxPerClass,
	}
	return c.request("PUT", "/plans?id="+planID, req)
}

func (c *APIClient) ClosePlan(planID string) (*common.APIResponse, error) {
	req := common.ClosePlanRequest{PlanID: planID}
	return c.request("POST", "/plans/close", req)
}

func (c *APIClient) SubmitRegistration(planID, studentName, idCard, parentPhone, address string) (*common.APIResponse, error) {
	req := common.SubmitRegistrationRequest{
		PlanID:      planID,
		StudentName: studentName,
		IDCard:      idCard,
		ParentPhone: parentPhone,
		Address:     address,
	}
	return c.request("POST", "/registrations", req)
}

func (c *APIClient) QueryRegistration(planID, idCard string) (*common.APIResponse, error) {
	req := common.QueryRegistrationByIDCardRequest{
		PlanID: planID,
		IDCard: idCard,
	}
	return c.request("GET", "/registrations/query?plan_id="+planID+"&id_card="+idCard, req)
}

func (c *APIClient) ListRegistrations(planID, status string) (*common.APIResponse, error) {
	req := common.QueryRegistrationsRequest{
		PlanID: planID,
		Status: status,
	}
	return c.request("GET", "/registrations?plan_id="+planID+"&status="+status, req)
}

func (c *APIClient) ReviewRegistration(regID, action, reason string) (*common.APIResponse, error) {
	req := common.ReviewRegistrationRequest{
		RegistrationID: regID,
		Action:         action,
		RejectReason:   reason,
	}
	return c.request("POST", "/registrations/review", req)
}
