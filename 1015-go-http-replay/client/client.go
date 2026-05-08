package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"http-replay/api"
)

type Client struct {
	serverURL string
}

func NewClient(serverURL string) *Client {
	return &Client{serverURL: strings.TrimRight(serverURL, "/")}
}

func (c *Client) CreateRecord(targetURL string) (*api.CreateRecordResponse, error) {
	req := api.CreateRecordRequest{
		TargetURL: targetURL,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.serverURL+"/replay/record", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("server error: %s", string(respBody))
	}

	var result api.CreateRecordResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) StopRecord(sessionID string) (*api.StopRecordResponse, error) {
	req, err := http.NewRequest(http.MethodPost, c.serverURL+"/replay/record/"+sessionID+"/stop", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server error: %s", string(respBody))
	}

	var result api.StopRecordResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetRecord(sessionID string) (*api.RecordedSessionData, error) {
	resp, err := http.Get(c.serverURL + "/replay/record/" + sessionID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server error: %s", string(respBody))
	}

	var result api.RecordedSessionData
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) Play(req api.PlayRequest) (*api.PlayResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.serverURL+"/replay/play", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server error: %s", string(respBody))
	}

	var result api.PlayResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) ListSessions() ([]api.RecordSession, error) {
	resp, err := http.Get(c.serverURL + "/replay/sessions")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server error: %s", string(respBody))
	}

	var result api.RecordSessionList
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return result.Sessions, nil
}

func (c *Client) RecordCommand(targetURL string) error {
	fmt.Printf("Creating recording session for target: %s\n", targetURL)

	resp, err := c.CreateRecord(targetURL)
	if err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}

	fmt.Printf("Session ID: %s\n", resp.SessionID)
	fmt.Printf("Proxy is listening at: %s\n", resp.ProxyAddr)
	fmt.Printf("Send requests to the proxy, they will be recorded and forwarded to %s\n", resp.TargetURL)
	fmt.Println("Press Ctrl+C to stop recording...")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nStopping recording...")

	stopResp, err := c.StopRecord(resp.SessionID)
	if err != nil {
		return fmt.Errorf("failed to stop record: %w", err)
	}

	fmt.Printf("Recording stopped.\n")
	fmt.Printf("Session ID: %s\n", stopResp.SessionID)
	fmt.Printf("Requests recorded: %d\n", stopResp.RequestCount)
	fmt.Printf("Saved to: %s\n", stopResp.FilePath)

	return nil
}

func (c *Client) ListCommand() error {
	sessions, err := c.ListSessions()
	if err != nil {
		return err
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found.")
		return nil
	}

	fmt.Printf("Total sessions: %d\n\n", len(sessions))
	for _, s := range sessions {
		status := "saved"
		if s.Active {
			status = "RECORDING"
		}
		fmt.Printf("Session ID:   %s\n", s.SessionID)
		fmt.Printf("Target URL:   %s\n", s.TargetURL)
		fmt.Printf("Start Time:   %s\n", s.StartTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("Requests:     %d\n", s.RequestCount)
		fmt.Printf("Status:       %s\n", status)
		if s.FilePath != "" {
			fmt.Printf("File:         %s\n", s.FilePath)
		}
		fmt.Println()
	}

	return nil
}

func (c *Client) PlayCommand(sessionID, targetURL, mode string, concurrency int, replaceRules []api.ReplaceRule, variableValues []api.VariableValue) error {
	if sessionID == "" {
		return fmt.Errorf("session_id is required")
	}

	if mode == "" {
		mode = "strict"
	}

	fmt.Printf("Replaying session: %s\n", sessionID)
	fmt.Printf("Mode: %s\n", mode)
	if mode == "fast" && concurrency > 0 {
		fmt.Printf("Concurrency: %d\n", concurrency)
	}
	if targetURL != "" {
		fmt.Printf("Target URL: %s\n", targetURL)
	}
	fmt.Println()

	req := api.PlayRequest{
		SessionID:      sessionID,
		TargetURL:      targetURL,
		Mode:           mode,
		Concurrency:    concurrency,
		ReplaceRules:   replaceRules,
		VariableValues: variableValues,
	}

	result, err := c.Play(req)
	if err != nil {
		return err
	}

	fmt.Printf("=== Replay Result ===\n")
	fmt.Printf("Session ID:    %s\n", result.SessionID)
	fmt.Printf("Total:         %d\n", result.Total)
	fmt.Printf("Success:       %d\n", result.Success)
	fmt.Printf("Failed:        %d\n", result.Failed)
	fmt.Printf("Duration:      %d ms\n", result.DurationMs)
	fmt.Println()

	if len(result.Results) > 0 {
		fmt.Printf("=== Individual Results ===\n")
		for _, r := range result.Results {
			status := "SUCCESS"
			if !r.Success {
				status = "FAILED"
			}
			fmt.Printf("[%s] #%d %s %s - %d ms",
				status, r.Index, r.Method, r.URL, r.DurationMs)
			if r.StatusCode > 0 {
				fmt.Printf(" (status: %d)", r.StatusCode)
			}
			if r.Error != "" {
				fmt.Printf(" (error: %s)", r.Error)
			}
			fmt.Println()
		}
	}

	return nil
}
