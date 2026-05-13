package engine

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"go-cicd-pipeline/internal/logmanager"
	"go-cicd-pipeline/internal/store"
	"go-cicd-pipeline/internal/types"
	"go-cicd-pipeline/internal/utils"
)

type Engine struct {
	store      *store.Store
	logManager *logmanager.LogManager
	mu         sync.Mutex
	executions map[string]*ExecutionContext
}

type ExecutionContext struct {
	Execution *types.Execution
	Pipeline  *types.Pipeline
	Cancel    context.CancelFunc
}

func New(s *store.Store, lm *logmanager.LogManager) *Engine {
	return &Engine{
		store:      s,
		logManager: lm,
		executions: make(map[string]*ExecutionContext),
	}
}

func (e *Engine) CreateExecution(pipelineID string, triggerType types.TriggerType, targetEnv types.TargetEnv, targetPhase types.PhaseType, isRollback bool, rollbackFrom string) (*types.Execution, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	hasRunning, err := e.store.HasRunningExecution(pipelineID)
	if err != nil {
		return nil, err
	}
	if hasRunning {
		return nil, fmt.Errorf("pipeline already has running execution")
	}

	pipeline, err := e.store.GetPipeline(pipelineID)
	if err != nil {
		return nil, err
	}
	if pipeline == nil {
		return nil, fmt.Errorf("pipeline not found")
	}

	if targetPhase != "" {
		if err := e.validatePhaseOrder(pipeline, targetPhase); err != nil {
			return nil, err
		}

		if targetPhase == types.PhaseDeploy && targetEnv == types.EnvProduction {
			hasTest, err := e.store.HasSuccessfulTestPhase(pipelineID)
			if err != nil {
				return nil, err
			}
			if !hasTest {
				return nil, fmt.Errorf("production deployment requires successful test phase")
			}
		}
	}

	execution := &types.Execution{
		ID:           utils.GenerateID(),
		PipelineID:   pipelineID,
		TriggerType:  triggerType,
		Status:       types.StatusPending,
		TargetEnv:    targetEnv,
		TargetPhase:  targetPhase,
		CreatedAt:    utils.Now(),
		IsRollback:   isRollback,
		RollbackFrom: rollbackFrom,
	}

	execution.PhaseResults = e.createPhaseResults(pipeline, execution.ID)

	if err := e.store.CreateExecution(execution); err != nil {
		return nil, err
	}

	return execution, nil
}

func (e *Engine) validatePhaseOrder(pipeline *types.Pipeline, targetPhase types.PhaseType) error {
	targetIdx := -1
	for i, phase := range pipeline.Phases {
		if phase.Type == targetPhase {
			targetIdx = i
			break
		}
	}

	if targetIdx <= 0 {
		return nil
	}

	for i := 0; i < targetIdx; i++ {
		if pipeline.Phases[i].Type != types.PhaseBuild && i == 0 {
			return fmt.Errorf("cannot skip build phase")
		}
	}
	return nil
}

func (e *Engine) createPhaseResults(pipeline *types.Pipeline, executionID string) []types.PhaseResult {
	var results []types.PhaseResult
	for _, phase := range pipeline.Phases {
		pr := types.PhaseResult{
			ID:          utils.GenerateID(),
			ExecutionID: executionID,
			PhaseType:   phase.Type,
			PhaseName:   phase.Name,
			Status:      types.StatusPending,
		}

		for _, task := range phase.Tasks {
			pr.TaskResults = append(pr.TaskResults, types.TaskResult{
				ID:            utils.GenerateID(),
				PhaseResultID: pr.ID,
				TaskID:        task.ID,
				TaskName:      task.Name,
				Status:        types.StatusPending,
			})
		}

		results = append(results, pr)
	}
	return results
}

