package archive

import (
	"audit-log/database"
	"audit-log/models"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	ArchiveDir       = "./data/archives"
	DefaultThreshold = int64(10 * 1024 * 1024 * 1024) // 10GB
	CheckInterval    = 5 * time.Minute
)

var StorageThreshold int64 = DefaultThreshold

func InitArchive() error {
	return os.MkdirAll(ArchiveDir, 0755)
}

func CheckAndArchive() error {
	dbSize, err := getDBSize()
	if err != nil {
		return err
	}

	if dbSize < StorageThreshold {
		return nil
	}

	return archiveOldLogs()
}

func getDBSize() (int64, error) {
	info, err := os.Stat(database.DBFile)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	return info.Size(), nil
}

func archiveOldLogs() error {
	targetArchiveSize := StorageThreshold / 10
	var totalArchived int64 = 0

	for totalArchived < targetArchiveSize {
		rows, err := database.DB.Query(`
			SELECT id, operator, resource_type, operation, resource_id, before_value, after_value, timestamp
			FROM audit_logs
			WHERE archived = 0
			ORDER BY timestamp ASC
			LIMIT 1000
		`)
		if err != nil {
			return err
		}

		var logs []models.AuditLog
		for rows.Next() {
			var log models.AuditLog
			err := rows.Scan(&log.ID, &log.Operator, &log.ResourceType, &log.Operation, &log.ResourceID, &log.BeforeValue, &log.AfterValue, &log.Timestamp)
			if err != nil {
				rows.Close()
				return err
			}
			logs = append(logs, log)
		}
		rows.Close()

		if len(logs) == 0 {
			break
		}

		archiveFile := filepath.Join(ArchiveDir, fmt.Sprintf("audit_%s_%d.csv", time.Now().Format("20060102"), logs[0].ID))
		if err := writeArchiveFile(archiveFile, logs); err != nil {
			return err
		}

		fileInfo, err := os.Stat(archiveFile)
		if err != nil {
			return err
		}
		totalArchived += fileInfo.Size()

		tx, err := database.DB.Begin()
		if err != nil {
			return err
		}

		for _, log := range logs {
			_, err = tx.Exec(`
				UPDATE audit_logs
				SET archived = 1, archive_file = ?, before_value = '', after_value = ''
				WHERE id = ?
			`, archiveFile, log.ID)
			if err != nil {
				tx.Rollback()
				return err
			}
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}

func writeArchiveFile(filename string, logs []models.AuditLog) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"id", "operator", "resource_type", "operation", "resource_id", "before_value", "after_value", "timestamp"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, log := range logs {
		record := []string{
			strconv.FormatInt(log.ID, 10),
			log.Operator,
			log.ResourceType,
			log.Operation,
			log.ResourceID,
			log.BeforeValue,
			log.AfterValue,
			log.Timestamp.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func ReadArchivedLog(archiveFile string, logID int64) (*models.AuditLog, error) {
	file, err := os.Open(archiveFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("archive file is empty or has no data")
	}

	idStr := strconv.FormatInt(logID, 10)
	for _, record := range records[1:] {
		if record[0] == idStr {
			timestamp, _ := time.Parse(time.RFC3339, record[7])
			return &models.AuditLog{
				ID:           logID,
				Operator:     record[1],
				ResourceType: record[2],
				Operation:    record[3],
				ResourceID:   record[4],
				BeforeValue:  record[5],
				AfterValue:   record[6],
				Timestamp:    timestamp,
				Archived:     true,
				ArchiveFile:  archiveFile,
			}, nil
		}
	}

	return nil, fmt.Errorf("log not found in archive")
}

func StartArchiveChecker() {
	ticker := time.NewTicker(CheckInterval)
	go func() {
		for range ticker.C {
			_ = CheckAndArchive()
		}
	}()
}
