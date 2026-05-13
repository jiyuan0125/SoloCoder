package repository

import (
	"database/sql"

	"points-mall/db"
	"points-mall/models"
)

func CreateUser(name string, points int) (*models.User, error) {
	result, err := db.DB.Exec(
		"INSERT INTO users (name, points_balance) VALUES (?, ?)",
		name, points,
	)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return GetUserByID(int(id))
}

func GetUserByID(id int) (*models.User, error) {
	u := &models.User{}
	err := db.DB.QueryRow(
		"SELECT id, name, points_balance, created_at FROM users WHERE id = ?",
		id,
	).Scan(&u.ID, &u.Name, &u.PointsBalance, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func DecreaseUserPoints(userID int, points int) (bool, error) {
	result, err := db.DB.Exec(
		"UPDATE users SET points_balance = points_balance - ? WHERE id = ? AND points_balance >= ?",
		points, userID, points,
	)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func IncreaseUserPoints(userID int, points int) error {
	_, err := db.DB.Exec(
		"UPDATE users SET points_balance = points_balance + ? WHERE id = ?",
		points, userID,
	)
	return err
}
