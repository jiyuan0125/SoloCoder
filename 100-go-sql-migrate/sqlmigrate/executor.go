package sqlmigrate

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Executor struct {
	db      *sql.DB
	dryRun  bool
}

func NewExecutor(db *sql.DB, dryRun bool) *Executor {
	return &Executor{
		db:      db,
		dryRun:  dryRun,
	}
}

func (e *Executor) EnsureSchemaMigrationsTable() error {
	if e.dryRun {
		fmt.Println("Would ensure schema_migrations table exists")
		return nil
	}

	createTableSQL := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP
		)
	`

	_, err := e.db.Exec(createTableSQL)
	return err
}

func (e *Executor) GetAppliedVersions() (map[string]bool, error) {
	applied := make(map[string]bool)

	if e.dryRun {
		return applied, nil
	}

	rows, err := e.db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, nil
}

func (e *Executor) ExecuteUp(migration *Migration) error {
	fmt.Printf("Executing up migration: %s_%s\n", migration.Version, migration.Description)

	if e.dryRun {
		fmt.Printf("Would execute SQL:\n%s\n\n", migration.UpSQL)
		return nil
	}

	if migration.UpSQL == "" {
		return fmt.Errorf("no up SQL found for version %s", migration.Version)
	}

	// 检查是否已应用
	applied, err := e.GetAppliedVersions()
	if err != nil {
		return err
	}

	if applied[migration.Version] {
		fmt.Printf("Migration %s already applied, skipping.\n", migration.Version)
		return nil
	}

	// 执行 SQL
	err = e.executeSQL(migration.UpSQL)
	if err != nil {
		return fmt.Errorf("failed to execute up migration %s: %w", migration.Version, err)
	}

	// 记录到 schema_migrations
	_, err = e.db.Exec(
		"INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)",
		migration.Version,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
	}

	fmt.Printf("Successfully applied migration: %s\n", migration.Version)
	return nil
}

func (e *Executor) ExecuteDown(migration *Migration) error {
	fmt.Printf("Executing down migration: %s_%s\n", migration.Version, migration.Description)

	if e.dryRun {
		fmt.Printf("Would execute SQL:\n%s\n\n", migration.DownSQL)
		return nil
	}

	if migration.DownSQL == "" {
		return fmt.Errorf("no down SQL found for version %s", migration.Version)
	}

	// 检查是否已应用
	applied, err := e.GetAppliedVersions()
	if err != nil {
		return err
	}

	if !applied[migration.Version] {
		fmt.Printf("Migration %s not applied, skipping.\n", migration.Version)
		return nil
	}

	// 执行 SQL
	err = e.executeSQL(migration.DownSQL)
	if err != nil {
		return fmt.Errorf("failed to execute down migration %s: %w", migration.Version, err)
	}

	// 从 schema_migrations 中删除记录
	_, err = e.db.Exec(
		"DELETE FROM schema_migrations WHERE version = ?",
		migration.Version,
	)
	if err != nil {
		return fmt.Errorf("failed to remove migration %s record: %w", migration.Version, err)
	}

	fmt.Printf("Successfully rolled back migration: %s\n", migration.Version)
	return nil
}

func (e *Executor) executeSQL(sqlStatements string) error {
	// 分割 SQL 语句
	statements := strings.Split(sqlStatements, ";")

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		_, err := e.db.Exec(stmt)
		if err != nil {
			return fmt.Errorf("error executing SQL: %s, error: %w", stmt, err)
		}
	}

	return nil
}

func (e *Executor) GetCurrentVersion() (string, error) {
	if e.dryRun {
		return "0", nil
	}

	var version string
	err := e.db.QueryRow(
		"SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1",
	).Scan(&version)

	if err == sql.ErrNoRows {
		return "0", nil
	} else if err != nil {
		return "", err
	}

	return version, nil
}
