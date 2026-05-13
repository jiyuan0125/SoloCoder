package database

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type Report struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	TemplateID     int64     `json:"template_id"`
	Status         string    `json:"status"`
	Data          string    `json:"data"`
	RejectReason   string    `json:"reject_reason,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Template struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	FilePath     string    `json:"file_path"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
}

type Entity struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type EntityDetail struct {
	ID        int64     `json:"id"`
	EntityID  int64     `json:"entity_id"`
	Content   string    `json:"content"`
	ChangeType string    `json:"change_type"`
	CreatedAt time.Time `json:"created_at"`
}

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS templates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			file_path TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS reports (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			template_id INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'draft',
			data TEXT,
			reject_reason TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (template_id) REFERENCES templates(id)
		)`,
		`CREATE TABLE IF NOT EXISTS entities (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS entity_details (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			entity_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			change_type TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (entity_id) REFERENCES entities(id)
		)`,
	}

	for _, q := range queries {
		_, err = db.Exec(q)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

func CreateTemplate(db *sql.DB, name, filePath, desc string) (int64, error) {
	res, err := db.Exec(
		"INSERT INTO templates (name, file_path, description) VALUES (?, ?, ?)",
		name, filePath, desc,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func GetTemplate(db *sql.DB, id int64) (*Template, error) {
	var t Template
	err := db.QueryRow(
		"SELECT id, name, file_path, description, created_at FROM templates WHERE id = ?",
		id,
	).Scan(&t.ID, &t.Name, &t.FilePath, &t.Description, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func ListTemplates(db *sql.DB) ([]*Template, error) {
	rows, err := db.Query("SELECT id, name, file_path, description, created_at FROM templates ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []*Template
	for rows.Next() {
		var t Template
		if err := rows.Scan(&t.ID, &t.Name, &t.FilePath, &t.Description, &t.CreatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, &t)
	}
	return templates, nil
}

func DeleteTemplate(db *sql.DB, id int64) error {
	_, err := db.Exec("DELETE FROM templates WHERE id = ?", id)
	return err
}

func CreateReport(db *sql.DB, name string, templateID int64, data string) (int64, error) {
	res, err := db.Exec(
		"INSERT INTO reports (name, template_id, status, data) VALUES (?, ?, 'draft', ?)",
		name, templateID, data,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func GetReport(db *sql.DB, id int64) (*Report, error) {
	var r Report
	var rejectReason sql.NullString
	err := db.QueryRow(
		"SELECT id, name, template_id, status, data, reject_reason, created_at, updated_at FROM reports WHERE id = ?",
		id,
	).Scan(&r.ID, &r.Name, &r.TemplateID, &r.Status, &r.Data, &rejectReason, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if rejectReason.Valid {
		r.RejectReason = rejectReason.String
	}
	return &r, nil
}

func ListReports(db *sql.DB) ([]*Report, error) {
	rows, err := db.Query("SELECT id, name, template_id, status, data, reject_reason, created_at, updated_at FROM reports ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*Report
	for rows.Next() {
		var r Report
		var rejectReason sql.NullString
		if err := rows.Scan(&r.ID, &r.Name, &r.TemplateID, &r.Status, &r.Data, &rejectReason, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		if rejectReason.Valid {
			r.RejectReason = rejectReason.String
		}
		reports = append(reports, &r)
	}
	return reports, nil
}

func UpdateReportStatus(db *sql.DB, id int64, status, reason string) error {
	_, err := db.Exec(
		"UPDATE reports SET status = ?, reject_reason = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		status, reason, id,
	)
	return err
}

func UpdateReportData(db *sql.DB, id int64, data string) error {
	_, err := db.Exec(
		"UPDATE reports SET data = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		data, id,
	)
	return err
}

func CreateEntity(db *sql.DB, name, desc string) (int64, error) {
	res, err := db.Exec(
		"INSERT INTO entities (name, description) VALUES (?, ?)",
		name, desc,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func GetEntity(db *sql.DB, id int64) (*Entity, error) {
	var e Entity
	err := db.QueryRow(
		"SELECT id, name, description, created_at FROM entities WHERE id = ?",
		id,
	).Scan(&e.ID, &e.Name, &e.Description, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func ListEntities(db *sql.DB) ([]*Entity, error) {
	rows, err := db.Query("SELECT id, name, description, created_at FROM entities ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []*Entity
	for rows.Next() {
		var e Entity
		if err := rows.Scan(&e.ID, &e.Name, &e.Description, &e.CreatedAt); err != nil {
			return nil, err
		}
		entities = append(entities, &e)
	}
	return entities, nil
}

func DeleteEntity(db *sql.DB, id int64) error {
	_, err := db.Exec("DELETE FROM entity_details WHERE entity_id = ?", id)
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM entities WHERE id = ?", id)
	return err
}

func AddEntityDetail(db *sql.DB, entityID int64, content, changeType string) (int64, error) {
	res, err := db.Exec(
		"INSERT INTO entity_details (entity_id, content, change_type) VALUES (?, ?, ?)",
		entityID, content, changeType,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func ListEntityDetails(db *sql.DB, entityID int64) ([]*EntityDetail, error) {
	rows, err := db.Query(
		"SELECT id, entity_id, content, change_type, created_at FROM entity_details WHERE entity_id = ? ORDER BY created_at DESC",
		entityID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var details []*EntityDetail
	for rows.Next() {
		var d EntityDetail
		if err := rows.Scan(&d.ID, &d.EntityID, &d.Content, &d.ChangeType, &d.CreatedAt); err != nil {
			return nil, err
		}
		details = append(details, &d)
	}
	return details, nil
}

func DeleteEntityDetail(db *sql.DB, id int64) error {
	_, err := db.Exec("DELETE FROM entity_details WHERE id = ?", id)
	return err
}
