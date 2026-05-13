package store

import (
	"database/sql"

	"go-cicd-pipeline/internal/types"
)

func (s *Store) CreatePipeline(p *types.Pipeline) error {
	tx, err := s.db.Begin()
	if err != nil {
		return FormatError("begin tx", err)
	}
	defer tx.Rollback()

	now := Now()
	_, err = tx.Exec(`
		INSERT INTO pipelines (id, name, description, total_amount, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, p.ID, p.Name, p.Description, p.TotalAmount, now, now)
	if err != nil {
		return FormatError("insert pipeline", err)
	}

	for _, trigger := range p.Triggers {
		_, err = tx.Exec(`
			INSERT INTO pipeline_triggers (pipeline_id, trigger_type)
			VALUES (?, ?)
		`, p.ID, string(trigger))
		if err != nil {
			return FormatError("insert trigger", err)
		}
	}

	for _, phase := range p.Phases {
		_, err = tx.Exec(`
			INSERT INTO phases (id, pipeline_id, type, name, task_mode, planned_amount, order_index)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, phase.ID, p.ID, string(phase.Type), phase.Name, string(phase.TaskMode), phase.PlannedAmount, phase.OrderIndex)
		if err != nil {
			return FormatError("insert phase", err)
		}

		for _, task := range phase.Tasks {
			_, err = tx.Exec(`
				INSERT INTO tasks (id, phase_id, name, script, timeout_sec, failure_strategy, order_index)
				VALUES (?, ?, ?, ?, ?, ?, ?)
			`, task.ID, phase.ID, task.Name, task.Script, task.TimeoutSec, string(task.FailureStrategy), task.OrderIndex)
			if err != nil {
				return FormatError("insert task", err)
			}
		}
	}

	return tx.Commit()
}

func (s *Store) GetPipeline(id string) (*types.Pipeline, error) {
	row := s.db.QueryRow(`
		SELECT id, name, description, total_amount, created_at, updated_at
		FROM pipelines WHERE id = ?
	`, id)

	var p types.Pipeline
	var totalAmount float64
	err := row.Scan(&p.ID, &p.Name, &p.Description, &totalAmount, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, FormatError("scan pipeline", err)
	}
	p.TotalAmount = totalAmount

	triggerRows, err := s.db.Query(`
		SELECT trigger_type FROM pipeline_triggers WHERE pipeline_id = ?
	`, id)
	if err != nil {
		return nil, FormatError("query triggers", err)
	}
	defer triggerRows.Close()

	for triggerRows.Next() {
		var trigger string
		if err := triggerRows.Scan(&trigger); err != nil {
			return nil, FormatError("scan trigger", err)
		}
		p.Triggers = append(p.Triggers, types.TriggerType(trigger))
	}

	phaseRows, err := s.db.Query(`
		SELECT id, type, name, task_mode, planned_amount, order_index
		FROM phases WHERE pipeline_id = ? ORDER BY order_index
	`, id)
	if err != nil {
		return nil, FormatError("query phases", err)
	}
	defer phaseRows.Close()

	for phaseRows.Next() {
		var phase types.Phase
		var phaseType, taskMode string
		var plannedAmount float64
		if err := phaseRows.Scan(&phase.ID, &phaseType, &phase.Name, &taskMode, &plannedAmount, &phase.OrderIndex); err != nil {
			return nil, FormatError("scan phase", err)
		}
		phase.PipelineID = id
		phase.Type = types.PhaseType(phaseType)
		phase.TaskMode = types.TaskMode(taskMode)
		phase.PlannedAmount = plannedAmount

		taskRows, err := s.db.Query(`
			SELECT id, name, script, timeout_sec, failure_strategy, order_index
			FROM tasks WHERE phase_id = ? ORDER BY order_index
		`, phase.ID)
		if err != nil {
			return nil, FormatError("query tasks", err)
		}
		defer taskRows.Close()

		for taskRows.Next() {
			var task types.Task
			var failureStrategy string
			if err := taskRows.Scan(&task.ID, &task.Name, &task.Script, &task.TimeoutSec, &failureStrategy, &task.OrderIndex); err != nil {
				return nil, FormatError("scan task", err)
			}
			task.PhaseID = phase.ID
			task.FailureStrategy = types.FailureStrategy(failureStrategy)
			phase.Tasks = append(phase.Tasks, task)
		}

		p.Phases = append(p.Phases, phase)
	}

	return &p, nil
}

func (s *Store) ListPipelines() ([]*types.Pipeline, error) {
	rows, err := s.db.Query(`
		SELECT id FROM pipelines ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, FormatError("list pipelines", err)
	}
	defer rows.Close()

	var pipelines []*types.Pipeline
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, FormatError("scan pipeline id", err)
		}
		p, err := s.GetPipeline(id)
		if err != nil {
			return nil, err
		}
		if p != nil {
			pipelines = append(pipelines, p)
		}
	}
	return pipelines, nil
}

func (s *Store) UpdatePipeline(p *types.Pipeline) error {
	tx, err := s.db.Begin()
	if err != nil {
		return FormatError("begin tx", err)
	}
	defer tx.Rollback()

	now := Now()
	_, err = tx.Exec(`
		UPDATE pipelines SET name = ?, description = ?, total_amount = ?, updated_at = ?
		WHERE id = ?
	`, p.Name, p.Description, p.TotalAmount, now, p.ID)
	if err != nil {
		return FormatError("update pipeline", err)
	}

	_, err = tx.Exec(`DELETE FROM pipeline_triggers WHERE pipeline_id = ?`, p.ID)
	if err != nil {
		return FormatError("delete triggers", err)
	}

	for _, trigger := range p.Triggers {
		_, err = tx.Exec(`
			INSERT INTO pipeline_triggers (pipeline_id, trigger_type)
			VALUES (?, ?)
		`, p.ID, string(trigger))
		if err != nil {
			return FormatError("insert trigger", err)
		}
	}

	_, err = tx.Exec(`DELETE FROM tasks WHERE phase_id IN (SELECT id FROM phases WHERE pipeline_id = ?)`, p.ID)
	if err != nil {
		return FormatError("delete tasks", err)
	}

	_, err = tx.Exec(`DELETE FROM phases WHERE pipeline_id = ?`, p.ID)
	if err != nil {
		return FormatError("delete phases", err)
	}

	for _, phase := range p.Phases {
		_, err = tx.Exec(`
			INSERT INTO phases (id, pipeline_id, type, name, task_mode, planned_amount, order_index)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, phase.ID, p.ID, string(phase.Type), phase.Name, string(phase.TaskMode), phase.PlannedAmount, phase.OrderIndex)
		if err != nil {
			return FormatError("insert phase", err)
		}

		for _, task := range phase.Tasks {
			_, err = tx.Exec(`
				INSERT INTO tasks (id, phase_id, name, script, timeout_sec, failure_strategy, order_index)
				VALUES (?, ?, ?, ?, ?, ?, ?)
			`, task.ID, phase.ID, task.Name, task.Script, task.TimeoutSec, string(task.FailureStrategy), task.OrderIndex)
			if err != nil {
				return FormatError("insert task", err)
			}
		}
	}

	return tx.Commit()
}

func (s *Store) DeletePipeline(id string) error {
	_, err := s.db.Exec(`DELETE FROM pipelines WHERE id = ?`, id)
	return err
}

func (s *Store) UpdatePipelineTotalAmount(pipelineID string, newTotal float64) error {
	_, err := s.db.Exec(`
		UPDATE pipelines SET total_amount = ?, updated_at = ?
		WHERE id = ?
	`, newTotal, Now(), pipelineID)
	return err
}
