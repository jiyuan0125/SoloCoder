package dao

import (
	"database/sql"
	"errors"
	"time"

	"release-flow/model"
)

var (
	ErrReleaseNotFound = errors.New("release not found")
)

type ReleaseDAO struct {
	db *sql.DB
}

func NewReleaseDAO(db *sql.DB) *ReleaseDAO {
	return &ReleaseDAO{db: db}
}

func (dao *ReleaseDAO) Create(release *model.Release) error {
	now := time.Now()
	release.CreatedAt = now
	release.UpdatedAt = now
	release.Status = model.StatusDeveloping
	release.CodeFrozen = false

	result, err := dao.db.Exec(`
		INSERT INTO releases (name, version, status, code_frozen, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, release.Name, release.Version, release.Status, release.CodeFrozen, release.CreatedAt, release.UpdatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	release.ID = id
	return nil
}

func (dao *ReleaseDAO) GetByID(id int64) (*model.Release, error) {
	var release model.Release
	var canaryRatio sql.NullInt64

	err := dao.db.QueryRow(`
		SELECT id, name, version, status, code_frozen, canary_ratio, created_at, updated_at
		FROM releases WHERE id = ?
	`, id).Scan(
		&release.ID, &release.Name, &release.Version, &release.Status,
		&release.CodeFrozen, &canaryRatio, &release.CreatedAt, &release.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrReleaseNotFound
	}
	if err != nil {
		return nil, err
	}

	if canaryRatio.Valid {
		ratio := int(canaryRatio.Int64)
		release.CanaryRatio = &ratio
	}

	return &release, nil
}

func (dao *ReleaseDAO) GetAll() ([]*model.Release, error) {
	rows, err := dao.db.Query(`
		SELECT id, name, version, status, code_frozen, canary_ratio, created_at, updated_at
		FROM releases ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var releases []*model.Release
	for rows.Next() {
		var release model.Release
		var canaryRatio sql.NullInt64

		err := rows.Scan(
			&release.ID, &release.Name, &release.Version, &release.Status,
			&release.CodeFrozen, &canaryRatio, &release.CreatedAt, &release.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if canaryRatio.Valid {
			ratio := int(canaryRatio.Int64)
			release.CanaryRatio = &ratio
		}

		releases = append(releases, &release)
	}

	return releases, rows.Err()
}

func (dao *ReleaseDAO) UpdateStatus(release *model.Release, log *model.StatusLog) error {
	tx, err := dao.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	release.UpdatedAt = now

	query := `UPDATE releases SET status = ?, updated_at = ?`
	args := []interface{}{release.Status, now}

	if release.CanaryRatio != nil {
		query += ", canary_ratio = ?"
		args = append(args, *release.CanaryRatio)
	}

	if release.CodeFrozen {
		query += ", code_frozen = ?"
		args = append(args, true)
	}

	query += " WHERE id = ?"
	args = append(args, release.ID)

	_, err = tx.Exec(query, args...)
	if err != nil {
		return err
	}

	log.CreatedAt = now
	_, err = tx.Exec(`
		INSERT INTO status_logs (release_id, from_status, to_status, operator, reason, canary_ratio, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, log.ReleaseID, log.FromStatus, log.ToStatus, log.Operator, log.Reason, log.CanaryRatio, log.CreatedAt)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (dao *ReleaseDAO) FreezeCode(id int64) error {
	_, err := dao.db.Exec(`
		UPDATE releases SET code_frozen = 1 WHERE id = ?
	`, id)
	return err
}

func (dao *ReleaseDAO) GetStatusLogs(releaseID int64) ([]*model.StatusLog, error) {
	rows, err := dao.db.Query(`
		SELECT id, release_id, from_status, to_status, operator, reason, canary_ratio, created_at
		FROM status_logs WHERE release_id = ? ORDER BY id DESC
	`, releaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*model.StatusLog
	for rows.Next() {
		var log model.StatusLog
		var canaryRatio sql.NullInt64

		err := rows.Scan(
			&log.ID, &log.ReleaseID, &log.FromStatus, &log.ToStatus,
			&log.Operator, &log.Reason, &canaryRatio, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if canaryRatio.Valid {
			ratio := int(canaryRatio.Int64)
			log.CanaryRatio = &ratio
		}

		logs = append(logs, &log)
	}

	return logs, rows.Err()
}

func (dao *ReleaseDAO) GetLastCompletedRelease() (*model.Release, error) {
	var release model.Release
	var canaryRatio sql.NullInt64

	err := dao.db.QueryRow(`
		SELECT id, name, version, status, code_frozen, canary_ratio, created_at, updated_at
		FROM releases WHERE status = ? ORDER BY id DESC LIMIT 1
	`, model.StatusCompleted).Scan(
		&release.ID, &release.Name, &release.Version, &release.Status,
		&release.CodeFrozen, &canaryRatio, &release.CreatedAt, &release.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if canaryRatio.Valid {
		ratio := int(canaryRatio.Int64)
		release.CanaryRatio = &ratio
	}

	return &release, nil
}

func (dao *ReleaseDAO) CreateRollbackRecord(record *model.RollbackRecord) error {
	record.CreatedAt = time.Now()
	_, err := dao.db.Exec(`
		INSERT INTO rollback_records (release_id, from_version, to_version, operator, reason, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, record.ReleaseID, record.FromVersion, record.ToVersion, record.Operator, record.Reason, record.CreatedAt)
	return err
}

func (dao *ReleaseDAO) GetRollbackRecords(releaseID int64) ([]*model.RollbackRecord, error) {
	rows, err := dao.db.Query(`
		SELECT id, release_id, from_version, to_version, operator, reason, created_at
		FROM rollback_records WHERE release_id = ? ORDER BY id DESC
	`, releaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*model.RollbackRecord
	for rows.Next() {
		var record model.RollbackRecord
		err := rows.Scan(
			&record.ID, &record.ReleaseID, &record.FromVersion, &record.ToVersion,
			&record.Operator, &record.Reason, &record.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, &record)
	}

	return records, rows.Err()
}
