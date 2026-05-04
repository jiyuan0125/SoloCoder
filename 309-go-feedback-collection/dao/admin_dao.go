package dao

import (
	"database/sql"

	"feedback-collection/database"
	"feedback-collection/models"
)

func GetAdminByUsername(username string) (*models.Admin, error) {
	query := `
	SELECT id, username, password, created_at
	FROM admins
	WHERE username = ?
	`

	var admin models.Admin
	err := database.DB.QueryRow(query, username).Scan(
		&admin.ID,
		&admin.Username,
		&admin.Password,
		&admin.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &admin, nil
}
