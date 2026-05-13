package handlers

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"data-export/config"
	"data-export/database"
	"data-export/export"
	"data-export/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SubmitExport(c *gin.Context) {
	var req models.ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	format := strings.ToLower(req.Format)
	if format != config.FormatCSV && format != config.FormatExcel {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件格式，仅支持 csv 和 excel"})
		return
	}
	req.Format = format

	if !database.TableExists(req.TableName) {
		c.JSON(http.StatusNotFound, gin.H{"error": "数据表不存在"})
		return
	}

	validCols, err := database.GetTableColumns(req.TableName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取表结构失败"})
		return
	}

	for _, f := range req.Fields {
		found := false
		for _, vc := range validCols {
			if strings.EqualFold(f, vc) {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusBadRequest, gin.H{"error": "字段不存在: " + f})
			return
		}
	}

	task, err := export.SubmitTask(req)
	if err != nil {
		var dupErr *export.DuplicateTaskError
		if errors.As(err, &dupErr) {
			c.JSON(http.StatusConflict, gin.H{"error": dupErr.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建任务失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "导出任务已提交",
		"task_id": task.ID,
	})
}

func GetProgress(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	progress, err := export.GetTaskProgress(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, progress)
}

func DownloadFile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	task, err := export.GetTaskFile(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if task.Status != config.TaskStatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "导出任务尚未完成"})
		return
	}

	if task.FilePath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件已过期或不存在"})
		return
	}

	if _, err := os.Stat(task.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件已过期或不存在"})
		return
	}

	fileName := task.FileName
	if fileName == "" {
		fileName = filepath.Base(task.FilePath)
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "application/octet-stream")
	c.File(task.FilePath)
}

func ListTables(c *gin.Context) {
	var tables []string
	rows, err := database.DB.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' AND name NOT LIKE 'export_tasks'").Rows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		rows.Scan(&name)
		tables = append(tables, name)
	}

	c.JSON(http.StatusOK, gin.H{"tables": tables})
}

func GetTableInfo(c *gin.Context) {
	tableName := c.Param("table")
	if !database.TableExists(tableName) {
		c.JSON(http.StatusNotFound, gin.H{"error": "数据表不存在"})
		return
	}

	columns, err := database.GetTableColumns(tableName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"table":   tableName,
		"columns": columns,
	})
}
