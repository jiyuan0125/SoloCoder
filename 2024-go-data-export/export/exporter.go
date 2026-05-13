package export

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"data-export/config"
	"data-export/database"
	"data-export/models"
	"data-export/utils"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

func ExecuteExport(task *models.ExportTask) {
	database.DB.Model(task).Updates(map[string]interface{}{
		"status":   config.TaskStatusProcessing,
		"progress": 0,
	})

	fields := strings.Split(task.Fields, ",")

	fileName := generateFileName(task)
	filePath := fmt.Sprintf("%s/%s", config.ExportDir, fileName)

	var err error
	if task.Format == config.FormatCSV {
		err = exportCSV(task, fields, filePath)
	} else {
		err = exportExcel(task, fields, filePath)
	}

	if err != nil {
		database.DB.Model(task).Updates(map[string]interface{}{
			"status":        config.TaskStatusFailed,
			"error_message": err.Error(),
		})
		log.Printf("Export task %d failed: %v", task.ID, err)
		return
	}

	database.DB.Model(task).Updates(map[string]interface{}{
		"status":   config.TaskStatusCompleted,
		"progress": 100,
		"file_path": filePath,
		"file_name": fileName,
	})

	log.Printf("Export task %d completed: %s", task.ID, fileName)
	notifyUser(task)
}

func generateFileName(task *models.ExportTask) string {
	timestamp := time.Now().Format("20060102150405")
	ext := ".csv"
	if task.Format == config.FormatExcel {
		ext = ".xlsx"
	}
	return fmt.Sprintf("%s_%s_%s%s", task.TableName, task.UserID, timestamp, ext)
}

func exportCSV(task *models.ExportTask, fields []string, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write(fields)

	query := "SELECT " + strings.Join(fields, ", ") + " FROM " + task.TableName
	if task.Filter != "" {
		query += " WHERE " + task.Filter
	}

	rows, err := database.DB.Raw(query).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	total, _ := database.CountRows(task.TableName, task.Filter)
	var processed int64 = 0

	columns, _ := rows.Columns()
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		rows.Scan(valuePtrs...)

		record := make([]string, len(fields))
		for i, field := range fields {
			for j, col := range columns {
				if strings.EqualFold(field, col) {
					record[i] = formatValue(values[j])
					record[i] = utils.MaskValue(field, record[i])
					break
				}
			}
		}

		writer.Write(record)
		processed++

		if total > 0 {
			progress := int(float64(processed) / float64(total) * 100)
			if progress > 100 {
				progress = 100
			}
			database.DB.Model(task).Update("progress", progress)
		}
	}

	return rows.Err()
}

func exportExcel(task *models.ExportTask, fields []string, filePath string) error {
	f := excelize.NewFile()
	sheetName := "Sheet1"

	for i, field := range fields {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, field)
	}

	query := "SELECT " + strings.Join(fields, ", ") + " FROM " + task.TableName
	if task.Filter != "" {
		query += " WHERE " + task.Filter
	}

	rows, err := database.DB.Raw(query).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	total, _ := database.CountRows(task.TableName, task.Filter)
	var processed int64 = 0
	rowNum := 2

	columns, _ := rows.Columns()
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		rows.Scan(valuePtrs...)

		for i, field := range fields {
			for j, col := range columns {
				if strings.EqualFold(field, col) {
					cell, _ := excelize.CoordinatesToCellName(i+1, rowNum)
					val := formatValue(values[j])
					val = utils.MaskValue(field, val)
					f.SetCellValue(sheetName, cell, val)
					break
				}
			}
		}

		rowNum++
		processed++

		if total > 0 {
			progress := int(float64(processed) / float64(total) * 100)
			if progress > 100 {
				progress = 100
			}
			database.DB.Model(task).Update("progress", progress)
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return f.SaveAs(filePath)
}

func formatValue(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case []byte:
		return string(val)
	case time.Time:
		return val.Format("2006-01-02 15:04:05")
	case sql.NullTime:
		if val.Valid {
			return val.Time.Format("2006-01-02 15:04:05")
		}
		return ""
	case sql.NullString:
		if val.Valid {
			return val.String
		}
		return ""
	case sql.NullInt64:
		if val.Valid {
			return strconv.FormatInt(val.Int64, 10)
		}
		return ""
	case sql.NullFloat64:
		if val.Valid {
			return strconv.FormatFloat(val.Float64, 'f', -1, 64)
		}
		return ""
	case gorm.DeletedAt:
		if val.Valid {
			return val.Time.Format("2006-01-02 15:04:05")
		}
		return ""
	default:
		return fmt.Sprintf("%v", val)
	}
}

func notifyUser(task *models.ExportTask) {
	log.Printf("[Notification] User %s: Your export task #%d is ready. File: %s",
		task.UserID, task.ID, task.FileName)
}
