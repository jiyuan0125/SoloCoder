package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go-api-doc-gen/internal/models"

	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
}

func New(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	s := &Storage{db: db}
	if err := s.initSchema(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS document_versions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		note TEXT,
		source_dir TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS apis (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		version_id INTEGER NOT NULL,
		module TEXT,
		path TEXT NOT NULL,
		method TEXT NOT NULL,
		handler_name TEXT,
		description TEXT,
		params_json TEXT,
		returns_json TEXT,
		example_request TEXT,
		example_response TEXT,
		is_complete INTEGER DEFAULT 0,
		missing_fields_json TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (version_id) REFERENCES document_versions(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_apis_version ON apis(version_id);
	CREATE INDEX IF NOT EXISTS idx_apis_module ON apis(module);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *Storage) CreateVersion(name, note, sourceDir string) (int64, error) {
	res, err := s.db.Exec(
		"INSERT INTO document_versions (name, note, source_dir) VALUES (?, ?, ?)",
		name, note, sourceDir,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Storage) SaveAPI(versionID int64, api *models.API) error {
	paramsJSON, _ := json.Marshal(api.Params)
	returnsJSON, _ := json.Marshal(api.Returns)
	missingJSON, _ := json.Marshal(api.MissingFields)

	_, err := s.db.Exec(`
		INSERT INTO apis 
		(version_id, module, path, method, handler_name, description, 
		 params_json, returns_json, example_request, example_response, 
		 is_complete, missing_fields_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		versionID, api.Module, api.Path, api.Method, api.HandlerName, api.Description,
		string(paramsJSON), string(returnsJSON), api.Example.Request, api.Example.Response,
		boolToInt(api.IsComplete), string(missingJSON),
	)
	return err
}

func (s *Storage) SaveAPIs(versionID int64, apis []models.API) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i := range apis {
		paramsJSON, _ := json.Marshal(apis[i].Params)
		returnsJSON, _ := json.Marshal(apis[i].Returns)
		missingJSON, _ := json.Marshal(apis[i].MissingFields)

		_, err := tx.Exec(`
			INSERT INTO apis 
			(version_id, module, path, method, handler_name, description, 
			 params_json, returns_json, example_request, example_response, 
			 is_complete, missing_fields_json)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			versionID, apis[i].Module, apis[i].Path, apis[i].Method,
			apis[i].HandlerName, apis[i].Description,
			string(paramsJSON), string(returnsJSON),
			apis[i].Example.Request, apis[i].Example.Response,
			boolToInt(apis[i].IsComplete), string(missingJSON),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) GetVersions() ([]models.DocumentVersion, error) {
	rows, err := s.db.Query(`
		SELECT id, name, note, source_dir, created_at 
		FROM document_versions 
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []models.DocumentVersion
	for rows.Next() {
		var v models.DocumentVersion
		var createdAtStr string
		if err := rows.Scan(&v.ID, &v.Name, &v.Note, &v.SourceDir, &createdAtStr); err != nil {
			return nil, err
		}
		v.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		versions = append(versions, v)
	}
	return versions, nil
}

func (s *Storage) GetVersion(id int64) (*models.DocumentVersion, error) {
	var v models.DocumentVersion
	var createdAtStr string
	err := s.db.QueryRow(`
		SELECT id, name, note, source_dir, created_at 
		FROM document_versions WHERE id = ?`, id,
	).Scan(&v.ID, &v.Name, &v.Note, &v.SourceDir, &createdAtStr)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	v.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	return &v, nil
}

func (s *Storage) GetAPIs(versionID int64) ([]models.API, error) {
	rows, err := s.db.Query(`
		SELECT id, version_id, module, path, method, handler_name, description,
		 params_json, returns_json, example_request, example_response,
		 is_complete, missing_fields_json, created_at
		FROM apis WHERE version_id = ?`, versionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apis []models.API
	for rows.Next() {
		var api models.API
		var paramsJSON, returnsJSON, missingJSON, createdAtStr string
		var isComplete int

		if err := rows.Scan(&api.ID, &api.VersionID, &api.Module, &api.Path,
			&api.Method, &api.HandlerName, &api.Description,
			&paramsJSON, &returnsJSON, &api.Example.Request, &api.Example.Response,
			&isComplete, &missingJSON, &createdAtStr); err != nil {
			return nil, err
		}

		api.IsComplete = isComplete == 1
		api.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		json.Unmarshal([]byte(paramsJSON), &api.Params)
		json.Unmarshal([]byte(returnsJSON), &api.Returns)
		json.Unmarshal([]byte(missingJSON), &api.MissingFields)

		apis = append(apis, api)
	}
	return apis, nil
}

func (s *Storage) GetModules(versionID int64) ([]string, error) {
	rows, err := s.db.Query(
		"SELECT DISTINCT module FROM apis WHERE version_id = ? ORDER BY module", versionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modules []string
	for rows.Next() {
		var m string
		if err := rows.Scan(&m); err != nil {
			return nil, err
		}
		modules = append(modules, m)
	}
	return modules, nil
}

func (s *Storage) GetAPIsByModule(versionID int64, module string) ([]models.API, error) {
	rows, err := s.db.Query(`
		SELECT id, version_id, module, path, method, handler_name, description,
		 params_json, returns_json, example_request, example_response,
		 is_complete, missing_fields_json, created_at
		FROM apis WHERE version_id = ? AND module = ?`, versionID, module,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apis []models.API
	for rows.Next() {
		var api models.API
		var paramsJSON, returnsJSON, missingJSON, createdAtStr string
		var isComplete int

		if err := rows.Scan(&api.ID, &api.VersionID, &api.Module, &api.Path,
			&api.Method, &api.HandlerName, &api.Description,
			&paramsJSON, &returnsJSON, &api.Example.Request, &api.Example.Response,
			&isComplete, &missingJSON, &createdAtStr); err != nil {
			return nil, err
		}

		api.IsComplete = isComplete == 1
		api.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		json.Unmarshal([]byte(paramsJSON), &api.Params)
		json.Unmarshal([]byte(returnsJSON), &api.Returns)
		json.Unmarshal([]byte(missingJSON), &api.MissingFields)

		apis = append(apis, api)
	}
	return apis, nil
}

func (s *Storage) SearchAPIs(versionID int64, keyword string) ([]models.API, error) {
	searchPattern := "%" + keyword + "%"
	rows, err := s.db.Query(`
		SELECT id, version_id, module, path, method, handler_name, description,
		 params_json, returns_json, example_request, example_response,
		 is_complete, missing_fields_json, created_at
		FROM apis 
		WHERE version_id = ? 
		AND (path LIKE ? OR description LIKE ? OR module LIKE ? OR handler_name LIKE ?)`,
		versionID, searchPattern, searchPattern, searchPattern, searchPattern,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apis []models.API
	for rows.Next() {
		var api models.API
		var paramsJSON, returnsJSON, missingJSON, createdAtStr string
		var isComplete int

		if err := rows.Scan(&api.ID, &api.VersionID, &api.Module, &api.Path,
			&api.Method, &api.HandlerName, &api.Description,
			&paramsJSON, &returnsJSON, &api.Example.Request, &api.Example.Response,
			&isComplete, &missingJSON, &createdAtStr); err != nil {
			return nil, err
		}

		api.IsComplete = isComplete == 1
		api.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		json.Unmarshal([]byte(paramsJSON), &api.Params)
		json.Unmarshal([]byte(returnsJSON), &api.Returns)
		json.Unmarshal([]byte(missingJSON), &api.MissingFields)

		apis = append(apis, api)
	}
	return apis, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
