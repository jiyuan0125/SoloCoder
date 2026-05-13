package db

import (
	"database/sql"
	"strings"
	"time"
)

func GetVoteByParticipant(pollID int64, participantID string) (*Vote, error) {
	var vote Vote
	var createdAtStr, updatedAtStr string

	err := DB.QueryRow(`
		SELECT id, poll_id, participant_id, option_ids, created_at, updated_at
		FROM votes WHERE poll_id = ? AND participant_id = ?
	`, pollID, participantID).Scan(
		&vote.ID, &vote.PollID, &vote.ParticipantID, &vote.OptionIDs,
		&createdAtStr, &updatedAtStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if vote.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr); err != nil {
		return nil, err
	}
	if vote.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr); err != nil {
		return nil, err
	}

	return &vote, nil
}

func CreateVote(pollID int64, participantID string, optionIDs []int64) error {
	optionIDsStr := int64SliceToString(optionIDs)
	now := time.Now().Format(time.RFC3339)

	_, err := DB.Exec(`
		INSERT INTO votes (poll_id, participant_id, option_ids, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, pollID, participantID, optionIDsStr, now, now)

	return err
}

func UpdateVote(voteID int64, optionIDs []int64) error {
	optionIDsStr := int64SliceToString(optionIDs)
	now := time.Now().Format(time.RFC3339)

	_, err := DB.Exec(`
		UPDATE votes SET option_ids = ?, updated_at = ?
		WHERE id = ?
	`, optionIDsStr, now, voteID)

	return err
}

func OptionIDsFromString(s string) []int64 {
	if s == "" {
		return []int64{}
	}

	parts := strings.Split(s, ",")
	result := make([]int64, 0, len(parts))
	for _, part := range parts {
		var id int64
		for _, c := range part {
			if c >= '0' && c <= '9' {
				id = id*10 + int64(c-'0')
			}
		}
		if id > 0 {
			result = append(result, id)
		}
	}
	return result
}

func int64SliceToString(ids []int64) string {
	if len(ids) == 0 {
		return ""
	}

	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = itoa(int(id))
	}
	return strings.Join(parts, ",")
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	neg := false
	if n < 0 {
		neg = true
		n = -n
	}

	var result []byte
	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}

	if neg {
		result = append([]byte{'-'}, result...)
	}

	return string(result)
}
