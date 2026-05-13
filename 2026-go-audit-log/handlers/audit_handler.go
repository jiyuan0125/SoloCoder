package handlers

import (
	"audit-log/archive"
	"audit-log/database"
	"audit-log/models"
	"database/sql"
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateAuditLog(c *gin.Context) {
	var req models.CreateAuditLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	result, err := database.DB.Exec(`
		INSERT INTO audit_logs (operator, resource_type, operation, resource_id, before_value, after_value, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, req.Operator, req.ResourceType, req.Operation, req.ResourceID, req.BeforeValue, req.AfterValue, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create audit log"})
		return
	}

	id, _ := result.LastInsertId()
	log := models.AuditLog{
		ID:           id,
		Operator:     req.Operator,
		ResourceType: req.ResourceType,
		Operation:    req.Operation,
		ResourceID:   req.ResourceID,
		BeforeValue:  req.BeforeValue,
		AfterValue:   req.AfterValue,
		Timestamp:    now,
		Archived:     false,
	}

	c.JSON(http.StatusCreated, log)
}

func scanAuditLog(rows *sql.Rows) (*models.AuditLog, error) {
	var log models.AuditLog
	var archived sql.NullBool
	var archiveFile sql.NullString
	
	err := rows.Scan(
		&log.ID, &log.Operator, &log.ResourceType, &log.Operation,
		&log.ResourceID, &log.BeforeValue, &log.AfterValue, &log.Timestamp,
		&archived, &archiveFile,
	)
	
	if err != nil {
		return nil, err
	}
	
	log.Archived = archived.Bool
	if archiveFile.Valid {
		log.ArchiveFile = archiveFile.String
	}
	
	return &log, nil
}

func QueryAuditLogs(c *gin.Context) {
	var req models.QueryAuditLogRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	whereClauses := []string{"1=1"}
	args := []interface{}{}

	if req.Operator != "" {
		whereClauses = append(whereClauses, "operator = ?")
		args = append(args, req.Operator)
	}

	if req.ResourceType != "" {
		whereClauses = append(whereClauses, "resource_type = ?")
		args = append(args, req.ResourceType)
	}

	if !req.StartTime.IsZero() {
		whereClauses = append(whereClauses, "timestamp >= ?")
		args = append(args, req.StartTime)
	}

	if !req.EndTime.IsZero() {
		whereClauses = append(whereClauses, "timestamp <= ?")
		args = append(args, req.EndTime)
	}

	query := `
		SELECT id, operator, resource_type, operation, resource_id, before_value, after_value, timestamp, archived, archive_file
		FROM audit_logs
		WHERE ` + strings.Join(whereClauses, " AND ") + `
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`

	args = append(args, req.PageSize, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query audit logs"})
		return
	}
	defer rows.Close()

	logs := []models.AuditLog{}
	for rows.Next() {
		log, err := scanAuditLog(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan audit log"})
			return
		}
		logs = append(logs, *log)
	}

	if logs == nil {
		logs = []models.AuditLog{}
	}

	c.JSON(http.StatusOK, logs)
}

func GetAuditLogByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid log ID"})
		return
	}

	rows, err := database.DB.Query(`
		SELECT id, operator, resource_type, operation, resource_id, before_value, after_value, timestamp, archived, archive_file
		FROM audit_logs
		WHERE id = ?
	`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get audit log"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		c.JSON(http.StatusNotFound, gin.H{"error": "audit log not found"})
		return
	}

	log, err := scanAuditLog(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan audit log"})
		return
	}

	if log.Archived {
		archivedLog, err := archive.ReadArchivedLog(log.ArchiveFile, log.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read archived log"})
			return
		}
		c.JSON(http.StatusOK, archivedLog)
		return
	}

	c.JSON(http.StatusOK, log)
}

func ExportAuditLogsCSV(c *gin.Context) {
	var req models.QueryAuditLogRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}

	whereClauses := []string{"1=1"}
	args := []interface{}{}

	if req.Operator != "" {
		whereClauses = append(whereClauses, "operator = ?")
		args = append(args, req.Operator)
	}

	if req.ResourceType != "" {
		whereClauses = append(whereClauses, "resource_type = ?")
		args = append(args, req.ResourceType)
	}

	if !req.StartTime.IsZero() {
		whereClauses = append(whereClauses, "timestamp >= ?")
		args = append(args, req.StartTime)
	}

	if !req.EndTime.IsZero() {
		whereClauses = append(whereClauses, "timestamp <= ?")
		args = append(args, req.EndTime)
	}

	query := `
		SELECT id, operator, resource_type, operation, resource_id, before_value, after_value, timestamp, archived, archive_file
		FROM audit_logs
		WHERE ` + strings.Join(whereClauses, " AND ") + `
		ORDER BY timestamp DESC
	`

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query audit logs"})
		return
	}
	defer rows.Close()

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", `attachment; filename="audit_logs_`+time.Now().Format("20060102_150405")+`.csv"`)

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	header := []string{"id", "operator", "resource_type", "operation", "resource_id", "before_value", "after_value", "timestamp", "archived"}
	if err := writer.Write(header); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write CSV"})
		return
	}

	for rows.Next() {
		log, err := scanAuditLog(rows)
		if err != nil {
			continue
		}

		if log.Archived && log.ArchiveFile != "" {
			archivedLog, err := archive.ReadArchivedLog(log.ArchiveFile, log.ID)
			if err == nil {
				log = archivedLog
			}
		}

		record := []string{
			strconv.FormatInt(log.ID, 10),
			log.Operator,
			log.ResourceType,
			log.Operation,
			log.ResourceID,
			log.BeforeValue,
			log.AfterValue,
			log.Timestamp.Format(time.RFC3339),
			strconv.FormatBool(log.Archived),
		}
		if err := writer.Write(record); err != nil {
			continue
		}
	}
}

func MethodNotAllowed(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed - audit logs can only be created and queried"})
}
