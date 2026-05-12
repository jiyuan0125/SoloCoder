package handlers

import (
	"medical-device-manager/database"
	"medical-device-manager/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateMaintenancePlanRequest struct {
	DeviceID      uint   `json:"device_id" binding:"required"`
	Type          string `json:"type" binding:"required"`
	ScheduledDate string `json:"scheduled_date" binding:"required"`
}

func CreateMaintenancePlan(c *gin.Context) {
	var req CreateMaintenancePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scheduledDate, err := parseDate(req.ScheduledDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "计划日期格式错误"})
		return
	}

	var device models.Device
	if err := database.DB.First(&device, req.DeviceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "设备不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if device.Status == models.StatusScrapped {
		c.JSON(http.StatusOK, gin.H{"message": "设备已报废计划已取消"})
		return
	}

	plan := models.MaintenancePlan{
		DeviceID:      req.DeviceID,
		Type:          models.MaintenanceType(req.Type),
		ScheduledDate: scheduledDate,
		Status:        models.MaintenancePlanStatusPending,
	}

	if err := database.DB.Create(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, plan)
}

func GetMaintenancePlans(c *gin.Context) {
	status := c.Query("status")
	deviceID := c.Query("device_id")

	query := database.DB.Preload("Device").Model(&models.MaintenancePlan{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}

	var plans []models.MaintenancePlan
	if err := query.Order("scheduled_date DESC").Find(&plans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	overdueThreshold := now.AddDate(0, 0, -7)
	for i := range plans {
		if plans[i].Status == models.MaintenancePlanStatusPending &&
			plans[i].ScheduledDate.Before(overdueThreshold) {
			plans[i].Status = models.MaintenancePlanStatusOverdue
			database.DB.Model(&plans[i]).Update("status", models.MaintenancePlanStatusOverdue)
		}
	}

	c.JSON(http.StatusOK, plans)
}

type CompleteMaintenancePlanRequest struct {
	Content         string `json:"content" binding:"required"`
	ReplacedParts   string `json:"replaced_parts"`
	DurationMinutes int    `json:"duration_minutes" binding:"required"`
}

func CompleteMaintenancePlan(c *gin.Context) {
	id := c.Param("id")

	var req CompleteMaintenancePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var plan models.MaintenancePlan
	if err := database.DB.First(&plan, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "维护计划不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if plan.Status != models.MaintenancePlanStatusPending && plan.Status != models.MaintenancePlanStatusOverdue {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只能完成待执行或逾期的维护计划"})
		return
	}

	now := time.Now()
	plan.Content = req.Content
	plan.ReplacedParts = req.ReplacedParts
	plan.DurationMinutes = req.DurationMinutes
	plan.Status = models.MaintenancePlanStatusCompleted
	plan.CompletedDate = &now

	if err := database.DB.Save(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, plan)
}

func CancelMaintenancePlan(c *gin.Context) {
	id := c.Param("id")

	var plan models.MaintenancePlan
	if err := database.DB.First(&plan, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "维护计划不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if plan.Status == models.MaintenancePlanStatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "已完成的维护计划不能取消"})
		return
	}

	plan.Status = models.MaintenancePlanStatusCancelled
	if err := database.DB.Save(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, plan)
}

func GenerateMaintenancePlans(c *gin.Context) {
	var devices []models.Device
	if err := database.DB.Where("status != ?", models.StatusScrapped).Find(&devices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	var createdPlans []models.MaintenancePlan

	for _, device := range devices {
		if device.Name == "CT机" {
			nextRoutine := now.AddDate(0, 1, 0)
			routinePlan := models.MaintenancePlan{
				DeviceID:      device.ID,
				Type:          models.MaintenanceTypeRoutine,
				ScheduledDate: nextRoutine,
				Status:        models.MaintenancePlanStatusPending,
			}

			var existingRoutine models.MaintenancePlan
			err := database.DB.Where("device_id = ? AND type = ? AND DATE(scheduled_date) = DATE(?)",
				device.ID, models.MaintenanceTypeRoutine, nextRoutine).First(&existingRoutine).Error
			if err == gorm.ErrRecordNotFound {
				database.DB.Create(&routinePlan)
				createdPlans = append(createdPlans, routinePlan)
			}

			nextDeep := now.AddDate(0, 3, 0)
			deepPlan := models.MaintenancePlan{
				DeviceID:      device.ID,
				Type:          models.MaintenanceTypeDeep,
				ScheduledDate: nextDeep,
				Status:        models.MaintenancePlanStatusPending,
			}

			var existingDeep models.MaintenancePlan
			err = database.DB.Where("device_id = ? AND type = ? AND DATE(scheduled_date) = DATE(?)",
				device.ID, models.MaintenanceTypeDeep, nextDeep).First(&existingDeep).Error
			if err == gorm.ErrRecordNotFound {
				database.DB.Create(&deepPlan)
				createdPlans = append(createdPlans, deepPlan)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"created_count": len(createdPlans),
		"plans":         createdPlans,
	})
}
