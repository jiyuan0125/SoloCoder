package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-report-scheduler/pkg/common"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type APIClient struct {
	baseURL string
	client  *http.Client
}

func NewAPIClient(host string, port int) *APIClient {
	return &APIClient{
		baseURL: fmt.Sprintf("http://%s:%d/api", host, port),
		client:  &http.Client{},
	}
}

func (c *APIClient) CreateTask(req *common.CreateTaskRequest) (*common.Task, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(c.baseURL+"/tasks", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		var errResp common.ErrorResponse
		json.Unmarshal(respBody, &errResp)
		return nil, fmt.Errorf("%s", errResp.Error)
	}

	var result common.CreateTaskResponse
	json.Unmarshal(respBody, &result)
	return result.Task, nil
}

func (c *APIClient) GetTask(id string) (*common.Task, error) {
	resp, err := c.client.Get(c.baseURL + "/tasks/" + url.PathEscape(id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.Unmarshal(respBody, &errResp)
		return nil, fmt.Errorf("%s", errResp.Error)
	}

	var result common.GetTaskResponse
	json.Unmarshal(respBody, &result)
	return result.Task, nil
}

func (c *APIClient) ListTasks(status *common.TaskStatus) ([]*common.Task, error) {
	urlStr := c.baseURL + "/tasks"
	if status != nil {
		urlStr += "?status=" + url.QueryEscape(string(*status))
	}

	resp, err := c.client.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.Unmarshal(respBody, &errResp)
		return nil, fmt.Errorf("%s", errResp.Error)
	}

	var result common.ListTasksResponse
	json.Unmarshal(respBody, &result)
	return result.Tasks, nil
}

func (c *APIClient) UpdateTask(req *common.UpdateTaskRequest) (*common.Task, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPut, c.baseURL+"/tasks", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.Unmarshal(respBody, &errResp)
		return nil, fmt.Errorf("%s", errResp.Error)
	}

	var result common.UpdateTaskResponse
	json.Unmarshal(respBody, &result)
	return result.Task, nil
}

func (c *APIClient) DeleteTask(id string) error {
	httpReq, err := http.NewRequest(http.MethodDelete, c.baseURL+"/tasks/"+url.PathEscape(id), nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.Unmarshal(respBody, &errResp)
		return fmt.Errorf("%s", errResp.Error)
	}

	return nil
}

func (c *APIClient) PauseTask(id string) (*common.Task, error) {
	httpReq, err := http.NewRequest(http.MethodPost, c.baseURL+"/tasks/"+url.PathEscape(id)+"/pause", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.Unmarshal(respBody, &errResp)
		return nil, fmt.Errorf("%s", errResp.Error)
	}

	var result common.PauseTaskResponse
	json.Unmarshal(respBody, &result)
	return result.Task, nil
}

func (c *APIClient) ResumeTask(id string) (*common.Task, error) {
	httpReq, err := http.NewRequest(http.MethodPost, c.baseURL+"/tasks/"+url.PathEscape(id)+"/resume", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.Unmarshal(respBody, &errResp)
		return nil, fmt.Errorf("%s", errResp.Error)
	}

	var result common.ResumeTaskResponse
	json.Unmarshal(respBody, &result)
	return result.Task, nil
}

