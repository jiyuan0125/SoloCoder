package customfield

import (
	"database/sql"
	"encoding/json"
	"errors"
	"multitenant/internal/database"
	"time"
)

var (
	ErrFieldExists       = errors.New("field already exists")
	ErrFieldHasData      = errors.New("field has data and cannot be deleted")
	ErrInvalidFieldType  = errors.New("invalid field type")
	ValidFieldTypes      = []string{"text", "number", "date", "select"}
)

type CustomField struct {
	ID         int64     `json:"id"`
	TenantID   string    `json:"tenant_id"`
	FieldName  string    `json:"field_name"`
	FieldType  string    `json:"field_type"`
	Options    []string  `json:"options"`
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type FieldValue struct {
	ID        int64     `json:"id"`
	FieldID   int64     `json:"field_id"`
	RecordID  string    `json:"record_id"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func isValidFieldType(ft string) bool {
	for _, t := range ValidFieldTypes {
		if t == ft {
			return true
		}
	}
	return false
}

func (s *Service) Create(f *CustomField) error {
	if !isValidFieldType(f.FieldType) {
		return ErrInvalidFieldType
	}
	now := database.Now()
	f.CreatedAt = now
	f.UpdatedAt = now
	f.Enabled = true
	if f.Options == nil {
		f.Options = []string{}
	}
	options, err := json.Marshal(f.Options)
	if err != nil {
		return err
	}
	stmt := `INSERT INTO custom_fields (tenant_id, field_name, field_type, options, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	res, err := s.db.Exec(stmt, f.TenantID, f.FieldName, f.FieldType, string(options), 1, f.CreatedAt, f.UpdatedAt)
	if err != nil {
		return err
	}
	f.ID, _ = res.LastInsertId()
	return nil
}

func (s *Service) Get(id int64) (*CustomField, error) {
	var optionsStr string
	var enabledInt int
	var f CustomField
	err := s.db.QueryRow(`SELECT id, tenant_id, field_name, field_type, options, enabled, created_at, updated_at FROM custom_fields WHERE id = ?`, id).
		Scan(&f.ID, &f.TenantID, &f.FieldName, &f.FieldType, &optionsStr, &enabledInt, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(optionsStr), &f.Options)
	f.Enabled = enabledInt != 0
	return &f, nil
}

func (s *Service) List(tenantID string, includeDisabled bool) ([]*CustomField, error) {
	query := `SELECT id, tenant_id, field_name, field_type, options, enabled, created_at, updated_at FROM custom_fields WHERE tenant_id = ?`
	args := []interface{}{tenantID}
	if !includeDisabled {
		query += ` AND enabled = 1`
	}
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	fields := []*CustomField{}
	for rows.Next() {
		var optionsStr string
		var enabledInt int
		var f CustomField
		err := rows.Scan(&f.ID, &f.TenantID, &f.FieldName, &f.FieldType, &optionsStr, &enabledInt, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(optionsStr), &f.Options)
		f.Enabled = enabledInt != 0
		fields = append(fields, &f)
	}
	return fields, nil
}

func (s *Service) Update(id int64, updates map[string]interface{}) (*CustomField, error) {
	updates["updated_at"] = database.Now()
	query := "UPDATE custom_fields SET "
	args := []interface{}{}
	first := true
	for k, v := range updates {
		if k == "id" || k == "tenant_id" || k == "created_at" {
			continue
		}
		if !first {
			query += ", "
		}
		if k == "options" {
			opts, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			query += k + " = ?"
			args = append(args, string(opts))
		} else if k == "enabled" {
			query += k + " = ?"
			if b, ok := v.(bool); ok {
				args = append(args, boolToInt(b))
			} else {
				args = append(args, v)
			}
		} else {
			query += k + " = ?"
			args = append(args, v)
		}
		first = false
	}
	query += " WHERE id = ?"
	args = append(args, id)
	_, err := s.db.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *Service) Delete(id int64) error {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM custom_field_values WHERE field_id = ?`, id).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrFieldHasData
	}
	_, err = s.db.Exec(`DELETE FROM custom_fields WHERE id = ?`, id)
	return err
}

func (s *Service) SetValue(fieldID int64, recordID, value string) error {
	now := database.Now()
	var id int64
	err := s.db.QueryRow(`SELECT id FROM custom_field_values WHERE field_id = ? AND record_id = ?`, fieldID, recordID).Scan(&id)
	if err == sql.ErrNoRows {
		_, err := s.db.Exec(`INSERT INTO custom_field_values (field_id, record_id, value, created_at) VALUES (?, ?, ?, ?)`, fieldID, recordID, value, now)
		return err
	}
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`UPDATE custom_field_values SET value = ? WHERE id = ?`, value, id)
	return err
}

func (s *Service) GetValues(recordID string) ([]*FieldValue, error) {
	rows, err := s.db.Query(`SELECT id, field_id, record_id, value, created_at FROM custom_field_values WHERE record_id = ?`, recordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []*FieldValue{}
	for rows.Next() {
		var v FieldValue
		err := rows.Scan(&v.ID, &v.FieldID, &v.RecordID, &v.Value, &v.CreatedAt)
		if err != nil {
			return nil, err
		}
		values = append(values, &v)
	}
	return values, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
