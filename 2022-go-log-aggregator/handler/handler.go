package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"log-aggregator/db"
	"log-aggregator/model"
)

func SubmitLog(c *gin.Context) {
	var req model.LogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "invalid_request",
			"detail": "无法解析请求体",
		})
		return
	}

	missingFields := req.Validate()
	if len(missingFields) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":          "missing_fields",
			"missing_fields": missingFields,
		})
		return
	}

	entry := req.ToLogEntry()
	if err := db.InsertLog(entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "database_error",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"id":     entry.ID,
	})
}

func QueryLogs(c *gin.Context) {
	params := &model.LogQueryParams{
		SortOrder: "desc",
	}

	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")
	if startTimeStr == "" || endTimeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "missing_query_params",
			"detail": "start_time 和 end_time 是必需的查询参数",
		})
		return
	}

	var err error
	params.StartTime, err = time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "invalid_time_format",
			"detail": "start_time 格式错误，应为 RFC3339 格式",
		})
		return
	}

	params.EndTime, err = time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "invalid_time_format",
			"detail": "end_time 格式错误，应为 RFC3339 格式",
		})
		return
	}

	if params.StartTime.After(params.EndTime) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "invalid_time_range",
			"detail": "start_time 不能晚于 end_time",
		})
		return
	}

	if levels := c.Query("levels"); levels != "" {
		params.Levels = strings.Split(levels, ",")
	}

	if services := c.Query("services"); services != "" {
		params.Services = strings.Split(services, ",")
	}

	if sort := c.Query("sort"); sort != "" {
		if sort == "asc" || sort == "desc" {
			params.SortOrder = sort
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  "invalid_sort_order",
				"detail": "sort 只能是 'asc' 或 'desc'",
			})
			return
		}
	}

	logs, err := db.QueryLogs(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "database_error",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, logs)
}

func GetDailyStats(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr == "" {
		startDate = time.Now().AddDate(0, 0, -7)
	} else {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  "invalid_date_format",
				"detail": "start_date 格式错误，应为 YYYY-MM-DD 格式",
			})
			return
		}
	}

	if endDateStr == "" {
		endDate = time.Now()
	} else {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  "invalid_date_format",
				"detail": "end_date 格式错误，应为 YYYY-MM-DD 格式",
			})
			return
		}
		endDate = endDate.Add(24*time.Hour - time.Nanosecond)
	}

	if startDate.After(endDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "invalid_date_range",
			"detail": "start_date 不能晚于 end_date",
		})
		return
	}

	stats, err := db.GetDailyStats(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "database_error",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func GetRetentionConfig(c *gin.Context) {
	retentionDays, err := db.GetRetentionDays()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "database_error",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.RetentionConfig{
		RetentionDays: retentionDays,
	})
}

func UpdateRetentionConfig(c *gin.Context) {
	var req model.RetentionConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "invalid_request",
			"detail": "无法解析请求体",
		})
		return
	}

	if req.RetentionDays < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "invalid_retention_days",
			"detail": "retention_days 必须大于 0",
		})
		return
	}

	if err := db.SetRetentionDays(req.RetentionDays); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "database_error",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":          "success",
		"retention_days":  req.RetentionDays,
	})
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func parseDateQueryParam(param string, defaultValue time.Time) (time.Time, error) {
	if param == "" {
		return defaultValue, nil
	}
	return time.Parse("2006-01-02", param)
}

func parseIntQueryParam(param string, defaultValue int) (int, error) {
	if param == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(param)
}
