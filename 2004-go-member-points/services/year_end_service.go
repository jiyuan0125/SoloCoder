package services

import (
	"database/sql"
	"errors"
	"time"

	"member-points/database"
	"member-points/models"
)

func ProcessYearEnd(year int) ([]*models.YearEndRecord, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	members, err := getMembersForYearEnd(tx, year)
	if err != nil {
		return nil, err
	}

	var records []*models.YearEndRecord
	for _, member := range members {
		record, err := processMemberYearEnd(tx, member, year)
		if err != nil {
			return nil, err
		}
		if record != nil {
			records = append(records, record)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return records, nil
}

func getMembersForYearEnd(tx *sql.Tx, year int) ([]*models.Member, error) {
	rows, err := tx.Query(
		`SELECT id, name, level, points, yearly_points, year, created_at, updated_at
		 FROM members WHERE year = ?`,
		year,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.Member
	for rows.Next() {
		var m models.Member
		if err := rows.Scan(&m.ID, &m.Name, &m.Level, &m.Points, &m.YearlyPoints, &m.Year, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		members = append(members, &m)
	}

	return members, nil
}

func processMemberYearEnd(tx *sql.Tx, member *models.Member, year int) (*models.YearEndRecord, error) {
	existing, err := getYearEndRecord(tx, member.ID, year)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if existing != nil {
		return nil, nil
	}

	newLevel := CalculateLevel(member.YearlyPoints)
	oldLevel := member.Level

	result, err := tx.Exec(
		`INSERT INTO year_end_records (member_id, year, total_points, old_level, new_level, processed)
		 VALUES (?, ?, ?, ?, ?, 0)`,
		member.ID, year, member.YearlyPoints, oldLevel, newLevel,
	)
	if err != nil {
		return nil, err
	}

	recordID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(
		`UPDATE members 
		 SET points = 0, yearly_points = 0, year = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		year+1, member.ID,
	)
	if err != nil {
		return nil, err
	}

	if oldLevel != newLevel {
		_, err = tx.Exec(
			"UPDATE members SET level = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			newLevel, member.ID,
		)
		if err != nil {
			return nil, err
		}
	}

	err = syncAndMarkYearEnd(tx, recordID, member.ID, year)
	if err != nil {
		return nil, err
	}

	record, err := getYearEndRecordByID(tx, recordID)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func getYearEndRecord(tx *sql.Tx, memberID int64, year int) (*models.YearEndRecord, error) {
	var record models.YearEndRecord
	err := tx.QueryRow(
		`SELECT id, member_id, year, total_points, old_level, new_level, processed, created_at
		 FROM year_end_records WHERE member_id = ? AND year = ?`,
		memberID, year,
	).Scan(&record.ID, &record.MemberID, &record.Year, &record.TotalPoints, &record.OldLevel, &record.NewLevel, &record.Processed, &record.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return &record, nil
}

func syncAndMarkYearEnd(tx *sql.Tx, recordID, memberID int64, year int) error {
	var record models.YearEndRecord
	err := tx.QueryRow(
		`SELECT id, member_id, year, total_points, old_level, new_level, processed, created_at
		 FROM year_end_records WHERE id = ? AND processed = 0`,
		recordID,
	).Scan(&record.ID, &record.MemberID, &record.Year, &record.TotalPoints, &record.OldLevel, &record.NewLevel, &record.Processed, &record.CreatedAt)

	if err != nil {
		return err
	}

	var points, yearlyPoints int64
	var currentYear int
	err = tx.QueryRow(
		"SELECT points, yearly_points, year FROM members WHERE id = ?",
		memberID,
	).Scan(&points, &yearlyPoints, &currentYear)

	if err != nil {
		return err
	}

	if points != 0 || yearlyPoints != 0 || currentYear != year+1 {
		return errors.New("year end processing verification failed")
	}

	_, err = tx.Exec("UPDATE year_end_records SET processed = 1 WHERE id = ?", recordID)
	return err
}

func getYearEndRecordByID(tx *sql.Tx, id int64) (*models.YearEndRecord, error) {
	var record models.YearEndRecord
	err := tx.QueryRow(
		`SELECT id, member_id, year, total_points, old_level, new_level, processed, created_at
		 FROM year_end_records WHERE id = ?`,
		id,
	).Scan(&record.ID, &record.MemberID, &record.Year, &record.TotalPoints, &record.OldLevel, &record.NewLevel, &record.Processed, &record.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("year end record not found")
	}

	if err != nil {
		return nil, err
	}

	return &record, nil
}

func AutoProcessYearEndIfNeeded() error {
	now := time.Now()
	if now.Month() == time.December && now.Day() == 31 {
		_, err := ProcessYearEnd(now.Year())
		return err
	}
	return nil
}
