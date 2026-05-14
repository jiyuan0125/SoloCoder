package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"log-aggregator/model"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

const defaultRetentionDays = 30

func Init(dbPath string) error {
	var err error
	dsn := dbPath + "?_loc=auto"
	DB, err = sql.Open("sqlite3", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err = createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	if err = createIndexes(); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	if err = ensureDefaultConfig(); err != nil {
		return fmt.Errorf("failed to ensure default config: %w", err)
	}

	return nil
}

func createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME NOT NULL,
			level TEXT NOT NULL,
			service TEXT NOT NULL,
			message TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS config (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)`,
	}

	for _, query := range queries {
		_, err := DB.Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}

func createIndexes() error {
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_service ON logs(service)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_timestamp_level_service ON logs(timestamp, level, service)`,
	}

	for _, idx := range indexes {
		_, err := DB.Exec(idx)
		if err != nil {
			return err
		}
	}
	return nil
}

func ensureDefaultConfig() error {
	var count int
	err := DB.QueryRow(`SELECT COUNT(*) FROM config WHERE key = 'retention_days'`).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		_, err = DB.Exec(
			`INSERT INTO config (key, value) VALUES (?, ?)`,
			"retention_days",
			fmt.Sprintf("%d", defaultRetentionDays),
		)
		return err
	}
	return nil
}

func InsertLog(entry *model.LogEntry) error {
	query := `INSERT INTO logs (timestamp, level, service, message) VALUES (?, ?, ?, ?)`
	_, err := DB.Exec(query, entry.Timestamp, entry.Level, entry.Service, entry.Message)
	return err
}

func parseSQLiteTime(s string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05-07:00",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}

func QueryLogs(params *model.LogQueryParams) ([]model.AggregatedLogEntry, error) {
	baseQuery := `
		SELECT 
			MIN(timestamp) as first_seen,
			MAX(timestamp) as last_seen,
			level,
			service,
			message,
			COUNT(*) as count
		FROM logs
		WHERE timestamp BETWEEN ? AND ?
	`
	args := []interface{}{params.StartTime, params.EndTime}

	if len(params.Levels) > 0 {
		placeholders := ""
		for i := range params.Levels {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, params.Levels[i])
		}
		baseQuery += fmt.Sprintf(" AND level IN (%s)", placeholders)
	}

	if len(params.Services) > 0 {
		placeholders := ""
		for i := range params.Services {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, params.Services[i])
		}
		baseQuery += fmt.Sprintf(" AND service IN (%s)", placeholders)
	}

	baseQuery += `
		GROUP BY 
			level, 
			service, 
			message,
			(strftime('%s', timestamp) / 300)
	`

	order := "DESC"
	if params.SortOrder == "asc" {
		order = "ASC"
	}
	baseQuery += fmt.Sprintf(" ORDER BY last_seen %s", order)

	rows, err := DB.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []model.AggregatedLogEntry{}
	for rows.Next() {
		var firstSeenStr, lastSeenStr string
		var entry model.AggregatedLogEntry
		err := rows.Scan(&firstSeenStr, &lastSeenStr, &entry.Level, &entry.Service, &entry.Message, &entry.Count)
		if err != nil {
			return nil, err
		}
		entry.FirstSeen, err = parseSQLiteTime(firstSeenStr)
		if err != nil {
			return nil, err
		}
		entry.LastSeen, err = parseSQLiteTime(lastSeenStr)
		if err != nil {
			return nil, err
		}
		entry.Timestamp = entry.LastSeen
		results = append(results, entry)
	}

	if results == nil {
		results = []model.AggregatedLogEntry{}
	}

	return results, nil
}

func GetDailyStats(startDate, endDate time.Time) ([]model.DailyStats, error) {
	query := `
		SELECT
			date(timestamp) as log_date,
			service,
			COUNT(*) as total_count,
			SUM(CASE WHEN level IN ('error', 'ERROR') THEN 1 ELSE 0 END) as error_count,
			SUM(CASE WHEN level IN ('fatal', 'FATAL') THEN 1 ELSE 0 END) as fatal_count
		FROM logs
		WHERE timestamp BETWEEN ? AND ?
		GROUP BY log_date, service
		ORDER BY log_date DESC, service
	`

	rows, err := DB.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []model.DailyStats{}
	for rows.Next() {
		var s model.DailyStats
		err := rows.Scan(&s.Date, &s.Service, &s.TotalCount, &s.ErrorCount, &s.FatalCount)
		if err != nil {
			return nil, err
		}
		if s.TotalCount > 0 {
			s.ErrorRate = float64(s.ErrorCount+s.FatalCount) / float64(s.TotalCount) * 100
		}
		stats = append(stats, s)
	}

	if stats == nil {
		stats = []model.DailyStats{}
	}

	return stats, nil
}

func GetRetentionDays() (int, error) {
	var value string
	err := DB.QueryRow(`SELECT value FROM config WHERE key = 'retention_days'`).Scan(&value)
	if err != nil {
		return 0, err
	}
	var days int
	fmt.Sscanf(value, "%d", &days)
	return days, nil
}

func SetRetentionDays(days int) error {
	_, err := DB.Exec(
		`INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)`,
		"retention_days",
		fmt.Sprintf("%d", days),
	)
	return err
}

func CleanOldLogs(retentionDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	result, err := DB.Exec(`DELETE FROM logs WHERE timestamp < ?`, cutoffDate)
	if err != nil {
		return err
	}
	deleted, _ := result.RowsAffected()
	if deleted > 0 {
		log.Printf("Cleaned %d old logs", deleted)
	}
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
