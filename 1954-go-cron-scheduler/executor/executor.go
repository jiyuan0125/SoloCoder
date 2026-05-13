package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"cron-scheduler/models"

	"github.com/google/uuid"
)

const maxResultLength = 2000

type Executor struct {
	httpClient *http.Client
}

func New() *Executor {
	return &Executor{
		httpClient: &http.Client{},
	}
}

func (e *Executor) Execute(task *models.Task) *models.ExecutionLog {
	log := &models.ExecutionLog{
		ID:        uuid.NewString(),
		TaskID:    task.ID,
		StartedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(task.TimeoutSec)*time.Second)
	defer cancel()

	var result string
	var status models.ExecutionStatus

	switch task.Type {
	case models.TaskTypeCommand:
		result, status = e.executeCommand(ctx, task.Command)
	case models.TaskTypeHTTP:
		result, status = e.executeHTTP(ctx, task.CallbackURL)
	default:
		result = fmt.Sprintf("unknown task type: %s", task.Type)
		status = models.ExecutionStatusFailed
	}

	log.FinishedAt = time.Now()
	log.DurationMS = log.FinishedAt.Sub(log.StartedAt).Milliseconds()
	log.Status = status
	log.Result = truncateResult(result)

	return log
}

func (e *Executor) executeCommand(ctx context.Context, command string) (string, models.ExecutionStatus) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "empty command", models.ExecutionStatusFailed
	}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Sprintf("command timeout: %s", err.Error()), models.ExecutionStatusTimeout
		}
		return fmt.Sprintf("command failed: %s\nstdout: %s\nstderr: %s", err.Error(), stdout.String(), stderr.String()), models.ExecutionStatusFailed
	}

	return fmt.Sprintf("stdout: %s\nstderr: %s", stdout.String(), stderr.String()), models.ExecutionStatusSuccess
}

func (e *Executor) executeHTTP(ctx context.Context, url string) (string, models.ExecutionStatus) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Sprintf("create request failed: %s", err.Error()), models.ExecutionStatusFailed
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Sprintf("http request timeout: %s", err.Error()), models.ExecutionStatusTimeout
		}
		return fmt.Sprintf("http request failed: %s", err.Error()), models.ExecutionStatusFailed
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("read response failed: %s", err.Error()), models.ExecutionStatusFailed
	}

	result := fmt.Sprintf("status: %d\nbody: %s", resp.StatusCode, string(body))

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return result, models.ExecutionStatusSuccess
	}

	return result, models.ExecutionStatusFailed
}

func truncateResult(s string) string {
	if len(s) > maxResultLength {
		return s[:maxResultLength] + "...(truncated)"
	}
	return s
}
