package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"cron-executor/pkg/model"
)

type Executor struct {
	logDir        string
	activeExecs   map[string]*model.Execution
	activeExecsMu sync.RWMutex
	onComplete    func(exec *model.Execution)
}

func NewExecutor(logDir string, onComplete func(exec *model.Execution)) *Executor {
	return &Executor{
		logDir:      logDir,
		activeExecs: make(map[string]*model.Execution),
		onComplete:  onComplete,
	}
}

func (e *Executor) Execute(task *model.Task, execID string, isManual bool) *model.Execution {
	e.activeExecsMu.Lock()
	if _, exists := e.activeExecs[execID]; exists {
		e.activeExecsMu.Unlock()
		return nil
	}

	exec := &model.Execution{
		ID:              execID,
		TaskID:          task.ID,
		TaskName:        task.Name,
		StartTime:       time.Now(),
		Status:          model.ExecStatusRunning,
		RetryCount:      0,
		MaxRetries:      task.MaxRetries,
		IsManualTrigger: isManual,
	}
	e.activeExecs[execID] = exec
	e.activeExecsMu.Unlock()

	go e.runWithRetry(task, exec)

	return exec
}

func (e *Executor) runWithRetry(task *model.Task, exec *model.Execution) {
	var lastStdout, lastStderr string
	var lastExitCode int
	var isTimeout bool

	for retryCount := 0; retryCount <= task.MaxRetries; retryCount++ {
		if retryCount > 0 {
			time.Sleep(task.RetryInterval)
		}

		exec.RetryCount = retryCount
		stdout, stderr, exitCode, timedOut := e.runOnce(task)

		lastStdout = stdout
		lastStderr = stderr
		lastExitCode = exitCode
		isTimeout = timedOut

		if exitCode == 0 && !timedOut {
			exec.Stdout = stdout
			exec.Stderr = stderr
			exec.ExitCode = 0
			exec.IsTimeout = false
			exec.Status = model.ExecStatusSuccess
			e.completeExecution(exec)
			return
		}
	}

	exec.Stdout = lastStdout
	exec.Stderr = lastStderr
	exec.ExitCode = lastExitCode
	exec.IsTimeout = isTimeout

	if isTimeout {
		exec.Status = model.ExecStatusTimeout
	} else {
		exec.Status = model.ExecStatusFailed
	}

	e.completeExecution(exec)
}

func (e *Executor) runOnce(task *model.Task) (stdout string, stderr string, exitCode int, isTimeout bool) {
	ctx, cancel := context.WithTimeout(context.Background(), task.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", task.Command)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()

	stdout = truncateLastLines(stdoutBuf.String(), 100)
	stderr = truncateLastLines(stderrBuf.String(), 100)

	if ctx.Err() == context.DeadlineExceeded {
		return stdout, stderr, -1, true
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return stdout, stderr, exitError.ExitCode(), false
		}
		return stdout, stderr, -1, false
	}

	return stdout, stderr, 0, false
}

func (e *Executor) completeExecution(exec *model.Execution) {
	exec.EndTime = time.Now()
	exec.Duration = exec.EndTime.Sub(exec.StartTime)

	e.activeExecsMu.Lock()
	delete(e.activeExecs, exec.ID)
	e.activeExecsMu.Unlock()

	e.saveLog(exec)

	if e.onComplete != nil {
		e.onComplete(exec)
	}
}

func (e *Executor) saveLog(exec *model.Execution) {
	if e.logDir == "" {
		return
	}

	if err := os.MkdirAll(e.logDir, 0755); err != nil {
		return
	}

	dateStr := exec.StartTime.Format("2006-01-02")
	logFileName := fmt.Sprintf("%s/%s_%s_%s.log", e.logDir, dateStr, exec.TaskName, exec.ID)

	f, err := os.Create(logFileName)
	if err != nil {
		return
	}
	defer f.Close()

	logContent := fmt.Sprintf(`Task: %s
Execution ID: %s
Start Time: %s
End Time: %s
Duration: %s
Exit Code: %d
Status: %s
Is Timeout: %v
Is Manual Trigger: %v
Retry Count: %d

=== STDOUT ===
%s

=== STDERR ===
%s
`,
		exec.TaskName,
		exec.ID,
		exec.StartTime.Format(time.RFC3339),
		exec.EndTime.Format(time.RFC3339),
		exec.Duration.String(),
		exec.ExitCode,
		exec.Status,
		exec.IsTimeout,
		exec.IsManualTrigger,
		exec.RetryCount,
		exec.Stdout,
		exec.Stderr,
	)

	f.WriteString(logContent)
}

func (e *Executor) IsTaskRunning(taskID string) bool {
	e.activeExecsMu.RLock()
	defer e.activeExecsMu.RUnlock()

	for _, exec := range e.activeExecs {
		if exec.TaskID == taskID {
			return true
		}
	}
	return false
}

func (e *Executor) GetActiveExecution(taskID string) *model.Execution {
	e.activeExecsMu.RLock()
	defer e.activeExecsMu.RUnlock()

	for _, exec := range e.activeExecs {
		if exec.TaskID == taskID {
			return exec
		}
	}
	return nil
}

func truncateLastLines(s string, maxLines int) string {
	if maxLines <= 0 {
		return s
	}

	lines := strings.Split(s, "\n")

	if len(lines) <= maxLines {
		return s
	}

	start := len(lines) - maxLines
	return "...\n" + strings.Join(lines[start:], "\n")
}
