package handlers

import (
	"medical-device-manager/config"
	"medical-device-manager/database"
	"medical-device-manager/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateWorkOrderRequest struct {
	DeviceID    uint   `json:"device_id" binding:"required"`
	Description string `json:"description" binding:"required"`
	Priority    string `json:"priority" binding:"required"`
	AssignedTo  string `json:"assigned_to"`
}

func CreateWorkOrder(c *gin.Context) {
	var req CreateWorkOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	reportTime := time.Now()
	responseStartTime := config.CalculateResponseStartTime(reportTime)

	workOrder := models.WorkOrder{
		DeviceID:          req.DeviceID,
		Description:       req.Description,
		Priority:          models.WorkOrderPriority(req.Priority),
		AssignedTo:        req.AssignedTo,
		Status:            models.WorkOrderStatusPending,
		ReportTime:        reportTime,
		ResponseStartTime: &responseStartTime,
	}

	if err := database.DB.Create(&workOrder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, workOrder)
}

func GetWorkOrders(c *gin.Context) {
	status := c.Query("status")
	priority := c.Query("priority")
	deviceID := c.Query("device_id")

	query := database.DB.Preload("Device").Model(&models.WorkOrder{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if priority != "" {
		query = query.Where("priority = ?", priority)
	}
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}

	var workOrders []models.WorkOrder
	if err := query.Order("created_at DESC").Find(&workOrders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, workOrders)
}

func GetWorkOrder(c *gin.Context) {
	id := c.Param("id")

	var workOrder models.WorkOrder
	if err := database.DB.Preload("Device").First(&workOrder, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "工单不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, workOrder)
}

type UpdateWorkOrderStatusRequest struct {
	Status        string `json:"status" binding:"required"`
	RepairContent string `json:"repair_content"`
	ReplacedParts string `json:"replaced_parts"`
}

func UpdateWorkOrderStatus(c *gin.Context) {
	id := c.Param("id")

	var req UpdateWorkOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var workOrder models.WorkOrder
	if err := database.DB.First(&workOrder, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "工单不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newStatus := models.WorkOrderStatus(req.Status)

	validTransitions := map[models.WorkOrderStatus][]models.WorkOrderStatus{
		models.WorkOrderStatusPending:          {models.WorkOrderStatusInProgress},
		models.WorkOrderStatusInProgress:       {models.WorkOrderStatusPendingAcceptance},
		models.WorkOrderStatusPendingAcceptance: {models.WorkOrderStatusCompleted},
		models.WorkOrderStatusCompleted:        {models.WorkOrderStatusClosed},
		models.WorkOrderStatusClosed:           {},
	}

	valid := false
	for _, allowed := range validTransitions[workOrder.Status] {
		if allowed == newStatus {
			valid = true
			break
		}
	}
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的工单状态流转"})
		return
	}

	now := time.Now()

	switch newStatus {
	case models.WorkOrderStatusInProgress:
		workOrder.AcceptedTime = &now
	case models.WorkOrderStatusPendingAcceptance:
		if req.RepairContent == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "维修内容不能为空"})
			return
		}
		workOrder.RepairContent = req.RepairContent
		workOrder.ReplacedParts = req.ReplacedParts
		workOrder.CompletionTime = &now
	case models.WorkOrderStatusCompleted:
		workOrder.AcceptanceTime = &now
	case models.WorkOrderStatusClosed:
		workOrder.CloseTime = &now
	}

	workOrder.Status = newStatus

	if err := database.DB.Save(&workOrder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, workOrder)
}

func EscalateWorkOrders(c *gin.Context) {
	now := time.Now()

	var pendingOrders []models.WorkOrder
	if err := database.DB.Where("status = ?", models.WorkOrderStatusPending).Find(&pendingOrders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	escalatedCount := 0

	for _, order := range pendingOrders {
		if order.ResponseStartTime == nil {
			continue
		}

		hoursElapsed := now.Sub(*order.ResponseStartTime).Hours()
		responseHours := order.Priority.GetResponseHours()

		thresholdHours := 0.0
		switch order.Priority {
		case models.PriorityCritical:
			thresholdHours = 2.0
		case models.PriorityUrgent:
			thresholdHours = 24.0
		case models.PriorityNormal:
			thresholdHours = 72.0
		}

		if hoursElapsed > thresholdHours && !order.Escalated {
			order.Escalated = true
			escalatedCount++
			database.DB.Save(&order)
		}

		if hoursElapsed > responseHours {
			if order.Priority == models.PriorityNormal {
				order.Priority = models.PriorityUrgent
				order.UpgradeCount++
				database.DB.Save(&order)
			} else if order.Priority == models.PriorityUrgent {
				order.Priority = models.PriorityCritical
				order.UpgradeCount++
				database.DB.Save(&order)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"escalated_count": escalatedCount,
		"message":         "已检查并升级逾期工单",
	})
}

func AutoCloseWorkOrders(c *gin.Context) {
	now := time.Now()
	cutoffTime := now.AddDate(0, 0, -7)

	result := database.DB.Model(&models.WorkOrder{}).
		Where("status = ? AND acceptance_time < ?", models.WorkOrderStatusCompleted, cutoffTime).
		Updates(map[string]interface{}{
			"status":     models.WorkOrderStatusClosed,
			"close_time": now,
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"closed_count": result.RowsAffected,
		"message":      "已自动关闭超过7天未验收的工单",
	})
}
