package db

import (
	"database/sql"
	"time"
)

func CreatePoll(poll *Poll, options []string) (*Poll, []*Option, error) {
	now := time.Now()
	poll.CreatedAt = now
	poll.UpdatedAt = now

	tx, err := DB.Begin()
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO polls (title, is_single_choice, max_choices, is_anonymous, deadline, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, poll.Title, poll.IsSingleChoice, poll.MaxChoices, poll.IsAnonymous, poll.Deadline, poll.CreatedAt, poll.UpdatedAt)
	if err != nil {
		return nil, nil, err
	}

	pollID, err := result.LastInsertId()
	if err != nil {
		return nil, nil, err
	}
	poll.ID = pollID

	createdOptions := make([]*Option, 0, len(options))
	for _, text := range options {
		optResult, err := tx.Exec(`
			INSERT INTO options (poll_id, text, votes)
			VALUES (?, ?, 0)
		`, pollID, text)
		if err != nil {
			return nil, nil, err
		}

		optID, err := optResult.LastInsertId()
		if err != nil {
			return nil, nil, err
		}

		createdOptions = append(createdOptions, &Option{
			ID:     optID,
			PollID: pollID,
			Text:   text,
			Votes:  0,
		})
	}

	if err = tx.Commit(); err != nil {
		return nil, nil, err
	}

	return poll, createdOptions, nil
}

func GetPollByID(pollID int64) (*Poll, error) {
	var poll Poll
	var deadlineStr, createdAtStr, updatedAtStr string

	err := DB.QueryRow(`
		SELECT id, title, is_single_choice, max_choices, is_anonymous, deadline, created_at, updated_at
		FROM polls WHERE id = ?
	`, pollID).Scan(
		&poll.ID, &poll.Title, &poll.IsSingleChoice, &poll.MaxChoices, &poll.IsAnonymous,
		&deadlineStr, &createdAtStr, &updatedAtStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if poll.Deadline, err = time.Parse(time.RFC3339, deadlineStr); err != nil {
		return nil, err
	}
	if poll.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr); err != nil {
		return nil, err
	}
	if poll.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr); err != nil {
		return nil, err
	}

	return &poll, nil
}

func GetOptionsByPollID(pollID int64) ([]*Option, error) {
	rows, err := DB.Query(`
		SELECT id, poll_id, text, votes
		FROM options WHERE poll_id = ?
		ORDER BY id
	`, pollID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	options := make([]*Option, 0)
	for rows.Next() {
		var opt Option
		if err := rows.Scan(&opt.ID, &opt.PollID, &opt.Text, &opt.Votes); err != nil {
			return nil, err
		}
		options = append(options, &opt)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return options, nil
}

func UpdateOptionsVotes(pollID int64, previousOptionIDs []int64, newOptionIDs []int64) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, optID := range previousOptionIDs {
		if _, err := tx.Exec(`
			UPDATE options SET votes = votes - 1
			WHERE poll_id = ? AND id = ? AND votes > 0
		`, pollID, optID); err != nil {
			return err
		}
	}

	for _, optID := range newOptionIDs {
		if _, err := tx.Exec(`
			UPDATE options SET votes = votes + 1
			WHERE poll_id = ? AND id = ?
		`, pollID, optID); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(`
		UPDATE polls SET updated_at = ?
		WHERE id = ?
	`, time.Now().Format(time.RFC3339), pollID); err != nil {
		return err
	}

	return tx.Commit()
}
