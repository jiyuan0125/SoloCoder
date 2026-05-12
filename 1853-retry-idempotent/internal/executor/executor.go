package executor

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"idempotent-retry/internal/models"
)

type Executor struct {
	client *http.Client
}

func New() *Executor {
	return &Executor{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (e *Executor) Execute(task *models.Task, attempt int) *models.ExecutionRecord {
	record := &models.ExecutionRecord{
		Attempt:   attempt,
		StartTime: time.Now(),
	}

	req, err := http.NewRequest(task.Method, task.URL, bytes.NewReader(task.RequestBody))
	if err != nil {
		record.EndTime = time.Now()
		record.Success = false
		record.Error = err.Error()
		return record
	}

	req.Header.Set("Content-Type", "application/json")
	if task.IdempotencyKey != "" {
		req.Header.Set("Idempotency-Key", task.IdempotencyKey)
	}

	resp, err := e.client.Do(req)
	record.EndTime = time.Now()

	if err != nil {
		record.Success = false
		record.Error = err.Error()
		return record
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		record.Success = false
		record.Error = err.Error()
		record.StatusCode = resp.StatusCode
		return record
	}

	record.StatusCode = resp.StatusCode
	record.Body = body
	record.Success = resp.StatusCode >= 200 && resp.StatusCode < 300

	return record
}

func (e *Executor) SendCallback(callbackURL string, task *models.Task) *models.CallbackResult {
	result := &models.CallbackResult{
		TaskID: task.ID,
		URL:    callbackURL,
		SentAt: time.Now(),
	}

	payload, err := buildCallbackPayload(task)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return result
	}

	req, err := http.NewRequest(http.MethodPost, callbackURL, bytes.NewReader(payload))
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return result
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()

	result.Success = resp.StatusCode >= 200 && resp.StatusCode < 300
	if !result.Success {
		body, _ := io.ReadAll(resp.Body)
		result.Error = string(body)
	}

	return result
}

func buildCallbackPayload(task *models.Task) ([]byte, error) {
	status := "unknown"
	if task.Status != "" {
		status = string(task.Status)
	}

	var resultBody string
	if task.Result != nil && task.Result.Body != nil {
		resultBody = string(task.Result.Body)
	}

	var resultError string
	if task.Result != nil {
		resultError = task.Result.Error
	}

	payload := `{"task_id":"` + task.ID + `","status":"` + status + `","result":` +
		`{"status_code":` + intToString(resultStatusCode(task)) + `,"body":` + escapeJSON(resultBody) + `,"error":"` + escapeJSON(resultError) + `"}}`

	return []byte(payload), nil
}

func resultStatusCode(task *models.Task) int {
	if task.Result != nil {
		return task.Result.StatusCode
	}
	return 0
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

func escapeJSON(s string) string {
	var buf bytes.Buffer
	buf.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"', '\\', '/', '\b', '\f', '\n', '\r', '\t':
			buf.WriteRune('\\')
			buf.WriteRune(r)
		default:
			if r < 0x20 {
				buf.WriteString(`\u00`)
				buf.WriteByte("0123456789abcdef"[r>>4])
				buf.WriteByte("0123456789abcdef"[r&0xF])
			} else {
				buf.WriteRune(r)
			}
		}
	}
	buf.WriteByte('"')
	return buf.String()
}
