package main

import (
	"encoding/json"
	"errors"
	"future-promise/api"
	"future-promise/promise"
	"strconv"
	"sync"
	"time"
)

type Task struct {
	ID         string
	State      api.TaskState
	Input      interface{}
	Result     interface{}
	Error      error
	StartTime  time.Time
	EndTime    time.Time
	Future     *promise.Future[interface{}]
	CancelFunc func()
}

type TaskManager struct {
	mu    sync.RWMutex
	tasks map[string]*Task
	logs  []api.TaskLogEntry
	counter int64
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks: make(map[string]*Task),
		logs:  make([]api.TaskLogEntry, 0),
	}
}

func (tm *TaskManager) log(taskID, action, details string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.logs = append(tm.logs, api.TaskLogEntry{
		Timestamp: time.Now(),
		TaskID:    taskID,
		Action:    action,
		Details:   details,
	})
}

func (tm *TaskManager) generateID() string {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.counter++
	return strconv.FormatInt(tm.counter, 10)
}

func (tm *TaskManager) SubmitTask(req *api.SubmitTaskRequest) (string, error) {
	taskID := tm.generateID()
	tm.log(taskID, "submit", "type="+req.Type)

	task := &Task{
		ID:        taskID,
		State:     api.StatePending,
		Input:     req.Input,
		StartTime: time.Now(),
	}

	var f *promise.Future[interface{}]
	var cancelFunc func()

	if req.Type == "chain" && len(req.Steps) > 0 {
		f, cancelFunc = tm.executeChainTask(taskID, req.Steps, req.Timeout)
	} else {
		f, cancelFunc = tm.executeSimpleTask(taskID, req.Input, req.Timeout)
	}

	task.Future = f
	task.CancelFunc = cancelFunc
	task.State = api.StateRunning

	tm.mu.Lock()
	tm.tasks[taskID] = task
	tm.mu.Unlock()

	f.Then(func(v interface{}) {
		tm.mu.Lock()
		if t, ok := tm.tasks[taskID]; ok {
			t.State = api.StateCompleted
			t.Result = v
			t.EndTime = time.Now()
		}
		tm.mu.Unlock()
		tm.log(taskID, "complete", "")
	})

	f.Catch(func(err error) {
		tm.mu.Lock()
		if t, ok := tm.tasks[taskID]; ok {
			t.Error = err
			t.EndTime = time.Now()
			if err.Error() == "cancelled" {
				t.State = api.StateCancelled
			} else if err.Error() == "timeout" {
				t.State = api.StateTimeout
			} else {
				t.State = api.StateFailed
			}
		}
		tm.mu.Unlock()
		tm.log(taskID, "error", err.Error())
	})

	return taskID, nil
}

func (tm *TaskManager) executeSimpleTask(taskID string, input interface{}, timeout time.Duration) (*promise.Future[interface{}], func()) {
	p := promise.New(func(resolve func(interface{}), reject func(error), cancel func()) {
		time.Sleep(2 * time.Second)
		result := map[string]interface{}{
			"task_id": taskID,
			"input":   input,
			"time":    time.Now().Format(time.RFC3339),
		}
		resolve(result)
	})

	if timeout > 0 {
		p.Timeout(timeout)
	}

	return p.Future(), func() { p.Cancel() }
}

func (tm *TaskManager) executeChainTask(taskID string, steps []api.TaskStep, timeout time.Duration) (*promise.Future[interface{}], func()) {
	if len(steps) == 0 {
		return nil, nil
	}

	executor := func(resolve func(interface{}), reject func(error), cancel func()) {
		currentValue := interface{}(nil)
		for i, step := range steps {
			_ = i
			tm.log(taskID, "step_start", "name="+step.Name)

			if step.Delay > 0 {
				time.Sleep(step.Delay)
			}

			if step.Fail {
				reject(errors.New("step " + step.Name + " failed"))
				return
			}

			result := map[string]interface{}{
				"step":      step.Name,
				"step_num":  i + 1,
				"input":     currentValue,
				"step_input": step.Input,
				"time":      time.Now().Format(time.RFC3339),
			}
			currentValue = result
			tm.log(taskID, "step_complete", "name="+step.Name)
		}
		resolve(currentValue)
	}

	p := promise.New(executor)
	if timeout > 0 {
		p.Timeout(timeout)
	}

	return p.Future(), func() { p.Cancel() }
}

func (tm *TaskManager) GetTask(taskID string) (*Task, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	task, ok := tm.tasks[taskID]
	return task, ok
}

func (tm *TaskManager) CancelTask(taskID string) bool {
	tm.mu.RLock()
	task, ok := tm.tasks[taskID]
	tm.mu.RUnlock()

	if !ok {
		return false
	}

	if task.CancelFunc != nil {
		task.CancelFunc()
		tm.log(taskID, "cancel", "requested")
		return true
	}
	return false
}

func (tm *TaskManager) GetLogs() []api.TaskLogEntry {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	logs := make([]api.TaskLogEntry, len(tm.logs))
	copy(logs, tm.logs)
	return logs
}

func (t *Task) ToResponse() *api.GetTaskResponse {
	resp := &api.GetTaskResponse{
		Success:   true,
		TaskID:    t.ID,
		State:     t.State,
		Input:     t.Input,
		StartTime: t.StartTime,
		EndTime:   t.EndTime,
	}

	if t.Result != nil {
		resp.Result = t.Result
	}
	if t.Error != nil {
		resp.Error = t.Error.Error()
	}

	return resp
}

func jsonMarshal(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
