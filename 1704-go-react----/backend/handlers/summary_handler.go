package handlers

import (
	"net/http"
	"rehab-system/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type SummaryHandler struct {
	service *services.SummaryService
}

func NewSummaryHandler() *SummaryHandler {
	return &SummaryHandler{service: services.NewSummaryService()}
}

func (h *SummaryHandler) GetPatientSummary(c *gin.Context) {
	patientID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")

	var startDate, endDate time.Time
	if startDateStr != "" {
		parsed, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "开始日期格式错误"})
			return
		}
		startDate = parsed
	}
	if endDateStr != "" {
		parsed, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "结束日期格式错误"})
			return
		}
		endDate = parsed
	}

	summary, err := h.service.GetPatientSummary(uint(patientID), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if summary == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "没有找到患者或康复计划"})
		return
	}
	c.JSON(http.StatusOK, summary)
}