func (c *APIClient) TriggerTask(id string) (string, error) {
	httpReq, err := http.NewRequest(http.MethodPost, c.baseURL+"/tasks/"+url.PathEscape(id)+"/trigger", nil)
	if err != nil {
		return "", err
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusAccepted {
		var errResp common.ErrorResponse
		json.Unmarshal(respBody, &errResp)
		return "", fmt.Errorf("%s", errResp.Error)
	}

	var result common.ManualTriggerResponse
	json.Unmarshal(respBody, &result)
	return result.ExecutionID, nil
}

func (c *APIClient) GetExecution(id string) (*common.Execution, error) {
	resp, err := c.client.Get(c.baseURL + "/executions/" + url.PathEscape(id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.Unmarshal(respBody, &errResp)
		return nil, fmt.Errorf("%s", errResp.Error)
	}

	var result common.GetExecutionResponse
	json.Unmarshal(respBody, &result)
	return result.Execution, nil
}

func (c *APIClient) ListExecutions(taskID *string, status *common.ExecutionStatus, triggerType *common.TriggerType) ([]*common.Execution, error) {
	queryParams := make([]string, 0)
	if taskID != nil {
		queryParams = append(queryParams, "task_id="+url.QueryEscape(*taskID))
	}
	if status != nil {
		queryParams = append(queryParams, "status="+url.QueryEscape(string(*status)))
	}
	if triggerType != nil {
		queryParams = append(queryParams, "trigger_type="+url.QueryEscape(string(*triggerType)))
	}

	urlStr := c.baseURL + "/executions"
	if len(queryParams) > 0 {
		urlStr += "?" + strings.Join(queryParams, "&")
	}

	resp, err := c.client.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.Unmarshal(respBody, &errResp)
		return nil, fmt.Errorf("%s", errResp.Error)
	}

	var result common.ListExecutionsResponse
	json.Unmarshal(respBody, &result)
	return result.Executions, nil
}

func ParseStatus(status string) common.TaskStatus {
	return common.TaskStatus(status)
}

func ParseExecutionStatus(status string) common.ExecutionStatus {
	return common.ExecutionStatus(status)
}

func ParseTriggerType(tt string) common.TriggerType {
	return common.TriggerType(tt)
}

func ParseParameters(paramsStr string) map[string]string {
	params := make(map[string]string)
	if paramsStr == "" {
		return params
	}

	pairs := strings.Split(paramsStr, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(kv) == 2 {
			params[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return params
}

func ParseDependsOn(depsStr string) []string {
	if depsStr == "" {
		return nil
	}

	deps := strings.Split(depsStr, ",")
	result := make([]string, 0, len(deps))
	for _, d := range deps {
		trimmed := strings.TrimSpace(d)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func FormatTask(task *common.Task) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("ID:              %s\n", task.ID))
	sb.WriteString(fmt.Sprintf("Name:            %s\n", task.Name))
	sb.WriteString(fmt.Sprintf("Status:          %s\n", task.Status))
	sb.WriteString(fmt.Sprintf("Cron Expression: %s\n", task.CronExpression))
	sb.WriteString(fmt.Sprintf("Next Run Time:   %s\n", task.NextRunTime.Format("2006-01-02 15:04:05")))
	
	if task.LastRunTime != nil {
		sb.WriteString(fmt.Sprintf("Last Run Time:   %s\n", task.LastRunTime.Format("2006-01-02 15:04:05")))
	} else {
		sb.WriteString("Last Run Time:   Never\n")
	}

	sb.WriteString(fmt.Sprintf("Retry Count:     %d\n", task.RetryCount))
	sb.WriteString(fmt.Sprintf("Consecutive Fails: %d\n", task.ConsecutiveFailures))

	if len(task.Parameters) > 0 {
		sb.WriteString("Parameters:\n")
		for k, v := range task.Parameters {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
	}

	if len(task.DependsOn) > 0 {
		sb.WriteString(fmt.Sprintf("Depends On:      %s\n", strings.Join(task.DependsOn, ", ")))
	}

	sb.WriteString(fmt.Sprintf("Created At:      %s\n", task.CreatedAt.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Updated At:      %s\n", task.UpdatedAt.Format("2006-01-02 15:04:05")))

	return sb.String()
}

func FormatTaskList(tasks []*common.Task) string {
	if len(tasks) == 0 {
		return "No tasks found."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d task(s):\n\n", len(tasks)))

	for i, task := range tasks {
		sb.WriteString(fmt.Sprintf("[%d] %s (%s)\n", i+1, task.Name, task.ID))
		sb.WriteString(fmt.Sprintf("    Status: %s | Cron: %s\n", task.Status, task.CronExpression))
		sb.WriteString(fmt.Sprintf("    Next Run: %s\n\n", task.NextRunTime.Format("2006-01-02 15:04:05")))
	}

	return sb.String()
}

func FormatExecution(exec *common.Execution) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("ID:           %s\n", exec.ID))
	sb.WriteString(fmt.Sprintf("Task ID:      %s\n", exec.TaskID))
	sb.WriteString(fmt.Sprintf("Trigger Type: %s\n", exec.TriggerType))
	sb.WriteString(fmt.Sprintf("Status:       %s\n", exec.Status))

	if exec.StartTime != nil {
		sb.WriteString(fmt.Sprintf("Start Time:   %s\n", exec.StartTime.Format("2006-01-02 15:04:05")))
	}
	if exec.EndTime != nil {
		sb.WriteString(fmt.Sprintf("End Time:     %s\n", exec.EndTime.Format("2006-01-02 15:04:05")))
	}
	sb.WriteString(fmt.Sprintf("Duration:     %d ms\n", exec.DurationMs))
	sb.WriteString(fmt.Sprintf("Retry:        %d\n", exec.RetryNumber))

	if exec.ErrorMsg != "" {
		sb.WriteString(fmt.Sprintf("Error:        %s\n", exec.ErrorMsg))
	}

	sb.WriteString(fmt.Sprintf("Created At:   %s\n", exec.CreatedAt.Format("2006-01-02 15:04:05")))

	return sb.String()
}

func FormatExecutionList(executions []*common.Execution) string {
	if len(executions) == 0 {
		return "No executions found."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d execution(s):\n\n", len(executions)))

	for i, exec := range executions {
		sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, exec.ID))
		sb.WriteString(fmt.Sprintf("    Task: %s | Type: %s | Status: %s\n", 
			exec.TaskID, exec.TriggerType, exec.Status))
		sb.WriteString(fmt.Sprintf("    Duration: %dms | Retry: %d\n", 
			exec.DurationMs, exec.RetryNumber))
		sb.WriteString(fmt.Sprintf("    Created: %s\n\n", 
			exec.CreatedAt.Format("2006-01-02 15:04:05")))
	}

	return sb.String()
}

func StrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