func (e *Engine) StartExecution(executionID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	execution, err := e.store.GetExecution(executionID)
	if err != nil {
		return err
	}
	if execution == nil {
		return fmt.Errorf("execution not found")
	}

	pipeline, err := e.store.GetPipeline(execution.PipelineID)
	if err != nil {
		return err
	}
	if pipeline == nil {
		return fmt.Errorf("pipeline not found")
	}

	ctx, cancel := context.WithCancel(context.Background())
	ec := &ExecutionContext{
		Execution: execution,
		Pipeline:  pipeline,
		Cancel:    cancel,
	}
	e.executions[executionID] = ec

	go e.runExecution(ctx, ec)
	return nil
}

func (e *Engine) runExecution(ctx context.Context, ec *ExecutionContext) {
	execution := ec.Execution
	pipeline := ec.Pipeline

	now := utils.Now()
	execution.StartedAt = &now
	execution.Status = types.StatusRunning
	e.store.UpdateExecutionStart(execution.ID, execution.Status, execution.StartedAt)

	defer func() {
		e.mu.Lock()
		delete(e.executions, execution.ID)
		e.mu.Unlock()

		finished := utils.Now()
		execution.FinishedAt = &finished
		e.store.UpdateExecutionFinish(execution.ID, execution.Status, execution.FinishedAt)

		go e.postExecutionCleanup(execution.ID)
	}()

	targetIdx := len(pipeline.Phases)
	if execution.TargetPhase != "" {
		for i, phase := range pipeline.Phases {
			if phase.Type == execution.TargetPhase {
				targetIdx = i + 1
				break
			}
		}
	}

	overallSuccess := true

	for i := 0; i < targetIdx; i++ {
		phase := pipeline.Phases[i]
		phaseResult := &execution.PhaseResults[i]

		select {
		case <-ctx.Done():
			phaseResult.Status = types.StatusCancelled
			e.store.UpdatePhaseResult(phaseResult)
			execution.Status = types.StatusCancelled
			return
		default:
		}

		phaseNow := utils.Now()
		phaseResult.StartedAt = &phaseNow
		phaseResult.Status = types.StatusRunning
		e.store.UpdatePhaseResult(phaseResult)

		phaseSuccess := e.executePhase(ctx, phase, phaseResult)

		phaseFinished := utils.Now()
		phaseResult.FinishedAt = &phaseFinished

		if !phaseSuccess {
			overallSuccess = false
			shouldStop := true
			for _, taskResult := range phaseResult.TaskResults {
				if taskResult.Status == types.StatusFailed || taskResult.Status == types.StatusTimeout {
					task := e.findTask(phase, taskResult.TaskID)
					if task != nil && task.FailureStrategy == types.FailureContinue {
						shouldStop = false
					}
				}
			}

			phaseResult.Status = types.StatusFailed
			e.store.UpdatePhaseResult(phaseResult)

			if shouldStop {
				execution.Status = types.StatusFailed
				return
			}
		} else {
			phaseResult.Status = types.StatusSuccess
			e.store.UpdatePhaseResult(phaseResult)
		}
	}

	if overallSuccess {
		execution.Status = types.StatusSuccess
	} else {
		execution.Status = types.StatusFailed
	}
}

func (e *Engine) executePhase(ctx context.Context, phase types.Phase, phaseResult *types.PhaseResult) bool {
	if phase.TaskMode == types.TaskModeParallel {
		return e.executeTasksParallel(ctx, phase, phaseResult)
	}
	return e.executeTasksSerial(ctx, phase, phaseResult)
}

func (e *Engine) executeTasksSerial(ctx context.Context, phase types.Phase, phaseResult *types.PhaseResult) bool {
	allSuccess := true

	for i, task := range phase.Tasks {
		select {
		case <-ctx.Done():
			phaseResult.TaskResults[i].Status = types.StatusCancelled
			e.store.UpdateTaskResult(&phaseResult.TaskResults[i])
			allSuccess = false
			continue
		default:
		}

		success := e.executeTask(ctx, task, &phaseResult.TaskResults[i])
		if !success {
			allSuccess = false
			if task.FailureStrategy == types.FailureStop {
				for j := i + 1; j < len(phase.Tasks); j++ {
					phaseResult.TaskResults[j].Status = types.StatusSkipped
					e.store.UpdateTaskResult(&phaseResult.TaskResults[j])
				}
				return false
			}
		}
	}

	return allSuccess
}

