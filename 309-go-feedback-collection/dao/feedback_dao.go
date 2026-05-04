package dao

import (
	"database/sql"
	"time"

	"feedback-collection/database"
	"feedback-collection/models"
)

func CreateFeedback(feedback *models.Feedback) error {
	query := `
	INSERT INTO feedbacks (user_id, feedback_type, rating, description, status, internal_note, processing_note, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	feedback.CreatedAt = now
	feedback.UpdatedAt = now

	result, err := database.DB.Exec(
		query,
		feedback.UserID,
		feedback.FeedbackType,
		feedback.Rating,
		feedback.Description,
		feedback.Status,
		feedback.InternalNote,
		feedback.ProcessingNote,
		feedback.CreatedAt,
		feedback.UpdatedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	feedback.ID = id

	return nil
}

func GetFeedbackByID(id int64) (*models.Feedback, error) {
	query := `
	SELECT id, user_id, feedback_type, rating, description, status, internal_note, processing_note, created_at, updated_at
	FROM feedbacks
	WHERE id = ?
	`

	var feedback models.Feedback
	err := database.DB.QueryRow(query, id).Scan(
		&feedback.ID,
		&feedback.UserID,
		&feedback.FeedbackType,
		&feedback.Rating,
		&feedback.Description,
		&feedback.Status,
		&feedback.InternalNote,
		&feedback.ProcessingNote,
		&feedback.CreatedAt,
		&feedback.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &feedback, nil
}

func CountUserFeedbacksToday(userID string) (int, error) {
	query := `
	SELECT COUNT(*)
	FROM feedbacks
	WHERE user_id = ? AND date(created_at) = date('now')
	`

	var count int
	err := database.DB.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func GetFeedbackList(feedbackType, status string, page, pageSize int) ([]*models.Feedback, int, error) {
	baseQuery := "FROM feedbacks WHERE 1=1"
	countQuery := "SELECT COUNT(*) " + baseQuery
	selectQuery := `SELECT id, user_id, feedback_type, rating, description, status, internal_note, processing_note, created_at, updated_at ` + baseQuery

	var args []interface{}

	if feedbackType != "" {
		baseQuery += " AND feedback_type = ?"
		countQuery += " AND feedback_type = ?"
		selectQuery += " AND feedback_type = ?"
		args = append(args, feedbackType)
	}

	if status != "" {
		baseQuery += " AND status = ?"
		countQuery += " AND status = ?"
		selectQuery += " AND status = ?"
		args = append(args, status)
	}

	var total int
	countQuery = countQuery
	err := database.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	selectQuery += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)

	rows, err := database.DB.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var feedbacks []*models.Feedback
	for rows.Next() {
		var feedback models.Feedback
		err := rows.Scan(
			&feedback.ID,
			&feedback.UserID,
			&feedback.FeedbackType,
			&feedback.Rating,
			&feedback.Description,
			&feedback.Status,
			&feedback.InternalNote,
			&feedback.ProcessingNote,
			&feedback.CreatedAt,
			&feedback.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		feedbacks = append(feedbacks, &feedback)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return feedbacks, total, nil
}

func UpdateFeedbackStatus(feedbackID int64, status models.FeedbackStatus, processingNote string) error {
	query := `
	UPDATE feedbacks
	SET status = ?, processing_note = ?, updated_at = ?
	WHERE id = ?
	`

	_, err := database.DB.Exec(
		query,
		status,
		processingNote,
		time.Now(),
		feedbackID,
	)
	return err
}

func AddInternalNote(feedbackID int64, note string) error {
	query := `
	UPDATE feedbacks
	SET internal_note = ?, updated_at = ?
	WHERE id = ?
	`

	_, err := database.DB.Exec(
		query,
		note,
		time.Now(),
		feedbackID,
	)
	return err
}

func GetStatistics() ([]*models.FeedbackStatistics, error) {
	query := `
	SELECT feedback_type, COUNT(*) as count, AVG(rating) as average_rating
	FROM feedbacks
	GROUP BY feedback_type
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*models.FeedbackStatistics
	for rows.Next() {
		var stat models.FeedbackStatistics
		err := rows.Scan(
			&stat.Type,
			&stat.Count,
			&stat.AverageRating,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, &stat)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}
