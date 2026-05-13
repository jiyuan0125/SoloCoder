package store

import (
	"database/sql"

	"go-cicd-pipeline/internal/types"
)

func (s *Store) CreateArtifact(a *types.Artifact) error {
	_, err := s.db.Exec(`
		INSERT INTO artifacts (id, execution_id, task_result_id, name, file_path, size, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, a.ID, a.ExecutionID, NullString(a.TaskResultID), a.Name, a.FilePath, a.Size, Now())
	return err
}

func (s *Store) GetArtifact(id string) (*types.Artifact, error) {
	row := s.db.QueryRow(`
		SELECT id, execution_id, task_result_id, name, file_path, size, created_at
		FROM artifacts WHERE id = ?
	`, id)

	var a types.Artifact
	var taskResultID sql.NullString
	if err := row.Scan(&a.ID, &a.ExecutionID, &taskResultID, &a.Name, &a.FilePath, &a.Size, &a.CreatedAt); err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, FormatError("scan artifact", err)
	}
	a.TaskResultID = ScanNullString(taskResultID)
	return &a, nil
}

func (s *Store) ListArtifacts(executionID string) ([]*types.Artifact, error) {
	rows, err := s.db.Query(`
		SELECT id, execution_id, task_result_id, name, file_path, size, created_at
		FROM artifacts WHERE execution_id = ? ORDER BY created_at
	`, executionID)
	if err != nil {
		return nil, FormatError("list artifacts", err)
	}
	defer rows.Close()

	var artifacts []*types.Artifact
	for rows.Next() {
		var a types.Artifact
		var taskResultID sql.NullString
		if err := rows.Scan(&a.ID, &a.ExecutionID, &taskResultID, &a.Name, &a.FilePath, &a.Size, &a.CreatedAt); err != nil {
			return nil, FormatError("scan artifact", err)
		}
		a.TaskResultID = ScanNullString(taskResultID)
		artifacts = append(artifacts, &a)
	}
	return artifacts, nil
}

func (s *Store) CreateCleanupRecord(cr *types.CleanupRecord) error {
	_, err := s.db.Exec(`
		INSERT INTO cleanup_records (id, execution_id, action, status, result, created_at, finished_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, cr.ID, cr.ExecutionID, cr.Action, string(cr.Status), NullString(cr.Result), Now(), NullTime(cr.FinishedAt))
	return err
}

func (s *Store) UpdateCleanupRecord(cr *types.CleanupRecord) error {
	_, err := s.db.Exec(`
		UPDATE cleanup_records SET status = ?, result = ?, finished_at = ?
		WHERE id = ?
	`, string(cr.Status), NullString(cr.Result), NullTime(cr.FinishedAt), cr.ID)
	return err
}

func (s *Store) ListCleanupRecords(executionID string) ([]*types.CleanupRecord, error) {
	rows, err := s.db.Query(`
		SELECT id, execution_id, action, status, result, created_at, finished_at
		FROM cleanup_records WHERE execution_id = ? ORDER BY created_at
	`, executionID)
	if err != nil {
		return nil, FormatError("list cleanup records", err)
	}
	defer rows.Close()

	var records []*types.CleanupRecord
	for rows.Next() {
		var cr types.CleanupRecord
		var status string
		var result sql.NullString
		var finishedAt sql.NullTime
		if err := rows.Scan(&cr.ID, &cr.ExecutionID, &cr.Action, &status, &result, &cr.CreatedAt, &finishedAt); err != nil {
			return nil, FormatError("scan cleanup record", err)
		}
		cr.Status = types.ExecutionStatus(status)
		cr.Result = ScanNullString(result)
		cr.FinishedAt = ScanNullTime(finishedAt)
		records = append(records, &cr)
	}
	return records, nil
}
