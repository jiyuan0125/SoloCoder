package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go-mailing-list/pkg/protocol"
)

const defaultBaseURL = "http://localhost:8080"
const defaultTimeout = 30 * time.Second

type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	return &APIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

func (c *APIClient) url(path string) string {
	return c.baseURL + path
}

func (c *APIClient) get(path string, result interface{}) error {
	resp, err := c.httpClient.Get(c.url(path))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *APIClient) post(path string, body interface{}, result interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(c.url(path), "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (c *APIClient) CreateMailingList(name, description string) (*protocol.MailingList, error) {
	req := protocol.CreateMailingListRequest{
		Name:        name,
		Description: description,
	}

	var resp protocol.CreateMailingListResponse
	if err := c.post("/api/lists", req, &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	return resp.MailingList, nil
}

func (c *APIClient) GetMailingList(id string) (*protocol.MailingList, error) {
	var resp protocol.GetMailingListResponse
	if err := c.get(fmt.Sprintf("/api/lists/%s", id), &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	return resp.MailingList, nil
}

func (c *APIClient) ListMailingLists() ([]*protocol.MailingList, error) {
	var resp protocol.ListMailingListsResponse
	if err := c.get("/api/lists", &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	return resp.MailingLists, nil
}

func (c *APIClient) PauseList(listID string) error {
	req := protocol.PauseListRequest{
		ListID: listID,
	}

	var resp protocol.PauseListResponse
	if err := c.post("/api/lists/pause", req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}

	return nil
}

func (c *APIClient) ResumeList(listID string) error {
	req := protocol.ResumeListRequest{
		ListID: listID,
	}

	var resp protocol.ResumeListResponse
	if err := c.post("/api/lists/resume", req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}

	return nil
}

func (c *APIClient) Subscribe(listID, email, name string, customFields map[string]string) (*protocol.SubscriberListEntry, error) {
	req := protocol.SubscribeRequest{
		Email:        email,
		Name:         name,
		CustomFields: customFields,
	}

	var resp protocol.SubscribeResponse
	if err := c.post(fmt.Sprintf("/api/lists/%s/subscribe", listID), req, &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	return resp.Subscriber, nil
}

func (c *APIClient) Unsubscribe(listID, email string) error {
	req := protocol.UnsubscribeRequest{
		Email: email,
	}

	var resp protocol.UnsubscribeResponse
	if err := c.post(fmt.Sprintf("/api/lists/%s/unsubscribe", listID), req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}

	return nil
}

func (c *APIClient) ListSubscribers(listID string) ([]*protocol.SubscriberListEntry, error) {
	var resp protocol.ListSubscribersResponse
	if err := c.get(fmt.Sprintf("/api/lists/%s/subscribers", listID), &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	return resp.Subscribers, nil
}

func (c *APIClient) ImportSubscribers(listID, filePath string) (total, imported, skipped int, err error) {
	req := protocol.ImportSubscribersRequest{
		FilePath: filePath,
	}

	var resp protocol.ImportSubscribersResponse
	if err := c.post(fmt.Sprintf("/api/lists/%s/import", listID), req, &resp); err != nil {
		return 0, 0, 0, err
	}

	if !resp.Success {
		return 0, 0, 0, fmt.Errorf("%s", resp.Error)
	}

	return resp.Total, resp.Imported, resp.Skipped, nil
}

func (c *APIClient) CreateTemplate(name, subject, htmlBody, textBody string, trackOpens, trackClicks bool, variables map[string]string) (*protocol.EmailTemplate, error) {
	req := protocol.CreateTemplateRequest{
		Name:        name,
		Subject:     subject,
		HTMLBody:    htmlBody,
		TextBody:    textBody,
		TrackOpens:  trackOpens,
		TrackClicks: trackClicks,
		Variables:   variables,
	}

	var resp protocol.CreateTemplateResponse
	if err := c.post("/api/templates", req, &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	return resp.Template, nil
}

func (c *APIClient) ListTemplates() ([]*protocol.EmailTemplate, error) {
	var resp protocol.ListTemplatesResponse
	if err := c.get("/api/templates", &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	return resp.Templates, nil
}

func (c *APIClient) SendCampaign(req *protocol.SendCampaignRequest) (*protocol.SendTask, error) {
	var resp protocol.SendCampaignResponse
	if err := c.post("/api/campaigns", req, &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	return resp.Task, nil
}

func (c *APIClient) GetTask(taskID string) (*protocol.SendTask, error) {
	var resp protocol.GetTaskResponse
	if err := c.get(fmt.Sprintf("/api/tasks/%s", taskID), &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	return resp.Task, nil
}

func (c *APIClient) ListTasks() ([]*protocol.SendTask, error) {
	var resp protocol.ListTasksResponse
	if err := c.get("/api/tasks", &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	return resp.Tasks, nil
}

func (c *APIClient) GetSendRecords(taskID string) ([]*protocol.SendRecord, error) {
	var resp protocol.GetSendRecordsResponse
	if err := c.get(fmt.Sprintf("/api/tasks/%s/records", taskID), &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	return resp.SendRecords, nil
}

func (c *APIClient) TrackOpen(taskID, email string) error {
	req := protocol.TrackOpenRequest{
		TaskID: taskID,
		Email:  email,
	}

	var resp protocol.TrackResponse
	if err := c.post("/api/track/open", req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}

	return nil
}

func (c *APIClient) TrackClick(taskID, email, url string) error {
	req := protocol.TrackClickRequest{
		TaskID: taskID,
		Email:  email,
		URL:    url,
	}

	var resp protocol.TrackResponse
	if err := c.post("/api/track/click", req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}

	return nil
}

func (c *APIClient) GetAlerts() ([]*protocol.Alert, error) {
	var resp protocol.GetAlertsResponse
	if err := c.get("/api/alerts", &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	return resp.Alerts, nil
}

func (c *APIClient) ResolveAlert(alertID string) error {
	req := protocol.ResolveAlertRequest{
		AlertID: alertID,
	}

	var resp protocol.ResolveAlertResponse
	if err := c.post("/api/alerts/resolve", req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}

	return nil
}

func (c *APIClient) GetStats() (*protocol.GetStatsResponse, error) {
	var resp protocol.GetStatsResponse
	if err := c.get("/api/stats", &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	return &resp, nil
}
