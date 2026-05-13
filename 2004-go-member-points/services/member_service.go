package services

import (
	"database/sql"
	"errors"
	"time"

	"member-points/database"
	"member-points/models"
)

func GetMultiplier(level models.MemberLevel) float64 {
	switch level {
	case models.LevelSilver:
		return 1.2
	case models.LevelGold:
		return 1.5
	case models.LevelDiamond:
		return 2.0
	default:
		return 1.0
	}
}

func CalculateLevel(yearlyPoints int64) models.MemberLevel {
	switch {
	case yearlyPoints >= 10000:
		return models.LevelDiamond
	case yearlyPoints >= 5000:
		return models.LevelGold
	case yearlyPoints >= 1000:
		return models.LevelSilver
	default:
		return models.LevelNormal
	}
}

func CreateMember(name string) (*models.Member, error) {
	currentYear := time.Now().Year()
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		"INSERT INTO members (name, level, points, yearly_points, year) VALUES (?, ?, ?, ?, ?)",
		name, models.LevelNormal, 0, 0, currentYear,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	member, err := getMemberByID(tx, id)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return member, nil
}

func GetMemberByID(id int64) (*models.Member, error) {
	return getMemberByID(database.DB, id)
}

func getMemberByID(db interface{}, id int64) (*models.Member, error) {
	var queryRow func(query string, args ...interface{}) *sql.Row
	switch d := db.(type) {
	case *sql.DB:
		queryRow = d.QueryRow
	case *sql.Tx:
		queryRow = d.QueryRow
	default:
		return nil, errors.New("invalid database type")
	}

	var member models.Member
	err := queryRow(
		"SELECT id, name, level, points, yearly_points, year, created_at, updated_at FROM members WHERE id = ?",
		id,
	).Scan(&member.ID, &member.Name, &member.Level, &member.Points, &member.YearlyPoints, &member.Year, &member.CreatedAt, &member.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("member not found")
	}

	if err != nil {
		return nil, err
	}

	return &member, nil
}

func updateMemberLevel(tx *sql.Tx, member *models.Member, newLevel models.MemberLevel) error {
	if member.Level == newLevel {
		return nil
	}

	member.Level = newLevel
	_, err := tx.Exec(
		"UPDATE members SET level = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		newLevel, member.ID,
	)
	return err
}
