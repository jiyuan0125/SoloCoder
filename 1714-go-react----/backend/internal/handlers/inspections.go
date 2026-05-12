package handlers

import (
	"net/http"
	"strings"
	"time"

	"health-supervision-system/internal/models"
	"health-supervision-system/internal/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateInspectionRequest struct {
	UnitID         uint                    `json:"unit_id" binding:"required"`
	InspectionDate string                  `json:"inspection_date" binding:"required"`
	Inspectors     string                  `json:"inspectors" binding:"required"`
	InspectionType models.InspectionType   `json:"inspection_type" binding:"required"`
	Items          []InspectionItemRequest `json:"items" binding:"required"`
}

type InspectionItemRequest struct {
	ItemName string                  `json:"item_name" binding:"required"`
	Result   models.InspectionResult `json:"result" binding:"required"`
}

type UpdateInspectionItemsRequest struct {
	Items []InspectionItemRequest `json:"items" binding:"required"`
}

func ListInspections(c *gin.Context) {
	unitID := c.Query("unit_id")
	inspectionType := c.Query("type")

	var inspections []models.InspectionRecord
	query := storage.DB.Preload("Items").Preload("Unit")

	if unitID != "" {
		query = query.Where("unit_id = ?", unitID)
	}
	if inspectionType != "" {
		query = query.Where("inspection_type = ?", inspectionType)
	}

	if err := query.Order("created_at DESC").Find(&inspections).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list inspections"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": inspections})
}

func GetInspection(c *gin.Context) {
	id := c.Param("id")
	var inspection models.InspectionRecord

	if err := storage.DB.Preload("Items").Preload("Unit").First(&inspection, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "inspection not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get inspection"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": inspection})
}

func GetInspectionTemplates(c *gin.Context) {
	unitType := c.Query("unit_type")
	if unitType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unit_type is required"})
		return
	}

	var templates []models.InspectionItemTemplate
	if err := storage.DB.Where("unit_type = ?", unitType).Find(&templates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": templates})
}

func CalculatePassRate(items []models.InspectionItem) (float64, bool) {
	applicable := 0
	pass := 0

	for _, item := range items {
		if item.Result != models.ResultNotApply {
			applicable++
			if item.Result == models.ResultPass {
				pass++
			}
		}
	}

	if applicable == 0 {
		return 0, false
	}

	passRate := float64(pass) / float64(applicable) * 100
	isQualified := passRate >= 80

	return passRate, isQualified
}

func CreateInspection(c *gin.Context) {
	var req CreateInspectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var unit models.SupervisedUnit
	if err := storage.DB.First(&unit, req.UnitID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "unit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get unit"})
		return
	}

	if unit.Status == models.StatusSuspended {
		c.JSON(http.StatusBadRequest, gin.H{"error": "被停业整顿的单位不能被选为新的检查对象"})
		return
	}

	inspectors := strings.Split(req.Inspectors, ",")
	if len(inspectors) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "检查人员必须两名以上"})
		return
	}

	inspectionDate, err := time.Parse("2006-01-02", req.InspectionDate)
	if err != nil {
		inspectionDate, err = time.Parse(time.RFC3339, req.InspectionDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "日期格式错误"})
			return
		}
	}

	tx := storage.DB.Begin()

	inspection := models.InspectionRecord{
		UnitID:         req.UnitID,
		InspectionDate: inspectionDate,
		Inspectors:     req.Inspectors,
		InspectionType: req.InspectionType,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := tx.Create(&inspection).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create inspection"})
		return
	}

	items := make([]models.InspectionItem, 0, len(req.Items))
	for _, itemReq := range req.Items {
		items = append(items, models.InspectionItem{
			InspectionID: inspection.ID,
			ItemName:     itemReq.ItemName,
			Result:       itemReq.Result,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		})
	}

	if err := tx.Create(&items).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create inspection items"})
		return
	}

	passRate, isQualified := CalculatePassRate(items)
	inspection.PassRate = passRate
	inspection.IsQualified = isQualified

	if err := tx.Save(&inspection).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update inspection"})
		return
	}

	if !isQualified {
		unit.Status = models.StatusRectifying
		if err := tx.Save(&unit).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update unit status"})
			return
		}
	}

	tx.Commit()

	inspectionID := inspection.ID
	storage.CreateAuditLog(0, "system", "创建", "检查记录", &inspectionID, "创建检查记录，单位: "+unit.Name, c.ClientIP())

	inspection.Items = items
	c.JSON(http.StatusCreated, gin.H{"data": inspection})
}

func UpdateInspectionItems(c *gin.Context) {
	id := c.Param("id")
	var inspection models.InspectionRecord

	if err := storage.DB.First(&inspection, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "inspection not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get inspection"})
		return
	}

	var req UpdateInspectionItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := storage.DB.Begin()

	if err := tx.Where("inspection_id = ?", inspection.ID).Delete(&models.InspectionItem{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete old items"})
		return
	}

	items := make([]models.InspectionItem, 0, len(req.Items))
	for _, itemReq := range req.Items {
		items = append(items, models.InspectionItem{
			InspectionID: inspection.ID,
			ItemName:     itemReq.ItemName,
			Result:       itemReq.Result,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		})
	}

	if err := tx.Create(&items).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create inspection items"})
		return
	}

	passRate, isQualified := CalculatePassRate(items)
	inspection.PassRate = passRate
	inspection.IsQualified = isQualified
	inspection.UpdatedAt = time.Now()

	if err := tx.Save(&inspection).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update inspection"})
		return
	}

	var unit models.SupervisedUnit
	if err := tx.First(&unit, inspection.UnitID).Error; err == nil {
		if !isQualified && unit.Status == models.StatusNormal {
			unit.Status = models.StatusRectifying
			tx.Save(&unit)
		}
	}

	tx.Commit()

	inspectionID := inspection.ID
	storage.CreateAuditLog(0, "system", "更新", "检查记录", &inspectionID, "更新检查记录项", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"data": inspection})
}

func DeleteInspection(c *gin.Context) {
	id := c.Param("id")
	var inspection models.InspectionRecord

	if err := storage.DB.First(&inspection, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "inspection not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get inspection"})
		return
	}

	tx := storage.DB.Begin()

	if err := tx.Where("inspection_id = ?", inspection.ID).Delete(&models.InspectionItem{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete inspection items"})
		return
	}

	if err := tx.Delete(&inspection).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete inspection"})
		return
	}

	tx.Commit()

	inspectionID := inspection.ID
	storage.CreateAuditLog(0, "system", "删除", "检查记录", &inspectionID, "删除检查记录", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "inspection deleted"})
}
