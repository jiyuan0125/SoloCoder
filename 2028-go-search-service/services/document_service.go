package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"search-service/database"
	"search-service/models"
)

func IndexDocument(req models.DocumentIndexRequest) (*models.Document, error) {
	if strings.TrimSpace(req.Title) == "" {
		return nil, fmt.Errorf("title is required")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`INSERT INTO documents (title, body) VALUES (?, ?)`, req.Title, req.Body)
	if err != nil {
		return nil, err
	}

	docID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	if len(req.CustomFields) > 0 {
		for name, value := range req.CustomFields {
			var valStr string
			switch v := value.(type) {
			case string:
				valStr = v
			default:
				jsonBytes, err := json.Marshal(v)
				if err != nil {
					return nil, err
				}
				valStr = string(jsonBytes)
			}
			_, err = tx.Exec(`INSERT INTO custom_fields (document_id, field_name, field_value) VALUES (?, ?, ?)`, docID, name, valStr)
			if err != nil {
				return nil, err
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	doc, err := GetDocumentByID(int(docID))
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func GetDocumentByID(id int) (*models.Document, error) {
	row := database.DB.QueryRow(`SELECT id, title, body, created_at, updated_at FROM documents WHERE id = ?`, id)

	doc := &models.Document{}
	err := row.Scan(&doc.ID, &doc.Title, &doc.Body, &doc.CreatedAt, &doc.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("document not found")
		}
		return nil, err
	}

	customFields, err := getCustomFields(doc.ID)
	if err != nil {
		return nil, err
	}
	doc.CustomFields = customFields

	return doc, nil
}

func getCustomFields(docID int) (map[string]interface{}, error) {
	rows, err := database.DB.Query(`SELECT field_name, field_value FROM custom_fields WHERE document_id = ?`, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fields := make(map[string]interface{})
	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}

		var val interface{}
		if err := json.Unmarshal([]byte(value), &val); err != nil {
			val = value
		}
		fields[name] = val
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(fields) == 0 {
		return nil, nil
	}

	return fields, nil
}

func ListDocuments() ([]*models.Document, error) {
	rows, err := database.DB.Query(`SELECT id, title, body, created_at, updated_at FROM documents ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	docs := []*models.Document{}
	for rows.Next() {
		doc := &models.Document{}
		if err := rows.Scan(&doc.ID, &doc.Title, &doc.Body, &doc.CreatedAt, &doc.UpdatedAt); err != nil {
			return nil, err
		}

		customFields, err := getCustomFields(doc.ID)
		if err != nil {
			return nil, err
		}
		doc.CustomFields = customFields

		docs = append(docs, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return docs, nil
}
