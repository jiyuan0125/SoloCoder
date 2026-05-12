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

type TestHandler struct {
	db *gorm.DB
}

func NewTestHandler(db *gorm.DB) *TestHandler {
	return &TestHandler{db: db}
}

type TestResults struct {
	ABOFront     string `json:"abo_front" binding:"required"`
	ABOBack      string `json:"abo_back" binding:"required"`
	RhD          string `json:"rhd" binding:"required"`
	ALT          string `json:"alt" binding:"required"`
	HBsAg        string `json:"hbsag" binding:"required"`
	AntiHCV      string `json:"anti_hcv" binding:"required"`
	AntiHIV      string `json:"anti_hiv" binding:"required"`
	AntiSyphilis string `json:"anti_syphilis" binding:"required"`
	NAT          string `json:"nat" binding:"required"`
}

type RecordTestRequest struct {
	CollectionID  string      `json:"collection_id" binding:"required"`
	TestRound     int         `json:"test_round" binding:"required"`
	ReagentVendor string      `json:"reagent_vendor"`
	Results       TestResults `json:"results" binding:"required"`
	OperatorID    string      `json:"operator_id"`
}

func isValidTestResult(result string) bool {
	return result == config.ResultNegative || 
		   result == config.ResultPositive || 
		   result == config.ResultInvalid
}

func hasAnyPositive(results TestResults) bool {
	return results.ABOFront == config.ResultPositive ||
		   results.ABOBack == config.ResultPositive ||
		   results.RhD == config.ResultPositive ||
		   results.ALT == config.ResultPositive ||
		   results.HBsAg == config.ResultPositive ||
		   results.AntiHCV == config.ResultPositive ||
		   results.AntiHIV == config.ResultPositive ||
		   results.AntiSyphilis == config.ResultPositive ||
		   results.NAT == config.ResultPositive
}

func hasAnyInvalid(results TestResults) bool {
	return results.ABOFront == config.ResultInvalid ||
		   results.ABOBack == config.ResultInvalid ||
		   results.RhD == config.ResultInvalid ||
		   results.ALT == config.ResultInvalid ||
		   results.HBsAg == config.ResultInvalid ||
		   results.AntiHCV == config.ResultInvalid ||
		   results.AntiHIV == config.ResultInvalid ||
		   results.AntiSyphilis == config.ResultInvalid ||
		   results.NAT == config.ResultInvalid
}

