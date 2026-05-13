package database

import (
	"database/sql"
	"log"
	"release-mgmt/internal/model"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func Init() error {
	var err error
	db, err = sql.Open("sqlite", "./release.db")
	if err != nil {
		return err
	}

	if err := createTables(); err != nil {
		return err
	}

	log.Println("Database initialized successfully")
	return nil
}

func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS releases (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		version TEXT UNIQUE NOT NULL,
		description TEXT,
		status TEXT NOT NULL DEFAULT 'draft',
		submitted_by TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS change_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		release_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		type TEXT,
		FOREIGN KEY (release_id) REFERENCES releases(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS approvals (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		release_id INTEGER NOT NULL,
		approver TEXT NOT NULL,
		approved BOOLEAN NOT NULL,
		reason TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (release_id) REFERENCES releases(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS deployments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		release_id INTEGER NOT NULL,
		environment TEXT NOT NULL,
		status TEXT NOT NULL,
		smoke_test_pass BOOLEAN,
		gray_release BOOLEAN,
		gray_percent INTEGER,
		is_rollback BOOLEAN,
		rollback_from INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (release_id) REFERENCES releases(id)
	);

	CREATE TABLE IF NOT EXISTS operation_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		release_id INTEGER NOT NULL,
		operation TEXT NOT NULL,
		operator TEXT NOT NULL,
		details TEXT,
		notes TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (release_id) REFERENCES releases(id)
	);

	CREATE INDEX IF NOT EXISTS idx_releases_version ON releases(version);
	CREATE INDEX IF NOT EXISTS idx_approvals_release ON approvals(release_id);
	CREATE INDEX IF NOT EXISTS idx_deployments_release ON deployments(release_id);
	CREATE INDEX IF NOT EXISTS idx_history_release ON operation_history(release_id);
	`

	_, err := db.Exec(schema)
	return err
}

func GetDB() *sql.DB {
	return db
}

func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

func CreateRelease(release *model.Release) (int64, error) {
	now := time.Now()
	release.CreatedAt = now
	release.UpdatedAt = now
	release.Status = model.StatusDraft

	result, err := db.Exec(`
		INSERT INTO releases (version, description, status, submitted_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, release.Version, release.Description, release.Status, release.SubmittedBy, now, now)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func GetReleaseByVersion(version string) (*model.Release, error) {
	var r model.Release
	err := db.QueryRow(`
		SELECT id, version, description, status, submitted_by, created_at, updated_at
		FROM releases WHERE version = ?
	`, version).Scan(&r.ID, &r.Version, &r.Description, &r.Status, &r.SubmittedBy, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &r, err
}

func GetReleaseByID(id int64) (*model.Release, error) {
	var r model.Release
	err := db.QueryRow(`
		SELECT id, version, description, status, submitted_by, created_at, updated_at
		FROM releases WHERE id = ?
	`, id).Scan(&r.ID, &r.Version, &r.Description, &r.Status, &r.SubmittedBy, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &r, err
}

func GetAllReleases() ([]model.Release, error) {
	rows, err := db.Query(`
		SELECT id, version, description, status, submitted_by, created_at, updated_at
		FROM releases ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var releases []model.Release
	for rows.Next() {
		var r model.Release
		if err := rows.Scan(&r.ID, &r.Version, &r.Description, &r.Status, &r.SubmittedBy, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		releases = append(releases, r)
	}
	return releases, nil
}

func UpdateReleaseStatus(id int64, status model.ReleaseStatus) error {
	_, err := db.Exec(`
		UPDATE releases SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, status, id)
	return err
}

func CreateChangeItem(item *model.ChangeItem) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO change_items (release_id, content, type) VALUES (?, ?, ?)
	`, item.ReleaseID, item.Content, item.Type)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func GetChangeItemsByRelease(releaseID int64) ([]model.ChangeItem, error) {
	rows, err := db.Query(`
		SELECT id, release_id, content, type FROM change_items WHERE release_id = ? ORDER BY id ASC
	`, releaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.ChangeItem
	for rows.Next() {
		var item model.ChangeItem
		if err := rows.Scan(&item.ID, &item.ReleaseID, &item.Content, &item.Type); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func CreateApproval(approval *model.Approval) (int64, error) {
	now := time.Now()
	approval.CreatedAt = now

	result, err := db.Exec(`
		INSERT INTO approvals (release_id, approver, approved, reason, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, approval.ReleaseID, approval.Approver, approval.Approved, approval.Reason, now)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func GetApprovalsByRelease(releaseID int64) ([]model.Approval, error) {
	rows, err := db.Query(`
		SELECT id, release_id, approver, approved, reason, created_at
		FROM approvals WHERE release_id = ? ORDER BY created_at ASC
	`, releaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var approvals []model.Approval
	for rows.Next() {
		var a model.Approval
		if err := rows.Scan(&a.ID, &a.ReleaseID, &a.Approver, &a.Approved, &a.Reason, &a.CreatedAt); err != nil {
			return nil, err
		}
		approvals = append(approvals, a)
	}
	return approvals, nil
}

func GetApprovalByUser(releaseID int64, approver string) (*model.Approval, error) {
	var a model.Approval
	err := db.QueryRow(`
		SELECT id, release_id, approver, approved, reason, created_at
		FROM approvals WHERE release_id = ? AND approver = ?
	`, releaseID, approver).Scan(&a.ID, &a.ReleaseID, &a.Approver, &a.Approved, &a.Reason, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &a, err
}

func CreateDeployment(deployment *model.Deployment) (int64, error) {
	now := time.Now()
	deployment.CreatedAt = now
	deployment.UpdatedAt = now

	result, err := db.Exec(`
		INSERT INTO deployments (release_id, environment, status, smoke_test_pass, gray_release, gray_percent, is_rollback, rollback_from, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, deployment.ReleaseID, deployment.Environment, deployment.Status,
		deployment.SmokeTestPass, deployment.GrayRelease, deployment.GrayPercent,
		deployment.IsRollback, deployment.RollbackFrom, now, now)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func UpdateDeploymentStatus(id int64, status string) error {
	_, err := db.Exec(`
		UPDATE deployments SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, status, id)
	return err
}

func GetLatestCompletedDeployment(environment string) (*model.Deployment, error) {
	var d model.Deployment
	err := db.QueryRow(`
		SELECT id, release_id, environment, status, smoke_test_pass, gray_release, gray_percent, is_rollback, rollback_from, created_at, updated_at
		FROM deployments
		WHERE environment = ? AND status = 'completed' AND is_rollback = 0
		ORDER BY created_at DESC LIMIT 1
	`, environment).Scan(&d.ID, &d.ReleaseID, &d.Environment, &d.Status, &d.SmokeTestPass,
		&d.GrayRelease, &d.GrayPercent, &d.IsRollback, &d.RollbackFrom, &d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &d, err
}

func GetLatestStagingDeployment(releaseID int64) (*model.Deployment, error) {
	var d model.Deployment
	err := db.QueryRow(`
		SELECT id, release_id, environment, status, smoke_test_pass, gray_release, gray_percent, is_rollback, rollback_from, created_at, updated_at
		FROM deployments
		WHERE release_id = ? AND environment = 'staging'
		ORDER BY created_at DESC LIMIT 1
	`, releaseID).Scan(&d.ID, &d.ReleaseID, &d.Environment, &d.Status, &d.SmokeTestPass,
		&d.GrayRelease, &d.GrayPercent, &d.IsRollback, &d.RollbackFrom, &d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &d, err
}

func AddOperationHistory(history *model.OperationHistory) (int64, error) {
	now := time.Now()
	history.CreatedAt = now

	result, err := db.Exec(`
		INSERT INTO operation_history (release_id, operation, operator, details, notes, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, history.ReleaseID, history.Operation, history.Operator, history.Details, history.Notes, now)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func GetOperationHistory(releaseID int64) ([]model.OperationHistory, error) {
	rows, err := db.Query(`
		SELECT id, release_id, operation, operator, details, notes, created_at
		FROM operation_history WHERE release_id = ? ORDER BY created_at ASC
	`, releaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []model.OperationHistory
	for rows.Next() {
		var h model.OperationHistory
		if err := rows.Scan(&h.ID, &h.ReleaseID, &h.Operation, &h.Operator, &h.Details, &h.Notes, &h.CreatedAt); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, nil
}

func UpdateOperationNotes(id int64, notes string) error {
	_, err := db.Exec(`
		UPDATE operation_history SET notes = ? WHERE id = ?
	`, notes, id)
	return err
}

func VersionExists(version string) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM releases WHERE version = ?`, version).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func GetReleaseHistory(releaseID int64) ([]model.OperationHistory, error) {
	return GetOperationHistory(releaseID)
}
