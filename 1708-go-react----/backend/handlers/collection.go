package handlers

import (
	"blood-management-system/config"
	"blood-management-system/models"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CollectionHandler struct {
	db *gorm.DB
}

func NewCollectionHandler(db *gorm.DB) *CollectionHandler {
	return &CollectionHandler{db: db}
}

type CreateCollectionRequest struct {
	DonorID          string    `json:"donor_id" binding:"required"`
	CollectionType   string    `json:"collection_type" binding:"required"`
	CollectorStaffID string    `json:"collector_staff_id"`
	CollectionTime   time.Time `json:"collection_time"`
}

var validStatusTransitions = map[string][]string{
	config.StatusPendingTest:  {config.StatusTesting},
	config.StatusTesting:      {config.StatusQualified, config.StatusDisqualified, config.StatusScrapped},
	config.StatusQualified:    {config.StatusPendingStorage},
	config.StatusPendingStorage: {config.StatusInStock},
	config.StatusInStock:      {config.StatusFrozen, config.StatusIssued, config.StatusExpired},
	config.StatusFrozen:       {config.StatusInStock},
	config.StatusScrapped:     {},
	config.StatusDisqualified: {},
	config.StatusIssued:       {},
	config.StatusExpired:      {},
}

func IsValidStatusTransition(current, next string) bool {
	validNext, exists := validStatusTransitions[current]
	if !exists {
		return false
	}
	for _, s := range validNext {
		if s == next {
			return true
		}
	}
	return false
}

func (h *CollectionHandler) Create(c *gin.Context) {
	var req CreateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.CollectionType != config.CollectionWhole200 && 
	   req.CollectionType != config.CollectionWhole400 && 
	   req.CollectionType != config.CollectionPlatelets {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection type"})
		return
	}

	if req.CollectionTime.IsZero() {
		req.CollectionTime = time.Now()
	}

	ok, daysLeft, err := CheckDonationInterval(h.db, req.DonorID, req.CollectionTime, req.CollectionType)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Donor not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "距上次献血不足" + fmt.Sprintf("%d", daysLeft) + "天"})
		return
	}

	var donor models.Donor
	if err := h.db.Where("id = ?", req.DonorID).First(&donor).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Donor not found"})
		return
	}

	volume := models.GetVolume(req.CollectionType)
	productType := models.GetProductType(req.CollectionType)
	expiryDate := models.GetExpiryDate(productType, req.CollectionTime)

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.BloodCollection{}).Count(&count).Error; err != nil {
			return err
		}
		barcode := models.GenerateBarcode(int(count) + 1)

		var existing models.BloodCollection
		for {
			if err := tx.Where("barcode = ?", barcode).First(&existing).Error; err != nil {
				break
			}
			count++
			barcode = models.GenerateBarcode(int(count) + 1)
		}

		collection := models.BloodCollection{
			ID:               models.GenerateUUID(),
			Barcode:          barcode,
			DonorID:          req.DonorID,
			CollectionType:   req.CollectionType,
			VolumeML:         volume,
			ProductType:      productType,
			CollectorStaffID: req.CollectorStaffID,
			CollectionTime:   req.CollectionTime,
			ExpiryDate:       expiryDate,
			Status:           config.StatusPendingTest,
			TestProgress:     0,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		if err := tx.Create(&collection).Error; err != nil {
			return err
		}

		donationType := req.CollectionType
		donor.LastDonationDate = &req.CollectionTime
		donor.LastDonationType = &donationType
		donor.UpdatedAt = time.Now()
		if err := tx.Save(&donor).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var createdCollection models.BloodCollection
	h.db.Where("donor_id = ? AND collection_time = ?", req.DonorID, req.CollectionTime).First(&createdCollection)
	c.JSON(http.StatusCreated, createdCollection)
}

func (h *CollectionHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var collection models.BloodCollection
	if err := h.db.Preload("Donor").Where("id = ? OR barcode = ?", id, id).First(&collection).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	}
	c.JSON(http.StatusOK, collection)
}

func (h *CollectionHandler) List(c *gin.Context) {
	var collections []models.BloodCollection
	query := h.db.Preload("Donor").Order("created_at DESC")

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if donorID := c.Query("donor_id"); donorID != "" {
		query = query.Where("donor_id = ?", donorID)
	}

	if err := query.Find(&collections).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, collections)
}

type UpdateStatusRequest struct {
	NewStatus string `json:"new_status" binding:"required"`
	Reason    string `json:"reason"`
}

func (h *CollectionHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var collection models.BloodCollection
	if err := h.db.Where("id = ? OR barcode = ?", id, id).First(&collection).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	}

	if !IsValidStatusTransition(collection.Status, req.NewStatus) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status transition: cannot jump from " + collection.Status + " to " + req.NewStatus})
		return
	}

	if collection.Status == config.StatusScrapped || collection.Status == config.StatusDisqualified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Blood is already scrapped and cannot be used"})
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		collection.Status = req.NewStatus
		collection.UpdatedAt = time.Now()
		if req.NewStatus == config.StatusScrapped || req.NewStatus == config.StatusDisqualified {
			collection.ScrapReason = req.Reason
		}

		if err := tx.Save(&collection).Error; err != nil {
			return err
		}

		if req.NewStatus == config.StatusPendingStorage {
			inventory := models.Inventory{
				ID:           models.GenerateUUID(),
				CollectionID: collection.ID,
				Barcode:      collection.Barcode,
				BloodType:    collection.Donor.BloodType,
				ProductType:  collection.ProductType,
				VolumeML:     collection.VolumeML,
				StorageDate:  time.Now(),
				ExpiryDate:   collection.ExpiryDate,
				Status:       config.StatusInStock,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
			if err := tx.Create(&inventory).Error; err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, collection)
}