func (h *TestHandler) Record(c *gin.Context) {
	var req RecordTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !isValidTestResult(req.Results.ABOFront) ||
	   !isValidTestResult(req.Results.ABOBack) ||
	   !isValidTestResult(req.Results.RhD) ||
	   !isValidTestResult(req.Results.ALT) ||
	   !isValidTestResult(req.Results.HBsAg) ||
	   !isValidTestResult(req.Results.AntiHCV) ||
	   !isValidTestResult(req.Results.AntiHIV) ||
	   !isValidTestResult(req.Results.AntiSyphilis) ||
	   !isValidTestResult(req.Results.NAT) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid test result value. Must be Negative, Positive, or Invalid"})
		return
	}

	var collection models.BloodCollection
	if err := h.db.Where("id = ?", req.CollectionID).First(&collection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if collection.Status == config.StatusScrapped || collection.Status == config.StatusDisqualified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Blood is already scrapped"})
		return
	}

	var existingRecords []models.TestRecord
	h.db.Where("collection_id = ?", req.CollectionID).Order("test_round asc").Find(&existingRecords)

	if req.TestRound == config.TestRound1 && len(existingRecords) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Round 1 test already recorded"})
		return
	}

	if req.TestRound == config.TestRound2 && len(existingRecords) < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Round 1 test must be completed first"})
		return
	}

	if req.TestRound == config.TestRound2 && len(existingRecords) >= 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Round 2 test already recorded"})
		return
	}

	if req.TestRound == config.TestRound3 {
		if len(existingRecords) < 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Round 2 test must be completed first"})
			return
		}
		if len(existingRecords) >= 3 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Round 3 test already recorded"})
			return
		}
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if collection.Status == config.StatusPendingTest {
			collection.Status = config.StatusTesting
			collection.UpdatedAt = time.Now()
			if err := tx.Save(&collection).Error; err != nil {
				return err
			}
		}

		record := models.TestRecord{
			ID:            models.GenerateUUID(),
			CollectionID:  req.CollectionID,
			TestRound:     req.TestRound,
			ReagentVendor: req.ReagentVendor,
			ABOFront:      req.Results.ABOFront,
			ABOBack:       req.Results.ABOBack,
			RhD:           req.Results.RhD,
			ALT:           req.Results.ALT,
			HBsAg:         req.Results.HBsAg,
			AntiHCV:       req.Results.AntiHCV,
			AntiHIV:       req.Results.AntiHIV,
			AntiSyphilis:  req.Results.AntiSyphilis,
			NAT:           req.Results.NAT,
			OperatorID:    req.OperatorID,
			TestTime:      time.Now(),
			CreatedAt:     time.Now(),
		}

		if err := tx.Create(&record).Error; err != nil {
			return err
		}

		collection.TestProgress = req.TestRound
		collection.UpdatedAt = time.Now()

		if hasAnyPositive(req.Results) {
			collection.Status = config.StatusScrapped
			collection.ScrapReason = fmt.Sprintf("Test round %d has positive result", req.TestRound)
			collection.UpdatedAt = time.Now()
			if err := tx.Save(&collection).Error; err != nil {
				return err
			}
			return nil
		}

		if req.TestRound == config.TestRound2 {
			if hasAnyInvalid(req.Results) {
				collection.TestProgress = 3
				collection.UpdatedAt = time.Now()
				if err := tx.Save(&collection).Error; err != nil {
					return err
				}
				return nil
			}

			collection.Status = config.StatusQualified
			collection.UpdatedAt = time.Now()
			if err := tx.Save(&collection).Error; err != nil {
				return err
			}

			collection.Status = config.StatusPendingStorage
			var donor models.Donor
			if err := tx.Where("id = ?", collection.DonorID).First(&donor).Error; err != nil {
				return err
			}

			inventory := models.Inventory{
				ID:           models.GenerateUUID(),
				CollectionID: collection.ID,
				Barcode:      collection.Barcode,
				BloodType:    donor.BloodType,
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

			collection.UpdatedAt = time.Now()
			if err := tx.Save(&collection).Error; err != nil {
				return err
			}
		}

		if req.TestRound == config.TestRound3 {
			if hasAnyPositive(req.Results) || hasAnyInvalid(req.Results) {
				collection.Status = config.StatusScrapped
				collection.ScrapReason = "Round 3 retest still has positive or invalid results"
				collection.UpdatedAt = time.Now()
				if err := tx.Save(&collection).Error; err != nil {
					return err
				}
			} else {
				collection.Status = config.StatusQualified
				collection.UpdatedAt = time.Now()
				if err := tx.Save(&collection).Error; err != nil {
					return err
				}

				collection.Status = config.StatusPendingStorage
				var donor models.Donor
				if err := tx.Where("id = ?", collection.DonorID).First(&donor).Error; err != nil {
					return err
				}

				inventory := models.Inventory{
					ID:           models.GenerateUUID(),
					CollectionID: collection.ID,
					Barcode:      collection.Barcode,
					BloodType:    donor.BloodType,
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

				collection.UpdatedAt = time.Now()
				if err := tx.Save(&collection).Error; err != nil {
					return err
				}
			}
		}

		if err := tx.Save(&collection).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var updatedCollection models.BloodCollection
	h.db.Where("id = ?", req.CollectionID).First(&updatedCollection)
	c.JSON(http.StatusCreated, gin.H{
		"collection": updatedCollection,
		"test_round": req.TestRound,
	})
}

func (h *TestHandler) Get(c *gin.Context) {
	collectionID := c.Param("collection_id")
	
	var collection models.BloodCollection
	if err := h.db.Where("id = ?", collectionID).First(&collection).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	}

	var records []models.TestRecord
	if err := h.db.Where("collection_id = ?", collectionID).Order("test_round asc").Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"collection":   collection,
		"test_records": records,
		"progress":     gin.H{
			"current_round": collection.TestProgress,
			"status":        collection.Status,
			"is_scrapped":   collection.Status == config.StatusScrapped || collection.Status == config.StatusDisqualified,
		},
	})
}
