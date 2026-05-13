package leaderboard

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"leaderboard/internal/database"
)

type Dimension string

const (
	DimensionDay   Dimension = "day"
	DimensionWeek Dimension = "week"
	DimensionAll  Dimension = "all"
)

type PlayerScore struct {
	Rank       int       `json:"rank"`
	PlayerID   string    `json:"player_id"`
	Score      int       `json:"score"`
	SubmittedAt time.Time `json:"submitted_at"`
}

type Service struct {
	mu sync.RWMutex
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetDimensionKey(dimension Dimension, t time.Time) string {
	switch dimension {
	case DimensionDay:
		return t.Format("2006-01-02")
	case DimensionWeek:
		year, week := t.ISOWeek()
		return fmt.Sprintf("%d-W%02d", year, week)
	case DimensionAll:
		return "all"
	default:
		return ""
	}
}

func (s *Service) IsValidDimension(dimension string) bool {
	d := Dimension(dimension)
	return d == DimensionDay || d == DimensionWeek || d == DimensionAll
}

func (s *Service) SubmitScore(ctx context.Context, playerID string, score int) error {
	if score < 0 {
		return fmt.Errorf("score cannot be negative")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()

	if _, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO players (id) VALUES (?)`, playerID); err != nil {
		return fmt.Errorf("failed to insert player: %w", err)
	}

	dimensions := []Dimension{DimensionDay, DimensionWeek, DimensionAll}
	for _, dim := range dimensions {
		key := s.GetDimensionKey(dim, now)
		
		var currentScore int
		err = tx.QueryRowContext(ctx, `SELECT score FROM scores WHERE player_id = ? AND dimension = ? AND dimension_key = ?`,
			playerID, dim, key).Scan(&currentScore)
		
		if err == sql.ErrNoRows {
			if _, err = tx.ExecContext(ctx, `INSERT INTO scores (player_id, dimension, dimension_key, score, submitted_at) VALUES (?, ?, ?, ?, ?)`,
				playerID, dim, key, score, now); err != nil {
				return fmt.Errorf("failed to insert score: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to query current score: %w", err)
		} else {
			if score > currentScore {
				if _, err = tx.ExecContext(ctx, `UPDATE scores SET score = ?, submitted_at = ? WHERE player_id = ? AND dimension = ? AND dimension_key = ?`,
					score, now, playerID, dim, key); err != nil {
					return fmt.Errorf("failed to update score: %w", err)
				}
			}
		}
	}

	var historyScore int
	err = tx.QueryRowContext(ctx, `SELECT highest_score FROM history_scores WHERE player_id = ?`, playerID).Scan(&historyScore)
	if err == sql.ErrNoRows {
		if _, err = tx.ExecContext(ctx, `INSERT INTO history_scores (player_id, highest_score, updated_at) VALUES (?, ?, ?)`,
			playerID, score, now); err != nil {
			return fmt.Errorf("failed to insert history score: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to query history score: %w", err)
	} else {
		if score > historyScore {
			if _, err = tx.ExecContext(ctx, `UPDATE history_scores SET highest_score = ?, updated_at = ? WHERE player_id = ?`,
				score, now, playerID); err != nil {
				return fmt.Errorf("failed to update history score: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Service) GetTopN(ctx context.Context, dimension string, n int) ([]PlayerScore, error) {
	if !s.IsValidDimension(dimension) {
		return nil, fmt.Errorf("invalid dimension")
	}

	if n <= 0 {
		return nil, fmt.Errorf("n must be positive")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.GetDimensionKey(Dimension(dimension), time.Now())

	rows, err := database.DB.QueryContext(ctx, `
		SELECT player_id, score, submitted_at 
		FROM scores 
		WHERE dimension = ? AND dimension_key = ? 
		ORDER BY score DESC, submitted_at ASC 
		LIMIT ?`,
		dimension, key, n)
	if err != nil {
		return nil, fmt.Errorf("failed to query scores: %w", err)
	}
	defer rows.Close()

	var scores []PlayerScore
	rank := 1
	for rows.Next() {
		var ps PlayerScore
		var submittedAtStr string
		if err := rows.Scan(&ps.PlayerID, &ps.Score, &submittedAtStr); err != nil {
			return nil, fmt.Errorf("failed to scan score: %w", err)
		}
		ps.Rank = rank
		ps.SubmittedAt, _ = parseTime(submittedAtStr)
		scores = append(scores, ps)
		rank++
	}

	return scores, nil
}

func (s *Service) GetPlayerRank(ctx context.Context, dimension string, playerID string) (*PlayerScore, error) {
	if !s.IsValidDimension(dimension) {
		return nil, fmt.Errorf("invalid dimension")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.GetDimensionKey(Dimension(dimension), time.Now())

	var score int
	var submittedAtStr string
	err := database.DB.QueryRowContext(ctx, `
		SELECT score, submitted_at FROM scores WHERE player_id = ? AND dimension = ? AND dimension_key = ?`,
		playerID, dimension, key).Scan(&score, &submittedAtStr)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("player not found")
	} else if err != nil {
		return nil, fmt.Errorf("failed to query player score: %w", err)
	}

	submittedAt, err := parseTime(submittedAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse time: %w", err)
	}

	rows, err := database.DB.QueryContext(ctx, `
		SELECT score, submitted_at FROM scores 
		WHERE dimension = ? AND dimension_key = ? 
		ORDER BY score DESC, submitted_at ASC`,
		dimension, key)
	if err != nil {
		return nil, fmt.Errorf("failed to query all scores for ranking: %w", err)
	}
	defer rows.Close()

	rank := 1
	for rows.Next() {
		var sScore int
		var sSubmittedAtStr string
		if err := rows.Scan(&sScore, &sSubmittedAtStr); err != nil {
			return nil, fmt.Errorf("failed to scan score for ranking: %w", err)
		}
		sSubmittedAt, err := parseTime(sSubmittedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse time for ranking: %w", err)
		}
		
		if sScore == score && sSubmittedAt.Equal(submittedAt) {
			break
		}
		rank++
	}

	return &PlayerScore{
		Rank:       rank,
		PlayerID:   playerID,
		Score:      score,
		SubmittedAt: submittedAt,
	}, nil
}

func parseTime(timeStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05-07:00",
		time.RFC3339,
	}
	
	for _, format := range formats {
		t, err := time.Parse(format, timeStr)
		if err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}

func (s *Service) CleanupExpired(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	todayKey := s.GetDimensionKey(DimensionDay, now)
	thisWeekKey := s.GetDimensionKey(DimensionWeek, now)

	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx, `DELETE FROM scores WHERE dimension = ? AND dimension_key != ?`,
		DimensionDay, todayKey); err != nil {
		return fmt.Errorf("failed to delete old day scores: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM scores WHERE dimension = ? AND dimension_key != ?`,
		DimensionWeek, thisWeekKey); err != nil {
		return fmt.Errorf("failed to delete old week scores: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
