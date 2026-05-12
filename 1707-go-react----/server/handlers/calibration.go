package handlers

import (
	"medical-device-manager/database"
	"medical-device-manager/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateCalibrationAgencyRequest struct {
	Name                string `json:"name" binding:"required"`
	CertificationNumber string `json:"certification_number" binding:"required"`
	ContactInfo         string `json:"contact_info" binding:"required"`
}

func CreateCalibrationAgency(c *gin.Context) {
	var req CreateCalibrationAgencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agency := models.CalibrationAgency{
		Name:                req.Name,
		CertificationNumber: req.CertificationNumber,
		ContactInfo:         req.ContactInfo,
	}

	if err := database.DB.Create(&agency).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, agency)
}

func GetCalibrationAgencies(c *gin.Context) {
	var agencies []models.CalibrationAgency
	if err := database.DB.Find(&agencies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, agencies)
}

type CreateCalibrationRecordRequest struct {
	DeviceID            uint   `json:"device_id" binding:"required"`
	AgencyID            uint   `json:"agency_id" binding:"required"`
	CalibrationDate     string `json:"calibration_date" binding:"required"`
	Result              string `json:"result"`
	CertificateNumber   string `json:"certificate_number"`
	LimitedFunctions    string `json:"limited_functions"`
}

func CreateCalibrationRecord(c *gin.Context) {
	var req CreateCalibrationRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	calibrationDate, err := parseDate(req.CalibrationDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "校准日期格式错误"})
		return
	}

	record := models.CalibrationRecord{
		DeviceID:          req.DeviceID,
		AgencyID:          req.AgencyID,
		CalibrationDate:   calibrationDate,
		CertificateNumber: req.CertificateNumber,
		LimitedFunctions:  req.LimitedFunctions,
		Status:            models.CalibrationStatusPending,
	}

	if req.Result != "" {
		record.Result = models.CalibrationResult(req.Result)
	}

	if err := database.DB.Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, record)
}

func GetCalibrationRecords(c *gin.Context) {
	deviceID := c.Query("device_id")
	status := c.Query("status")

	query := database.DB.Preload("Device").Preload("Agency").Model(&models.CalibrationRecord{})

	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var records []models.CalibrationRecord
	if err := query.Order("calibration_date DESC").Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, records)
}

type UpdateCalibrationStatusRequest struct {
	Status            string `json:"status" binding:"required"`
	Result            string `json:"result"`
	LimitedFunctions  string `json:"limited_functions"`
	CertificateNumber string `json:"certificate_number"`
}

func UpdateCalibrationStatus(c *gin.Context) {
	id := c.Param("id")

	var req UpdateCalibrationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var record models.CalibrationRecord
	if err := database.DB.First(&record, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "校准记录不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newStatus := models.CalibrationStatus(req.Status)

	validTransitions := map[models.CalibrationStatus][]models.CalibrationStatus{
		models.CalibrationStatusPending:    {models.CalibrationStatusInProgress},
		models.CalibrationStatusInProgress: {models.CalibrationStatusCompleted},
		models.CalibrationStatusCompleted:  {},
	}

	valid := false
	for _, allowed := range validTransitions[record.Status] {
		if allowed == newStatus {
			valid = true
			break
		}
	}
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的状态流转"})
		return
	}

	record.Status = newStatus

	if newStatus == models.CalibrationStatusCompleted {
		if req.Result == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "完成校准需要指定结果"})
			return
		}
		record.Result = models.CalibrationResult(req.Result)
		record.LimitedFunctions = req.LimitedFunctions
		record.CertificateNumber = req.CertificateNumber

		var device models.Device
		if err := database.DB.First(&device, record.DeviceID).Error; err == nil {
			now := time.Now()
			record.CalibrationDate = now
			nextCalibrationDate := now.Add(device.Category.GetCalibrationInterval())
			record.NextCalibrationDate = nextCalibrationDate
			device.LastCalibrationDate = &now
			device.NextCalibrationDate = &nextCalibrationDate

			if record.Result == models.CalibrationResultFail {
				device.Status = models.StatusDisabled

				autoWorkOrder := models.WorkOrder{
					DeviceID:    device.ID,
					Description: "设备校准不合格，需要维修",
					Priority:    models.PriorityUrgent,
					Status:      models.WorkOrderStatusPending,
					ReportTime:  time.Now(),
				}
				database.DB.Create(&autoWorkOrder)
			}

			database.DB.Save(&device)
		}
	}

	if err := database.DB.Save(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"record": record,
	}

	if record.Result == models.CalibrationResultFail {
		response["message"] = "设备校准不合格，已标记为停用并自动生成维修工单"
	} else if record.Result == models.CalibrationResultConditional {
		response["message"] = "设备条件合格，已记录受限功能"
	}

	c.JSON(http.StatusOK, response)
}

func GetCalibrationReminders(c *gin.Context) {
	now := time.Now()
	soon := now.AddDate(0, 0, 30)

	var devices []models.Device
	if err := database.DB.Where(
		"next_calibration_date IS NOT NULL AND next_calibration_date <= ? AND status != ?",
		soon, models.StatusScrapped,
	).Find(&devices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type Reminder struct {
		DeviceID   uint      `json:"device_id"`
		DeviceName string    `json:"device_name"`
		AssetNumber string   `json:"asset_number"`
		DueDate    time.Time `json:"due_date"`
		IsOverdue  bool      `json:"is_overdue"`
		DaysLeft   int       `json:"days_left"`
	}

	var reminders []Reminder
	for _, d := range devices {
		if d.NextCalibrationDate == nil {
			continue
		}
		daysLeft := int(d.NextCalibrationDate.Sub(now).Hours() / 24)
		reminders = append(reminders, Reminder{
			DeviceID:   d.ID,
			DeviceName: d.Name,
			AssetNumber: d.AssetNumber,
			DueDate:    *d.NextCalibrationDate,
			IsOverdue:  daysLeft < 0,
			DaysLeft:   daysLeft,
		})
	}

	c.JSON(http.StatusOK, reminders)
}
