package services

import (
	"database/sql"
	"errors"
	"math"
	"time"

	"member-points/database"
	"member-points/models"
)

func RecordConsumption(memberID int64, amount int64) (*models.Consumption, *models.Member, error) {
	if amount <= 0 {
		return nil, nil, errors.New("amount must be positive")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	member, err := getMemberByID(tx, memberID)
	if err != nil {
		return nil, nil, err
	}

	currentYear := time.Now().Year()
	if member.Year != currentYear {
		member.YearlyPoints = 0
		member.Year = currentYear
		_, err = tx.Exec(
			"UPDATE members SET yearly_points = 0, year = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			currentYear, memberID,
		)
		if err != nil {
			return nil, nil, err
		}
	}

	multiplier := GetMultiplier(member.Level)
	basePoints := amount / 100
	earnedPoints := int64(math.Floor(float64(basePoints) * multiplier))

	result, err := tx.Exec(
		`INSERT INTO consumptions (member_id, amount, points_earned, multiplier, processed)
		 VALUES (?, ?, ?, ?, 0)`,
		memberID, amount, earnedPoints, multiplier,
	)
	if err != nil {
		return nil, nil, err
	}

	consumptionID, err := result.LastInsertId()
	if err != nil {
		return nil, nil, err
	}

	_, err = tx.Exec(
		`UPDATE members 
		 SET points = points + ?, yearly_points = yearly_points + ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		earnedPoints, earnedPoints, memberID,
	)
	if err != nil {
		return nil, nil, err
	}

	updatedMember, err := getMemberByID(tx, memberID)
	if err != nil {
		return nil, nil, err
	}

	newLevel := CalculateLevel(updatedMember.YearlyPoints)
	if updatedMember.Level != newLevel {
		err = updateMemberLevel(tx, updatedMember, newLevel)
		if err != nil {
			return nil, nil, err
		}
	}

	err = syncAndMarkConsumption(tx, consumptionID, memberID)
	if err != nil {
		return nil, nil, err
	}

	consumption, err := getConsumptionByID(tx, consumptionID)
	if err != nil {
		return nil, nil, err
	}

	finalMember, err := getMemberByID(tx, memberID)
	if err != nil {
		return nil, nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, nil, err
	}

	return consumption, finalMember, nil
}

func syncAndMarkConsumption(tx *sql.Tx, consumptionID int64, memberID int64) error {
	var consumption models.Consumption
	err := tx.QueryRow(
		`SELECT id, member_id, amount, points_earned, multiplier, processed, created_at 
		 FROM consumptions WHERE id = ? AND processed = 0`,
		consumptionID,
	).Scan(&consumption.ID, &consumption.MemberID, &consumption.Amount, &consumption.PointsEarned,
		&consumption.Multiplier, &consumption.Processed, &consumption.CreatedAt)

	if err != nil {
		return err
	}

	var points int64
	var yearlyPoints int64
	err = tx.QueryRow(
		"SELECT points, yearly_points FROM members WHERE id = ?",
		memberID,
	).Scan(&points, &yearlyPoints)

	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE consumptions SET processed = 1 WHERE id = ?", consumptionID)
	return err
}

func getConsumptionByID(tx *sql.Tx, id int64) (*models.Consumption, error) {
	var consumption models.Consumption
	err := tx.QueryRow(
		`SELECT id, member_id, amount, points_earned, multiplier, processed, created_at 
		 FROM consumptions WHERE id = ?`,
		id,
	).Scan(&consumption.ID, &consumption.MemberID, &consumption.Amount, &consumption.PointsEarned,
		&consumption.Multiplier, &consumption.Processed, &consumption.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("consumption not found")
	}

	if err != nil {
		return nil, err
	}

	return &consumption, nil
}
