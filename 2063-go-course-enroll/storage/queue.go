package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func (s *SQLiteStore) AddToQueue(ctx context.Context, tx Tx, item *WaitQueueItem) (int64, error) {
	stx := tx.(*sqliteTx).tx

	item.CreatedAt = time.Now()

	row := stx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(position), 0) + 1 FROM wait_queue WHERE course_id = ?
	`, item.CourseID)

	var position int
	if err := row.Scan(&position); err != nil {
		return 0, fmt.Errorf("get queue position: %w", err)
	}
	item.Position = position

	result, err := stx.ExecContext(ctx, `
		INSERT INTO wait_queue (course_id, student_id, position, created_at)
		VALUES (?, ?, ?, ?)
	`,
		item.CourseID,
		item.StudentID,
		item.Position,
		timeToStr(item.CreatedAt),
	)
	if err != nil {
		return 0, fmt.Errorf("add to queue: %w", err)
	}

	return result.LastInsertId()
}

func (s *SQLiteStore) RemoveFromQueue(ctx context.Context, tx Tx, courseID, studentID int64) error {
	stx := tx.(*sqliteTx).tx

	_, err := stx.ExecContext(ctx, `
		DELETE FROM wait_queue WHERE course_id = ? AND student_id = ?
	`, courseID, studentID)
	if err != nil {
		return fmt.Errorf("remove from queue: %w", err)
	}

	_, err = stx.ExecContext(ctx, `
		UPDATE wait_queue SET position = position - 1 WHERE course_id = ? AND position > (
			SELECT position FROM wait_queue WHERE course_id = ? AND student_id = ?
		)
	`, courseID, courseID, studentID)

	return nil
}

func (s *SQLiteStore) GetQueuePosition(ctx context.Context, courseID, studentID int64) (int, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT position FROM wait_queue WHERE course_id = ? AND student_id = ?
	`, courseID, studentID)

	var position int
	err := row.Scan(&position)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("get queue position: %w", err)
	}
	return position, nil
}

func (s *SQLiteStore) ListQueueItems(ctx context.Context, courseID int64) ([]*WaitQueueItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, course_id, student_id, position, created_at
		FROM wait_queue WHERE course_id = ? ORDER BY position ASC
	`, courseID)
	if err != nil {
		return nil, fmt.Errorf("list queue items: %w", err)
	}
	defer rows.Close()

	var items []*WaitQueueItem
	for rows.Next() {
		item := &WaitQueueItem{}
		var createdAt string

		err := rows.Scan(&item.ID, &item.CourseID, &item.StudentID, &item.Position, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("scan queue item: %w", err)
		}
		item.CreatedAt = strToTime(createdAt)
		items = append(items, item)
	}

	return items, nil
}

func (s *SQLiteStore) GetQueueLength(ctx context.Context, tx Tx, courseID int64) (int, error) {
	stx := tx.(*sqliteTx).tx

	row := stx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM wait_queue WHERE course_id = ?
	`, courseID)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("get queue length: %w", err)
	}
	return count, nil
}

func (s *SQLiteStore) PopQueue(ctx context.Context, tx Tx, courseID int64) (*WaitQueueItem, error) {
	stx := tx.(*sqliteTx).tx

	row := stx.QueryRowContext(ctx, `
		SELECT id, course_id, student_id, position, created_at
		FROM wait_queue WHERE course_id = ? ORDER BY position ASC LIMIT 1
	`, courseID)

	item := &WaitQueueItem{}
	var createdAt string

	err := row.Scan(&item.ID, &item.CourseID, &item.StudentID, &item.Position, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("pop queue: %w", err)
	}
	item.CreatedAt = strToTime(createdAt)

	_, err = stx.ExecContext(ctx, `
		DELETE FROM wait_queue WHERE id = ?
	`, item.ID)
	if err != nil {
		return nil, fmt.Errorf("delete from queue: %w", err)
	}

	_, err = stx.ExecContext(ctx, `
		UPDATE wait_queue SET position = position - 1 WHERE course_id = ? AND position > ?
	`, courseID, item.Position)
	if err != nil {
		return nil, fmt.Errorf("reorder queue: %w", err)
	}

	return item, nil
}
