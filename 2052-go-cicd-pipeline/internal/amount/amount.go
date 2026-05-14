package amount

import (
	"go-cicd-pipeline/internal/store"
	"go-cicd-pipeline/internal/types"
)

type Manager struct {
	store *store.Store
}

func New(s *store.Store) *Manager {
	return &Manager{store: s}
}

func (m *Manager) AdjustTotalAmount(pipelineID string, newTotal float64) error {
	pipeline, err := m.store.GetPipeline(pipelineID)
	if err != nil {
		return err
	}
	if pipeline == nil {
		return nil
	}

	executions, err := m.store.ListExecutions(pipelineID)
	if err != nil {
		return err
	}

	completedPhaseTypes := make(map[types.PhaseType]bool)
	for _, exec := range executions {
		if exec.Status == types.StatusSuccess {
			for _, pr := range exec.PhaseResults {
				if pr.Status == types.StatusSuccess {
					completedPhaseTypes[pr.PhaseType] = true
				}
			}
		}
	}

	var completedTotal float64
	var oldIncompleteTotal float64

	for i := range pipeline.Phases {
		if completedPhaseTypes[pipeline.Phases[i].Type] {
			completedTotal += pipeline.Phases[i].PlannedAmount
		} else {
			oldIncompleteTotal += pipeline.Phases[i].PlannedAmount
		}
	}

	remainingAmount := newTotal - completedTotal
	if remainingAmount < 0 {
		remainingAmount = 0
	}

	if oldIncompleteTotal > 0 && remainingAmount > 0 {
		ratio := remainingAmount / oldIncompleteTotal
		for i := range pipeline.Phases {
			if !completedPhaseTypes[pipeline.Phases[i].Type] {
				pipeline.Phases[i].PlannedAmount *= ratio
			}
		}
	} else if remainingAmount > 0 {
		var incompleteCount int
		for i := range pipeline.Phases {
			if !completedPhaseTypes[pipeline.Phases[i].Type] {
				incompleteCount++
			}
		}
		if incompleteCount > 0 {
			equalShare := remainingAmount / float64(incompleteCount)
			for i := range pipeline.Phases {
				if !completedPhaseTypes[pipeline.Phases[i].Type] {
					pipeline.Phases[i].PlannedAmount = equalShare
				}
			}
		}
	}

	pipeline.TotalAmount = newTotal

	return m.store.UpdatePipeline(pipeline)
}

func (m *Manager) RedistributeForIncompletePhases(pipelineID string) error {
	pipeline, err := m.store.GetPipeline(pipelineID)
	if err != nil {
		return err
	}
	if pipeline == nil {
		return nil
	}

	executions, err := m.store.ListExecutions(pipelineID)
	if err != nil {
		return err
	}

	completedPhaseTypes := make(map[types.PhaseType]bool)
	for _, exec := range executions {
		if exec.Status == types.StatusSuccess || exec.Status == types.StatusFailed {
			for _, pr := range exec.PhaseResults {
				if pr.Status == types.StatusSuccess {
					completedPhaseTypes[pr.PhaseType] = true
				}
			}
		}
	}

	var remainingTotal float64
	var incompletePhases []*types.Phase

	for i := range pipeline.Phases {
		if completedPhaseTypes[pipeline.Phases[i].Type] {
			continue
		}
		remainingTotal += pipeline.Phases[i].PlannedAmount
		incompletePhases = append(incompletePhases, &pipeline.Phases[i])
	}

	if remainingTotal <= 0 || len(incompletePhases) == 0 {
		return nil
	}

	equalShare := remainingTotal / float64(len(incompletePhases))
	for _, phase := range incompletePhases {
		phase.PlannedAmount = equalShare
	}

	return m.store.UpdatePipeline(pipeline)
}
