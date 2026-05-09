package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/workstealing/pool/common"
)

type TaskClient struct {
	baseURL string
	client  *http.Client
}

func NewTaskClient(baseURL string) *TaskClient {
	return &TaskClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *TaskClient) SubmitTask(taskType string, payload interface{}, timeout time.Duration) (*common.SubmitTaskResponse, error) {
	req := common.SubmitTaskRequest{
		TaskType: taskType,
		Payload:  payload,
		Timeout:  timeout,
	}

	body, _ := json.Marshal(req)
	resp, err := c.client.Post(c.baseURL+"/submit", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(errBody))
	}

	var result common.SubmitTaskResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return &result, nil
}

func (c *TaskClient) GetResult(taskID string) (*common.GetTaskResultResponse, error) {
	resp, err := c.client.Get(c.baseURL + "/result?task_id=" + taskID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(errBody))
	}

	var result common.GetTaskResultResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return &result, nil
}

func (c *TaskClient) GetStats() (*common.PoolStatsResponse, error) {
	resp, err := c.client.Get(c.baseURL + "/stats")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(errBody))
	}

	var result common.PoolStatsResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return &result, nil
}

func (c *TaskClient) AdjustWorkers(add, remove int) (*common.PoolStatsResponse, error) {
	req := common.AdjustWorkersRequest{Add: add, Remove: remove}
	body, _ := json.Marshal(req)
	resp, err := c.client.Post(c.baseURL+"/adjust", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(errBody))
	}

	var result common.PoolStatsResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return &result, nil
}

func (c *TaskClient) Shutdown(force bool) (*common.ShutdownResponse, error) {
	req := common.ShutdownRequest{Force: force}
	body, _ := json.Marshal(req)
	resp, err := c.client.Post(c.baseURL+"/shutdown", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(errBody))
	}

	var result common.ShutdownResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return &result, nil
}
