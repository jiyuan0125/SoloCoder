package store

import (
	"database/sql"
	"time"

	"go-cicd-pipeline/internal/types"
)

func (s *Store) CreateExecution(e *types.Execution) error {
	tx, err := s.db.Begin()
	if err != nil {
		return FormatError("begin tx", err)
	}
	defer tx.Rollback()

	isRollback := 0
	if e.IsRollback {
		isRollback = 1
	}

	_, err = tx.Exec(`
		INSERT INTO executions (id, pipeline_id, trigger_type, status, started_at, finished_at, target_env, target_phase, created_at, is_rollback, rollback_from)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, e.ID, e.PipelineID, string(e.TriggerType), string(e.Status), NullTime(e.StartedAt), NullTime(e.FinishedAt), NullString(string(e.TargetEnv)), NullString(string(e.TargetPhase)), Now(), isRollback, NullString(e.RollbackFrom))
	if err != nil {
		return FormatError("insert execution", err)
	}

	for i, pr := range e.PhaseResults {
		_, err = tx.Exec(`
			INSERT INTO phase_results (id, execution_id, phase_type, phase_name, status, started_at, finished_at, order_index)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, pr.ID, e.ID, string(pr.PhaseType), pr.PhaseName, string(pr.Status), NullTime(pr.StartedAt), NullTime(pr.FinishedAt), i)
		if err != nil {
			return FormatError("insert phase_result", err)
		}

		for j, tr := range pr.TaskResults {
			_, err = tx.Exec(`
				INSERT INTO task_results (id, phase_result_id, task_id, task_name, status, started_at, finished_at, log_path, order_index)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, tr.ID, pr.ID, tr.TaskID, tr.TaskName, string(tr.Status), NullTime(tr.StartedAt), NullTime(tr.FinishedAt), NullString(tr.LogPath), j)
			if err != nil {
				return FormatError("insert task_result", err)
			}
		}
	}

	return tx.Commit()
}

func (s *Store) GetExecution(id string) (*types.Execution, error) {
	row := s.db.QueryRow(`
		SELECT id, pipeline_id, trigger_type, status, started_at, finished_at, target_env, target_phase, created_at, is_rollback, rollback_from
		FROM executions WHERE id = ?
	`, id)

	var e types.Execution
	var triggerType, status string
	var targetEnv, targetPhase sql.NullString
	var startedAt, finishedAt sql.NullTime
	var isRollback int
	var rollbackFrom sql.NullString

	err := row.Scan(&e.ID, &e.PipelineID, &triggerType, &status, &startedAt, &finishedAt, &targetEnv, &targetPhase, &e.CreatedAt, &isRollback, &rollbackFrom)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, FormatError("scan execution", err)
	}

	e.TriggerType = types.TriggerType(triggerType)
	e.Status = types.ExecutionStatus(status)
	e.StartedAt = ScanNullTime(startedAt)
	e.FinishedAt = ScanNullTime(finishedAt)
	if targetEnv.Valid {
		e.TargetEnv = types.TargetEnv(targetEnv.String)
	}
	if targetPhase.Valid {
		e.TargetPhase = types.PhaseType(targetPhase.String)
	}
	e.IsRollback = isRollback == 1
	e.RollbackFrom = ScanNullString(rollbackFrom)

	phaseRows, err := s.db.Query(`
		SELECT id, phase_type, phase_name, status, started_at, finished_at
		FROM phase_results WHERE execution_id = ? ORDER BY order_index
	`, id)
	if err != nil {
		return nil, FormatError("query phase_results", err)
	}
	defer phaseRows.Close()

	for phaseRows.Next() {
		var pr types.PhaseResult
		var phaseType, phaseName, status string
		var startedAt, finishedAt sql.NullTime
		if err := phaseRows.Scan(&pr.ID, &phaseType, &phaseName, &status, &startedAt, &finishedAt); err != nil {
			return nil, FormatError("scan phase_result", err)
		}
		pr.PhaseName = phaseName
		pr.ExecutionID = id
		pr.PhaseType = types.PhaseType(phaseType)
		pr.Status = types.ExecutionStatus(status)
		pr.StartedAt = ScanNullTime(startedAt)
		pr.FinishedAt = ScanNullTime(finishedAt)

		taskRows, err := s.db.Query(`
			SELECT id, task_id, task_name, status, started_at, finished_at, log_path
			FROM task_results WHERE phase_result_id = ? ORDER BY order_index
		`, pr.ID)
		if err != nil {
			return nil, FormatError("query task_results", err)
		}
		defer taskRows.Close()

		for taskRows.Next() {
			var tr types.TaskResult
			var status string
			var startedAt, finishedAt sql.NullTime
			var logPath sql.NullString
			if err := taskRows.Scan(&tr.ID, &tr.TaskID, &tr.TaskName, &status, &startedAt, &finishedAt, &logPath); err != nil {
				return nil, FormatError("scan task_result", err)
			}
			tr.PhaseResultID = pr.ID
			tr.Status = types.ExecutionStatus(status)
			tr.StartedAt = ScanNullTime(startedAt)
			tr.FinishedAt = ScanNullTime(finishedAt)
			tr.LogPath = ScanNullString(logPath)
			pr.TaskResults = append(pr.TaskResults, tr)
		}

		e.PhaseResults = append(e.PhaseResults, pr)
	}

	return &e, nil
}

func (s *Store) ListExecutions(pipelineID string) ([]*types.Execution, error) {
	rows, err := s.db.Query(`
		SELECT id FROM executions WHERE pipeline_id = ? ORDER BY created_at DESC
	`, pipelineID)
	if err != nil {
		return nil, FormatError("list executions", err)
	}
	defer rows.Close()

	var executions []*types.Execution
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, FormatError("scan execution id", err)
		}
		e, err := s.GetExecution(id)
		if err != nil {
			return nil, err
		}
		if e != nil {
			executions = append(executions, e)
		}
	}
	return executions, nil
}

func (s *Store) UpdateExecutionStatus(id string, status types.ExecutionStatus) error {
	_, err := s.db.Exec(`
		UPDATE executions SET status = ? WHERE id = ?
	`, string(status), id)
	return err
}

func (s *Store) UpdateExecutionStart(id string, status types.ExecutionStatus, startedAt *time.Time) error {
	_, err := s.db.Exec(`
		UPDATE executions SET status = ?, started_at = ? WHERE id = ?
	`, string(status), NullTime(startedAt), id)
	return err
}

func (s *Store) UpdateExecutionFinish(id string, status types.ExecutionStatus, finishedAt *time.Time) error {
	_, err := s.db.Exec(`
		UPDATE executions SET status = ?, finished_at = ? WHERE id = ?
	`, string(status), NullTime(finishedAt), id)
	return err
}

func (s *Store) UpdatePhaseResult(pr *types.PhaseResult) error {
	_, err := s.db.Exec(`
		UPDATE phase_results SET status = ?, started_at = ?, finished_at = ?
		WHERE id = ?
	`, string(pr.Status), NullTime(pr.StartedAt), NullTime(pr.FinishedAt), pr.ID)
	return err
}

func (s *Store) UpdateTaskResult(tr *types.TaskResult) error {
	_, err := s.db.Exec(`
		UPDATE task_results SET status = ?, started_at = ?, finished_at = ?, log_path = ?
		WHERE id = ?
	`, string(tr.Status), NullTime(tr.StartedAt), NullTime(tr.FinishedAt), NullString(tr.LogPath), tr.ID)
	return err
}

func (s *Store) HasRunningExecution(pipelineID string) (bool, error) {
	row := s.db.QueryRow(`
		SELECT COUNT(*) FROM executions
		WHERE pipeline_id = ? AND status IN ('pending', 'running')
	`, pipelineID)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Store) HasSuccessfulTestPhase(pipelineID string) (bool, error) {
	row := s.db.QueryRow(`
		SELECT COUNT(*) FROM phase_results pr
		JOIN executions e ON pr.execution_id = e.id
		WHERE e.pipeline_id = ? AND pr.phase_type = ? AND pr.status = ?
		ORDER BY e.created_at DESC LIMIT 1
	`, pipelineID, string(types.PhaseTest), string(types.StatusSuccess))
	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Store) LastSuccessfulDeploy(pipelineID string) (*types.Execution, error) {
	rows, err := s.db.Query(`
		SELECT e.id FROM executions e
		JOIN phase_results pr ON pr.execution_id = e.id
		WHERE e.pipeline_id = ? AND pr.phase_type = ? AND pr.status = ?
		ORDER BY e.created_at DESC LIMIT 1
	`, pipelineID, string(types.PhaseDeploy), string(types.StatusSuccess))
	if err != nil {
		return nil, FormatError("query last successful deploy", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var id string
	if err := rows.Scan(&id); err != nil {
		return nil, FormatError("scan execution id", err)
	}

	return s.GetExecution(id)
}
