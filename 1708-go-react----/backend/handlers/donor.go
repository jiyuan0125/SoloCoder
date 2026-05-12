package handlers

import (
	"blood-management-system/config"
	"blood-management-system/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DonorHandler struct {
	db *gorm.DB
}

func NewDonorHandler(db *gorm.DB) *DonorHandler {
	return &DonorHandler{db: db}
}

type RegisterDonorRequest struct {
	Name            string    `json:"name" binding:"required"`
	IDCard          string    `json:"id_card" binding:"required"`
	Gender          string    `json:"gender" binding:"required"`
	BirthDate       time.Time `json:"birth_date" binding:"required"`
	ABOBloodType    string    `json:"abo_blood_type" binding:"required"`
	RhFactor        string    `json:"rh_factor" binding:"required"`
	Phone           string    `json:"phone"`
	Address         string    `json:"address"`
	HealthCheck     *HealthCheckData `json:"health_check"`
	DonationType    string    `json:"donation_type" binding:"required"`
	CurrentDonationDate time.Time `json:"current_donation_date"`
}

type HealthCheckData struct {
	HeightCm         float64 `json:"height_cm"`
	WeightKg         float64 `json:"weight_kg"`
	RecentMedication string  `json:"recent_medication"`
	IsFasting        bool    `json:"is_fasting"`
	Notes            string  `json:"notes"`
}

func (h *DonorHandler) Register(c *gin.Context) {
	var req RegisterDonorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingDonor models.Donor
	if err := h.db.Where("id_card = ?", req.IDCard).First(&existingDonor).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Donor with this ID card already exists"})
		return
	}

	if req.CurrentDonationDate.IsZero() {
		req.CurrentDonationDate = time.Now()
	}

	bloodType := req.ABOBloodType + "_" + req.RhFactor

	donor := models.Donor{
		ID:               models.GenerateUUID(),
		DonorNumber:      "",
		Name:             req.Name,
		IDCard:           req.IDCard,
		Gender:           req.Gender,
		BirthDate:        req.BirthDate,
		ABOBloodType:     req.ABOBloodType,
		RhFactor:         req.RhFactor,
		BloodType:        bloodType,
		Phone:            req.Phone,
		Address:          req.Address,
		LastDonationDate: nil,
		LastDonationType: nil,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Donor{}).Count(&count).Error; err != nil {
			return err
		}
		donor.DonorNumber = models.GenerateDonorNumber(int(count) + 1)

		if err := tx.Create(&donor).Error; err != nil {
			return err
		}

		if req.HealthCheck != nil {
			healthCheck := models.HealthCheck{
				ID:               models.GenerateUUID(),
				DonorID:          donor.ID,
				HeightCm:         req.HealthCheck.HeightCm,
				WeightKg:         req.HealthCheck.WeightKg,
				RecentMedication: req.HealthCheck.RecentMedication,
				IsFasting:        req.HealthCheck.IsFasting,
				CheckDate:        req.CurrentDonationDate,
				Notes:            req.HealthCheck.Notes,
				CreatedAt:        time.Now(),
			}
			if err := tx.Create(&healthCheck).Error; err != nil {
				return err
			}
		}

		donationType := req.DonationType
		donor.LastDonationDate = &req.CurrentDonationDate
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

	c.JSON(http.StatusCreated, donor)
}

func CheckDonationInterval(db *gorm.DB, donorID string, currentDate time.Time, donationType string) (bool, int, error) {
	var donor models.Donor
	if err := db.Where("id = ?", donorID).First(&donor).Error; err != nil {
		return false, 0, err
	}

	if donor.LastDonationDate == nil {
		return true, 0, nil
	}

	var requiredDays int
	if donationType == config.CollectionPlatelets {
		requiredDays = 14
	} else {
		requiredDays = 180
	}

	daysSinceLast := int(currentDate.Sub(*donor.LastDonationDate).Hours() / 24)
	if daysSinceLast < requiredDays {
		return false, requiredDays - daysSinceLast, nil
	}

	return true, 0, nil
}

func (h *DonorHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var donor models.Donor
	if err := h.db.Where("id = ? OR donor_number = ?", id, id).First(&donor).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Donor not found"})
		return
	}
	c.JSON(http.StatusOK, donor)
}

func (h *DonorHandler) List(c *gin.Context) {
	var donors []models.Donor
	query := h.db.Order("created_at DESC")

	if name := c.Query("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if idCard := c.Query("id_card"); idCard != "" {
		query = query.Where("id_card = ?", idCard)
	}

	if err := query.Find(&donors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, donors)
}
