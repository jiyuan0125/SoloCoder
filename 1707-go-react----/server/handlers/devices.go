package handlers

import (
	"encoding/csv"
	"medical-device-manager/database"
	"medical-device-manager/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateDeviceRequest struct {
	AssetNumber        string  `json:"asset_number" binding:"required"`
	Name               string  `json:"name" binding:"required"`
	BrandModel         string  `json:"brand_model" binding:"required"`
	SerialNumber       string  `json:"serial_number" binding:"required"`
	Category           string  `json:"category" binding:"required"`
	Department         string  `json:"department" binding:"required"`
	Location           string  `json:"location" binding:"required"`
	PurchaseDate       string  `json:"purchase_date" binding:"required"`
	PurchasePrice      float64 `json:"purchase_price" binding:"required"`
	WarrantyExpiryDate string  `json:"warranty_expiry_date"`
}

func parseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}
	return time.Parse("2006-01-02", dateStr)
}

func CreateDevice(c *gin.Context) {
	var req CreateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingDevice models.Device
	if err := database.DB.Where("asset_number = ?", req.AssetNumber).First(&existingDevice).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "资产编号已存在"})
		return
	}

	if req.PurchasePrice < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "购入价格不能为负数"})
		return
	}

	purchaseDate, err := parseDate(req.PurchaseDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "购入日期格式错误"})
		return
	}

	if purchaseDate.After(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "购入日期不能晚于当前日期"})
		return
	}

	warrantyExpiryDate, _ := parseDate(req.WarrantyExpiryDate)

	category := models.DeviceCategory(req.Category)
	nextCalibrationDate := purchaseDate.Add(category.GetCalibrationInterval())

	device := models.Device{
		AssetNumber:        req.AssetNumber,
		Name:               req.Name,
		BrandModel:         req.BrandModel,
		SerialNumber:       req.SerialNumber,
		Category:           category,
		Department:         req.Department,
		Location:           req.Location,
		PurchaseDate:       purchaseDate,
		PurchasePrice:      req.PurchasePrice,
		WarrantyExpiryDate: warrantyExpiryDate,
		Status:             models.StatusInUse,
		NextCalibrationDate: &nextCalibrationDate,
	}

	if err := database.DB.Create(&device).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, device)
}

func GetDevices(c *gin.Context) {
	department := c.Query("department")
	category := c.Query("category")
	status := c.Query("status")

	query := database.DB.Model(&models.Device{})

	if department != "" {
		query = query.Where("department = ?", department)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var devices []models.Device
	if err := query.Find(&devices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, devices)
}

func GetDevice(c *gin.Context) {
	id := c.Param("id")

	var device models.Device
	if err := database.DB.Preload("CalibrationRecords").Preload("WorkOrders").Preload("MaintenancePlans").First(&device, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "设备不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, device)
}

func UpdateDeviceStatus(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var device models.Device
	if err := database.DB.First(&device, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "设备不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newStatus := models.DeviceStatus(req.Status)

	device.Status = newStatus
	if err := database.DB.Save(&device).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	message := ""
	if newStatus == models.StatusScrapped {
		result := database.DB.Model(&models.MaintenancePlan{}).
			Where("device_id = ? AND status IN ?", device.ID, []models.MaintenancePlanStatus{
				models.MaintenancePlanStatusPending,
				models.MaintenancePlanStatusOverdue,
			}).
			Update("status", models.MaintenancePlanStatusCancelled)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
		message = "设备已报废计划已取消"
	}

	response := gin.H{
		"device": device,
	}
	if message != "" {
		response["message"] = message
	}
	c.JSON(http.StatusOK, response)
}

func ExportDevicesCSV(c *gin.Context) {
	var devices []models.Device
	if err := database.DB.Find(&devices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=devices.csv")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	headers := []string{"ID", "资产编号", "设备名称", "品牌型号", "序列号", "分类", "科室", "位置", "购入日期", "购入价格", "保修截止", "状态"}
	if err := writer.Write(headers); err != nil {
		return
	}

	for _, d := range devices {
		row := []string{
			strconv.FormatUint(uint64(d.ID), 10),
			d.AssetNumber,
			d.Name,
			d.BrandModel,
			d.SerialNumber,
			string(d.Category),
			d.Department,
			d.Location,
			d.PurchaseDate.Format("2006-01-02"),
			strconv.FormatFloat(d.PurchasePrice, 'f', 2, 64),
			d.WarrantyExpiryDate.Format("2006-01-02"),
			string(d.Status),
		}
		if err := writer.Write(row); err != nil {
			return
		}
	}
}
