package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-helpdesk-ticket/common"
	"io"
	"net/http"
)

const BaseURL = "http://localhost:8080"

type APIClient struct {
	baseURL string
}

func NewAPIClient() *APIClient {
	return &APIClient{baseURL: BaseURL}
}

func (c *APIClient) doRequest(method, path string, body interface{}) (*common.APIResponse, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp common.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return &apiResp, nil
}

func (c *APIClient) CreateTicket(req *common.CreateTicketRequest) (*common.Ticket, error) {
	resp, err := c.doRequest(http.MethodPost, "/tickets", req)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var ticket common.Ticket
	if err := json.Unmarshal(dataBytes, &ticket); err != nil {
		return nil, err
	}

	return &ticket, nil
}

func (c *APIClient) GetTicket(id string) (*common.Ticket, error) {
	resp, err := c.doRequest(http.MethodGet, "/tickets/"+id, nil)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var ticket common.Ticket
	if err := json.Unmarshal(dataBytes, &ticket); err != nil {
		return nil, err
	}

	return &ticket, nil
}

func (c *APIClient) ListTickets() ([]*common.Ticket, error) {
	resp, err := c.doRequest(http.MethodGet, "/tickets", nil)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var tickets []*common.Ticket
	if err := json.Unmarshal(dataBytes, &tickets); err != nil {
		return nil, err
	}

	return tickets, nil
}

func (c *APIClient) ReassignTicket(ticketID string, req *common.ReassignTicketRequest) error {
	resp, err := c.doRequest(http.MethodPost, "/tickets/"+ticketID+"/reassign", req)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func (c *APIClient) UpdateStatus(ticketID string, req *common.UpdateStatusRequest) error {
	resp, err := c.doRequest(http.MethodPost, "/tickets/"+ticketID+"/status", req)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func (c *APIClient) RateTicket(ticketID string, req *common.RateTicketRequest) error {
	resp, err := c.doRequest(http.MethodPost, "/tickets/"+ticketID+"/rate", req)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func (c *APIClient) LinkTickets(req *common.LinkTicketsRequest) error {
	resp, err := c.doRequest(http.MethodPost, "/tickets/link", req)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func (c *APIClient) TransferTicket(ticketID string, req *common.TransferTicketRequest) error {
	resp, err := c.doRequest(http.MethodPost, "/tickets/"+ticketID+"/transfer", req)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func (c *APIClient) GetLogs(ticketID string) ([]*common.OperationLog, error) {
	resp, err := c.doRequest(http.MethodGet, "/tickets/"+ticketID+"/logs", nil)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var logs []*common.OperationLog
	if err := json.Unmarshal(dataBytes, &logs); err != nil {
		return nil, err
	}

	return logs, nil
}

func (c *APIClient) GetStatistics() (*common.StatisticsReport, error) {
	resp, err := c.doRequest(http.MethodGet, "/tickets/statistics", nil)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var stats common.StatisticsReport
	if err := json.Unmarshal(dataBytes, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

func (c *APIClient) GetHandlers() (map[string]*common.Handler, error) {
	resp, err := c.doRequest(http.MethodGet, "/tickets/handlers", nil)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var handlers map[string]*common.Handler
	if err := json.Unmarshal(dataBytes, &handlers); err != nil {
		return nil, err
	}

	return handlers, nil
}

func (c *APIClient) GetTemplates() (map[string]*common.QuickReplyTemplate, error) {
	resp, err := c.doRequest(http.MethodGet, "/tickets/templates", nil)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}

	var templates map[string]*common.QuickReplyTemplate
	if err := json.Unmarshal(dataBytes, &templates); err != nil {
		return nil, err
	}

	return templates, nil
}

func (c *APIClient) AutoClassify(title, description string) (string, error) {
	req := map[string]string{
		"title":       title,
		"description": description,
	}
	resp, err := c.doRequest(http.MethodPost, "/tickets/autoclassify", req)
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf(resp.Error)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return "", err
	}

	var result map[string]string
	if err := json.Unmarshal(dataBytes, &result); err != nil {
		return "", err
	}

	return result["category"], nil
}
