package handlers

import (
	"net/http"
	"time"

	"health-supervision-system/internal/models"
	"health-supervision-system/internal/storage"
	"health-supervision-system/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateUnitRequest struct {
	CreditCode            string `json:"credit_code" binding:"required"`
	Name                  string `json:"name" binding:"required"`
	Type                  models.UnitType `json:"type" binding:"required"`
	Address               string `json:"address"`
	LegalRepresentative   string `json:"legal_representative"`
	Phone                 string `json:"phone"`
	HealthLicenseNumber   string `json:"health_license_number" binding:"required"`
	HealthLicenseValidity string `json:"health_license_validity" binding:"required"`
}

type UpdateUnitRequest struct {
	Name                  string `json:"name"`
	Type                  models.UnitType `json:"type"`
	Address               string `json:"address"`
	LegalRepresentative   string `json:"legal_representative"`
	Phone                 string `json:"phone"`
	HealthLicenseNumber   string `json:"health_license_number"`
	HealthLicenseValidity string `json:"health_license_validity"`
	Status                models.UnitStatus `json:"status"`
}

func ListUnits(c *gin.Context) {
	unitType := c.Query("type")
	status := c.Query("status")

	var units []models.SupervisedUnit
	query := storage.DB.Model(&models.SupervisedUnit{})

	if unitType != "" {
		query = query.Where("type = ?", unitType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&units).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list units"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": units})
}

func GetUnit(c *gin.Context) {
	id := c.Param("id")
	var unit models.SupervisedUnit

	if err := storage.DB.First(&unit, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "unit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get unit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": unit})
}

func CreateUnit(c *gin.Context) {
	var req CreateUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !utils.IsValidCreditCode(req.CreditCode) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "统一社会信用代码格式不对"})
		return
	}

	validity, err := time.Parse("2006-01-02", req.HealthLicenseValidity)
	if err != nil {
		validity, err = time.Parse(time.RFC3339, req.HealthLicenseValidity)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "日期格式错误"})
			return
		}
	}

	var existingUnit models.SupervisedUnit
	if err := storage.DB.Where("credit_code = ?", req.CreditCode).First(&existingUnit).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "统一社会信用代码已存在"})
		return
	}

	if err := storage.DB.Where("health_license_number = ?", req.HealthLicenseNumber).First(&existingUnit).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "许可证编号重复"})
		return
	}

	if err := storage.DB.Where("name = ?", req.Name).First(&existingUnit).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "企业名称重复"})
		return
	}

	unit := models.SupervisedUnit{
		CreditCode:            req.CreditCode,
		Name:                  req.Name,
		Type:                  req.Type,
		Address:               req.Address,
		LegalRepresentative:   req.LegalRepresentative,
		Phone:                 req.Phone,
		HealthLicenseNumber:   req.HealthLicenseNumber,
		HealthLicenseValidity: validity,
		Status:                models.StatusNormal,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if err := storage.DB.Create(&unit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create unit"})
		return
	}

	unitID := unit.ID
	storage.CreateAuditLog(0, "system", "创建", "被监督单位", &unitID, "创建被监督单位: "+unit.Name, c.ClientIP())

	c.JSON(http.StatusCreated, gin.H{"data": unit})
}

func UpdateUnit(c *gin.Context) {
	id := c.Param("id")
	var unit models.SupervisedUnit

	if err := storage.DB.First(&unit, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "unit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get unit"})
		return
	}

	var req UpdateUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" && req.Name != unit.Name {
		var existingUnit models.SupervisedUnit
		if err := storage.DB.Where("name = ? AND id != ?", req.Name, unit.ID).First(&existingUnit).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "企业名称重复"})
			return
		}
		unit.Name = req.Name
	}
	if req.Type != "" {
		unit.Type = req.Type
	}
	if req.Address != "" {
		unit.Address = req.Address
	}
	if req.LegalRepresentative != "" {
		unit.LegalRepresentative = req.LegalRepresentative
	}
	if req.Phone != "" {
		unit.Phone = req.Phone
	}
	if req.HealthLicenseNumber != "" {
		var existingUnit models.SupervisedUnit
		if err := storage.DB.Where("health_license_number = ? AND id != ?", req.HealthLicenseNumber, unit.ID).First(&existingUnit).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "许可证编号重复"})
			return
		}
		unit.HealthLicenseNumber = req.HealthLicenseNumber
	}
	if req.HealthLicenseValidity != "" {
		validity, err := time.Parse("2006-01-02", req.HealthLicenseValidity)
		if err != nil {
			validity, err = time.Parse(time.RFC3339, req.HealthLicenseValidity)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "日期格式错误"})
				return
			}
		}
		unit.HealthLicenseValidity = validity
	}
	if req.Status != "" {
		unit.Status = req.Status
	}

	unit.UpdatedAt = time.Now()

	if err := storage.DB.Save(&unit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update unit"})
		return
	}

	unitID := unit.ID
	storage.CreateAuditLog(0, "system", "更新", "被监督单位", &unitID, "更新被监督单位: "+unit.Name, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"data": unit})
}

func DeleteUnit(c *gin.Context) {
	id := c.Param("id")
	var unit models.SupervisedUnit

	if err := storage.DB.First(&unit, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "unit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get unit"})
		return
	}

	if err := storage.DB.Delete(&unit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete unit"})
		return
	}

	unitID := unit.ID
	storage.CreateAuditLog(0, "system", "删除", "被监督单位", &unitID, "删除被监督单位: "+unit.Name, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "unit deleted"})
}

func GetUnitsWithLicenseWarning(c *gin.Context) {
	now := time.Now()
	warningDate := now.AddDate(0, 0, 30)

	var units []models.SupervisedUnit
	if err := storage.DB.Where("health_license_validity <= ? AND health_license_validity > ?", warningDate, now).
		Where("status != ?", models.StatusCancelled).
		Find(&units).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get warning units"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": units})
}
