package handlers

import (
	"blood-management-system/config"
	"blood-management-system/models"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type InventoryHandler struct {
	db *gorm.DB
}

func NewInventoryHandler(db *gorm.DB) *InventoryHandler {
	return &InventoryHandler{db: db}
}

type FreezeRequest struct {
	Barcode string `json:"barcode" binding:"required"`
}

func (h *InventoryHandler) Freeze(c *gin.Context) {
	var req FreezeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.db.Transaction(func(tx *gorm.DB) error {
		var inventory models.Inventory
		if err := tx.Where("barcode = ?", req.Barcode).First(&inventory).Error; err != nil {
			return err
		}

		if inventory.Status != config.StatusInStock {
			return errors.New("inventory is not in stock")
		}

		inventory.Status = config.StatusFrozen
		inventory.UpdatedAt = time.Now()
		if err := tx.Save(&inventory).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inventory frozen successfully"})
}

func (h *InventoryHandler) Unfreeze(c *gin.Context) {
	var req FreezeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.db.Transaction(func(tx *gorm.DB) error {
		var inventory models.Inventory
		if err := tx.Where("barcode = ?", req.Barcode).First(&inventory).Error; err != nil {
			return err
		}

		if inventory.Status != config.StatusFrozen {
			return errors.New("inventory is not frozen")
		}

		inventory.Status = config.StatusInStock
		inventory.UpdatedAt = time.Now()
		if err := tx.Save(&inventory).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inventory unfrozen successfully"})
}

func (h *InventoryHandler) List(c *gin.Context) {
	var inventory []models.Inventory
	query := h.db.Preload("Collection").Order("storage_date ASC")

	if bloodType := c.Query("blood_type"); bloodType != "" {
		query = query.Where("blood_type = ?", bloodType)
	}
	if productType := c.Query("product_type"); productType != "" {
		query = query.Where("product_type = ?", productType)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&inventory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type GroupedInventory struct {
		BloodType   string `json:"blood_type"`
		ProductType string `json:"product_type"`
		Count       int64  `json:"count"`
		TotalVolume float64 `json:"total_volume"`
		MinStock    int    `json:"min_stock"`
		BelowMin    bool   `json:"below_min"`
		Items       []models.Inventory `json:"items"`
	}

	result := make([]GroupedInventory, 0)
	for _, bt := range config.BloodTypeCombinations {
		for _, pt := range config.ProductTypes {
			var items []models.Inventory
			var count int64
			var totalVolume float64

			h.db.Model(&models.Inventory{}).
				Where("blood_type = ? AND product_type = ? AND status IN ?", bt, pt, []string{config.StatusInStock, config.StatusFrozen}).
				Count(&count)

			h.db.Model(&models.Inventory{}).
				Where("blood_type = ? AND product_type = ? AND status IN ?", bt, pt, []string{config.StatusInStock, config.StatusFrozen}).
				Select("COALESCE(SUM(volume_ml), 0)").Scan(&totalVolume)

			var safetyStock models.SafetyStock
			h.db.Where("blood_type = ? AND product_type = ?", bt, pt).First(&safetyStock)

			h.db.Where("blood_type = ? AND product_type = ? AND status IN ?", bt, pt, []string{config.StatusInStock, config.StatusFrozen}).
				Order("storage_date ASC").Find(&items)

			belowMin := count < int64(safetyStock.MinQuantity)

			result = append(result, GroupedInventory{
				BloodType:   bt,
				ProductType: pt,
				Count:       count,
				TotalVolume: totalVolume,
				MinStock:    safetyStock.MinQuantity,
				BelowMin:    belowMin,
				Items:       items,
			})
		}
	}

	c.JSON(http.StatusOK, result)
}

func (h *InventoryHandler) GetByType(c *gin.Context) {
	bloodType := c.Param("blood_type")
	productType := c.Param("product_type")

	var inventory []models.Inventory
	if err := h.db.Where("blood_type = ? AND product_type = ?", bloodType, productType).
		Order("storage_date ASC").Find(&inventory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, inventory)
}

func IsBloodTypeCompatible(donorABO, donorRh, patientABO, patientRh string, isEmergency bool) bool {
	if donorRh != patientRh {
		return false
	}

	if isEmergency {
		return donorABO == config.ABO_O
	}

	return donorABO == patientABO
}

func GetAvailableInventoryForRequest(tx *gorm.DB, bloodType, productType string, quantity int) ([]models.Inventory, error) {
	var available []models.Inventory

	err := tx.Where("blood_type = ? AND product_type = ? AND status = ? AND expiry_date > ?", 
		bloodType, productType, config.StatusInStock, time.Now()).
		Order("storage_date ASC").
		Limit(quantity).
		Find(&available).Error

	if err != nil {
		return nil, err
	}

	if len(available) >= quantity {
		return available[:quantity], nil
	}

	return available, nil
}

func GetUniversalBlood(tx *gorm.DB, productType string, quantity int) ([]models.Inventory, error) {
	universalType := "O_Negative"
	
	var available []models.Inventory
	err := tx.Where("blood_type = ? AND product_type = ? AND status = ? AND expiry_date > ?",
		universalType, productType, config.StatusInStock, time.Now()).
		Order("storage_date ASC").
		Limit(quantity).
		Find(&available).Error

	if err != nil {
		return nil, err
	}

	return available, nil
}
