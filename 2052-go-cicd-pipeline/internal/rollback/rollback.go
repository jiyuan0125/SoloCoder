package rollback

import (
	"fmt"

	"go-cicd-pipeline/internal/engine"
	"go-cicd-pipeline/internal/store"
	"go-cicd-pipeline/internal/types"
	"go-cicd-pipeline/internal/utils"
)

type Manager struct {
	store  *store.Store
	engine *engine.Engine
}

func New(s *store.Store, e *engine.Engine) *Manager {
	return &Manager{store: s, engine: e}
}

func (m *Manager) RollbackTo(executionID string) (*types.Execution, error) {
	sourceExec, err := m.store.GetExecution(executionID)
	if err != nil {
		return nil, err
	}
	if sourceExec == nil {
		return nil, fmt.Errorf("source execution not found")
	}

	if sourceExec.Status != types.StatusSuccess {
		return nil, fmt.Errorf("can only rollback to successful execution")
	}

	hasRunning, err := m.store.HasRunningExecution(sourceExec.PipelineID)
	if err != nil {
		return nil, err
	}
	if hasRunning {
		return nil, fmt.Errorf("pipeline already has running execution")
	}

	pipeline, err := m.store.GetPipeline(sourceExec.PipelineID)
	if err != nil {
		return nil, err
	}
	if pipeline == nil {
		return nil, fmt.Errorf("pipeline not found")
	}

	rollbackExec := &types.Execution{
		ID:           utils.GenerateID(),
		PipelineID:   sourceExec.PipelineID,
		TriggerType:  types.TriggerManual,
		Status:       types.StatusPending,
		TargetEnv:    sourceExec.TargetEnv,
		TargetPhase:  types.PhaseDeploy,
		CreatedAt:    utils.Now(),
		IsRollback:   true,
		RollbackFrom: executionID,
	}

	rollbackExec.PhaseResults = m.createPhaseResults(pipeline, rollbackExec.ID)

	if err := m.store.CreateExecution(rollbackExec); err != nil {
		return nil, err
	}

	return rollbackExec, nil
}

func (m *Manager) createPhaseResults(pipeline *types.Pipeline, executionID string) []types.PhaseResult {
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

func (m *Manager) ListHistory(pipelineID string) ([]*types.Execution, error) {
	return m.store.ListExecutions(pipelineID)
}