func (e *Engine) executeTasksParallel(ctx context.Context, phase types.Phase, phaseResult *types.PhaseResult) bool {
	var wg sync.WaitGroup
	results := make(chan bool, len(phase.Tasks))

	for i, task := range phase.Tasks {
		wg.Add(1)
		go func(t types.Task, idx int) {
			defer wg.Done()
			results <- e.executeTask(ctx, t, &phaseResult.TaskResults[idx])
		}(task, i)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	allSuccess := true
	for success := range results {
		if !success {
			allSuccess = false
		}
	}

	return allSuccess
}

func (e *Engine) executeTask(ctx context.Context, task types.Task, taskResult *types.TaskResult) bool {
	logPath, err := e.logManager.CreateLog(taskResult.PhaseResultID, taskResult.ID, taskResult.ID)
	if err != nil {
		taskResult.Status = types.StatusFailed
		e.store.UpdateTaskResult(taskResult)
		return false
	}
	taskResult.LogPath = logPath

	taskNow := utils.Now()
	taskResult.StartedAt = &taskNow
	taskResult.Status = types.StatusRunning
	e.store.UpdateTaskResult(taskResult)

	e.logManager.AppendLog(logPath, fmt.Sprintf("Starting task: %s", task.Name))
	e.logManager.AppendLog(logPath, fmt.Sprintf("Script: %s", task.Script))

	taskCtx, cancel := context.WithTimeout(ctx, time.Duration(task.TimeoutSec)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(taskCtx, "sh", "-c", task.Script)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	taskFinished := utils.Now()
	taskResult.FinishedAt = &taskFinished

	if stdout.Len() > 0 {
		e.logManager.AppendLogRaw(logPath, stdout.String())
	}
	if stderr.Len() > 0 {
		e.logManager.AppendLogRaw(logPath, stderr.String())
	}

	if err != nil {
		if taskCtx.Err() == context.DeadlineExceeded {
			taskResult.Status = types.StatusTimeout
			e.logManager.AppendLog(logPath, fmt.Sprintf("Task timed out after %d seconds", task.TimeoutSec))
		} else {
			taskResult.Status = types.StatusFailed
			e.logManager.AppendLog(logPath, fmt.Sprintf("Task failed: %v", err))
		}
		e.store.UpdateTaskResult(taskResult)
		return false
	}

	taskResult.Status = types.StatusSuccess
	e.logManager.AppendLog(logPath, "Task completed successfully")
	e.store.UpdateTaskResult(taskResult)

	return true
}

func (e *Engine) findTask(phase types.Phase, taskID string) *types.Task {
	for i := range phase.Tasks {
		if phase.Tasks[i].ID == taskID {
			return &phase.Tasks[i]
		}
	}
	return nil
}

func (e *Engine) postExecutionCleanup(executionID string) {
	record := &types.CleanupRecord{
		ID:          utils.GenerateID(),
		ExecutionID: executionID,
		Action:      "post-execution-cleanup",
		Status:      types.StatusRunning,
		CreatedAt:   utils.Now(),
	}
	e.store.CreateCleanupRecord(record)

	time.Sleep(100 * time.Millisecond)

	finished := utils.Now()
	record.FinishedAt = &finished
	record.Status = types.StatusSuccess
	record.Result = "Cleanup completed successfully"
	e.store.UpdateCleanupRecord(record)
}

func (e *Engine) CancelExecution(executionID string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if ec, ok := e.executions[executionID]; ok {
		ec.Cancel()
	}
}
