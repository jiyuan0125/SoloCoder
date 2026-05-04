package main

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type NonRetryableError struct {
	Message string
}

func (e *NonRetryableError) Error() string {
	return e.Message
}

type TaskRunner struct {
	taskType string
	payload  map[string]any
	result   map[string]any
	attempts int
}

func NewTaskRunner(taskType string, payload map[string]any) *TaskRunner {
	return &TaskRunner{
		taskType: taskType,
		payload:  payload,
		result:   make(map[string]any),
	}
}

func (t *TaskRunner) Run() error {
	t.attempts++

	switch t.taskType {
	case "mock_success":
		return t.runMockSuccess()
	case "mock_failure":
		return t.runMockFailure()
	case "mock_non_retryable":
		return t.runMockNonRetryable()
	case "mock_flaky":
		return t.runMockFlaky()
	default:
		return fmt.Errorf("unknown task type: %s", t.taskType)
	}
}

func (t *TaskRunner) runMockSuccess() error {
	delay := t.getDelayFromPayload()
	if delay > 0 {
		time.Sleep(delay)
	}
	t.result["status"] = "success"
	t.result["attempts"] = t.attempts
	t.result["timestamp"] = time.Now().Unix()
	return nil
}

func (t *TaskRunner) runMockFailure() error {
	delay := t.getDelayFromPayload()
	if delay > 0 {
		time.Sleep(delay)
	}
	return errors.New("mock failure: task always fails")
}

func (t *TaskRunner) runMockNonRetryable() error {
	delay := t.getDelayFromPayload()
	if delay > 0 {
		time.Sleep(delay)
	}
	return &NonRetryableError{
		Message: "non-retryable error: insufficient balance",
	}
}

func (t *TaskRunner) runMockFlaky() error {
	delay := t.getDelayFromPayload()
	if delay > 0 {
		time.Sleep(delay)
	}

	succeedOnAttempt, ok := t.payload["succeed_on_attempt"].(float64)
	if !ok {
		succeedOnAttempt = 3
	}

	if t.attempts >= int(succeedOnAttempt) {
		t.result["status"] = "success"
		t.result["attempts"] = t.attempts
		t.result["timestamp"] = time.Now().Unix()
		return nil
	}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	if random.Float64() < 0.7 {
		return errors.New("mock flaky: transient error occurred")
	}

	t.result["status"] = "success"
	t.result["attempts"] = t.attempts
	t.result["timestamp"] = time.Now().Unix()
	return nil
}

func (t *TaskRunner) getDelayFromPayload() time.Duration {
	delayMs, ok := t.payload["delay_ms"].(float64)
	if !ok {
		return 0
	}
	return time.Duration(delayMs) * time.Millisecond
}

func (t *TaskRunner) GetResult() map[string]any {
	return t.result
}

func (t *TaskRunner) GetAttempts() int {
	return t.attempts
}
