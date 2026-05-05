package scheduler

import (
	"bytes"
	"context"
	"os/exec"
	"syscall"
	"time"

	"github.com/google/uuid"

	"scheduler/internal/protocol"
)

type executeResult struct {
	output   string
	errStr   string
	success  bool
	duration time.Duration
}

func (s *Scheduler) executeTask(task *Task) {
	defer func() {
		task.mu.Lock()
		task.IsRunning = false
		if task.CronField != nil && !task.Deleted {
			task.NextRun = task.CronField.Next(time.Now())
		}
		task.mu.Unlock()
	}()

	retryCount := 0
	maxRetry := defaultMaxRetry
	if task.Config.MaxRetry != nil {
		maxRetry = *task.Config.MaxRetry
	}
	retryInterval := task.Config.RetryInterval

	var lastResult executeResult

	for retryCount <= maxRetry {
		result := s.runCommand(task)
		lastResult = result

		if result.success {
			rec := s.createExecutionRecord(task, result, retryCount)
			s.AddExecution(rec)

			task.mu.Lock()
			task.LastExec = &rec
			task.mu.Unlock()
			return
		}

		retryCount++
		if retryCount <= maxRetry {
			time.Sleep(retryInterval)
		}
	}

	rec := s.createExecutionRecord(task, lastResult, retryCount-1)
	s.AddExecution(rec)

	task.mu.Lock()
	task.LastExec = &rec
	task.mu.Unlock()
}

func (s *Scheduler) runCommand(task *Task) executeResult {
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), task.Config.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", task.Config.Command)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Start()
	if err != nil {
		return executeResult{
			output:   "",
			errStr:   err.Error(),
			success:  false,
			duration: time.Since(startTime),
		}
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		if cmd.Process != nil {
			syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		<-done
		return executeResult{
			output:   stdout.String(),
			errStr:   "command timed out after " + task.Config.Timeout.String(),
			success:  false,
			duration: time.Since(startTime),
		}
	case err := <-done:
		if err != nil {
			return executeResult{
				output:   stdout.String(),
				errStr:   err.Error() + "\n" + stderr.String(),
				success:  false,
				duration: time.Since(startTime),
			}
		}
		return executeResult{
			output:   stdout.String(),
			errStr:   "",
			success:  true,
			duration: time.Since(startTime),
		}
	}
}

func (s *Scheduler) createExecutionRecord(task *Task, result executeResult, retryCount int) protocol.ExecutionRecord {
	endTime := time.Now()
	return protocol.ExecutionRecord{
		ID:         uuid.New().String(),
		TaskName:   task.Config.Name,
		StartTime:  endTime.Add(-result.duration),
		EndTime:    endTime,
		Duration:   result.duration,
		Success:    result.success,
		Output:     result.output,
		Error:      result.errStr,
		RetryCount: retryCount,
	}
}
