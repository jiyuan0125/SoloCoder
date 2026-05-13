package store

import (
	"database/sql"
	"errors"

	"approval-flow/pkg/db"
	"approval-flow/pkg/model"
	"approval-flow/pkg/util"
)

var ErrUserNotFound = errors.New("user not found")

func CreateUser(user *model.User) error {
	if user.ID == "" {
		user.ID = util.NewUUID()
	}
	query := `INSERT INTO users (id, name, token, manager_id) VALUES (?, ?, ?, ?)`
	_, err := db.DB.Exec(query, user.ID, user.Name, user.Token, user.ManagerID)
	return err
}

func GetUserByID(id string) (*model.User, error) {
	query := `SELECT id, name, token, manager_id, created_at FROM users WHERE id = ?`
	row := db.DB.QueryRow(query, id)

	user := &model.User{}
	var managerID sql.NullString
	err := row.Scan(&user.ID, &user.Name, &user.Token, &managerID, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if managerID.Valid {
		user.ManagerID = managerID.String
	}
	return user, nil
}

func GetUserByToken(token string) (*model.User, error) {
	query := `SELECT id, name, token, manager_id, created_at FROM users WHERE token = ?`
	row := db.DB.QueryRow(query, token)

	user := &model.User{}
	var managerID sql.NullString
	err := row.Scan(&user.ID, &user.Name, &user.Token, &managerID, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if managerID.Valid {
		user.ManagerID = managerID.String
	}
	return user, nil
}

func ListUsers() ([]*model.User, error) {
	query := `SELECT id, name, token, manager_id, created_at FROM users`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*model.User{}
	for rows.Next() {
		user := &model.User{}
		var managerID sql.NullString
		err := rows.Scan(&user.ID, &user.Name, &user.Token, &managerID, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		if managerID.Valid {
			user.ManagerID = managerID.String
		}
		users = append(users, user)
	}
	return users, nil
}

func GetManager(userID string) (*model.User, error) {
	user, err := GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	if user.ManagerID == "" {
		return nil, ErrUserNotFound
	}
	return GetUserByID(user.ManagerID)
}
