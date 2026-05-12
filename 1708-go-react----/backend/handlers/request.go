package handlers

import (
	"blood-management-system/config"
	"blood-management-system/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RequestHandler struct {
	db *gorm.DB
}

func NewRequestHandler(db *gorm.DB) *RequestHandler {
	return &RequestHandler{db: db}
}

type CreateRequestRequest struct {
	HospitalName string `json:"hospital_name" binding:"required"`
	HospitalID   string `json:"hospital_id"`
	BloodType    string `json:"blood_type" binding:"required"`
	ProductType  string `json:"product_type" binding:"required"`
	Quantity     int    `json:"quantity" binding:"required,min=1"`
	Urgency      string `json:"urgency" binding:"required"`
	PatientInfo  string `json:"patient_info"`
	Notes        string `json:"notes"`
}

func (h *RequestHandler) Create(c *gin.Context) {
	var req CreateRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Urgency != config.UrgencyRegular && 
	   req.Urgency != config.UrgencyUrgent && 
	   req.Urgency != config.UrgencySpecial {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid urgency level"})
		return
	}

	validBloodType := false
	for _, bt := range config.BloodTypeCombinations {
		if bt == req.BloodType {
			validBloodType = true
			break
		}
	}
	if !validBloodType {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid blood type"})
		return
	}

	request := models.BloodRequest{
		ID:           models.GenerateUUID(),
		HospitalName: req.HospitalName,
		HospitalID:   req.HospitalID,
		BloodType:    req.BloodType,
		ProductType:  req.ProductType,
		Quantity:     req.Quantity,
		Urgency:      req.Urgency,
		PatientInfo:  req.PatientInfo,
		RequestTime:  time.Now(),
		Status:       config.ReqStatusPending,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&request).Error; err != nil {
			return err
		}

		available, err := GetAvailableInventoryForRequest(tx, req.BloodType, req.ProductType, req.Quantity)
		if err != nil {
			return err
		}

		if len(available) >= req.Quantity {
			for _, inv := range available[:req.Quantity] {
				inv.Status = config.StatusIssued
				inv.RequestID = &request.ID
				issuedDate := time.Now()
				inv.IssuedDate = &issuedDate
				inv.UpdatedAt = time.Now()
				if err := tx.Save(&inv).Error; err != nil {
					return err
				}

				var collection models.BloodCollection
				if err := tx.Where("id = ?", inv.CollectionID).First(&collection).Error; err != nil {
					return err
				}
				collection.Status = config.StatusIssued
				collection.UpdatedAt = time.Now()
				if err := tx.Save(&collection).Error; err != nil {
					return err
				}
			}
			request.Status = config.ReqStatusIssued
			request.UpdatedAt = time.Now()
			if err := tx.Save(&request).Error; err != nil {
				return err
			}
		} else {
			if req.Urgency == config.UrgencySpecial {
				universal, err := GetUniversalBlood(tx, req.ProductType, req.Quantity)
				if err != nil {
					return err
				}

				if len(universal) >= req.Quantity {
					for _, inv := range universal[:req.Quantity] {
						inv.Status = config.StatusIssued
						inv.RequestID = &request.ID
						issuedDate := time.Now()
						inv.IssuedDate = &issuedDate
						inv.UpdatedAt = time.Now()
						if err := tx.Save(&inv).Error; err != nil {
							return err
						}

						var collection models.BloodCollection
						if err := tx.Where("id = ?", inv.CollectionID).First(&collection).Error; err != nil {
							return err
						}
						collection.Status = config.StatusIssued
						collection.UpdatedAt = time.Now()
						if err := tx.Save(&collection).Error; err != nil {
							return err
						}
					}
					request.Status = config.ReqStatusIssued
					request.UsedUniversal = true
					request.UpdatedAt = time.Now()
					if err := tx.Save(&request).Error; err != nil {
						return err
					}
				} else {
					request.Status = config.ReqStatusQueued
					request.UpdatedAt = time.Now()
					if err := tx.Save(&request).Error; err != nil {
						return err
					}
				}
			} else if req.Urgency == config.UrgencyRegular {
				request.Status = config.ReqStatusQueued
				request.UpdatedAt = time.Now()
				if err := tx.Save(&request).Error; err != nil {
					return err
				}
			} else {
				request.Status = config.ReqStatusQueued
				request.UpdatedAt = time.Now()
				if err := tx.Save(&request).Error; err != nil {
					return err
				}
			}
		}

		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var savedRequest models.BloodRequest
	h.db.Where("id = ?", request.ID).First(&savedRequest)
	
	c.JSON(http.StatusCreated, savedRequest)
}

func (h *RequestHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var request models.BloodRequest
	if err := h.db.Where("id = ?", id).First(&request).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
		return
	}

	var inventory []models.Inventory
	h.db.Where("request_id = ?", id).Find(&inventory)

	c.JSON(http.StatusOK, gin.H{
		"request":   request,
		"inventory": inventory,
	})
}

func (h *RequestHandler) List(c *gin.Context) {
	var requests []models.BloodRequest
	query := h.db.Order("created_at DESC")

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if urgency := c.Query("urgency"); urgency != "" {
		query = query.Where("urgency = ?", urgency)
	}
	if hospital := c.Query("hospital"); hospital != "" {
		query = query.Where("hospital_name LIKE ? OR hospital_id = ?", "%"+hospital+"%", hospital)
	}

	if err := query.Find(&requests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, requests)
}

func (h *RequestHandler) Process(c *gin.Context) {
	id := c.Param("id")
	
	var request models.BloodRequest
	if err := h.db.Where("id = ?", id).First(&request).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
		return
	}

	if request.Status != config.ReqStatusQueued && request.Status != config.ReqStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request already processed"})
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		available, err := GetAvailableInventoryForRequest(tx, request.BloodType, request.ProductType, request.Quantity)
		if err != nil {
			return err
		}

		if len(available) < request.Quantity {
			if request.Urgency == config.UrgencySpecial {
				universal, err := GetUniversalBlood(tx, request.ProductType, request.Quantity)
				if err != nil {
					return err
				}
				if len(universal) < request.Quantity {
					return gin.Error{Type: gin.ErrorTypeBind, Err: err}
				}
				available = universal
				request.UsedUniversal = true
			} else {
				return gin.Error{Type: gin.ErrorTypeBind, Err: err}
			}
		}

		for _, inv := range available[:request.Quantity] {
			inv.Status = config.StatusIssued
			inv.RequestID = &request.ID
			issuedDate := time.Now()
			inv.IssuedDate = &issuedDate
			inv.UpdatedAt = time.Now()
			if err := tx.Save(&inv).Error; err != nil {
				return err
			}

			var collection models.BloodCollection
			if err := tx.Where("id = ?", inv.CollectionID).First(&collection).Error; err != nil {
				return err
			}
			collection.Status = config.StatusIssued
			collection.UpdatedAt = time.Now()
			if err := tx.Save(&collection).Error; err != nil {
				return err
			}
		}

		request.Status = config.ReqStatusIssued
		request.UpdatedAt = time.Now()
		if err := tx.Save(&request).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient inventory", "available": 0})
		return
	}

	var updatedRequest models.BloodRequest
	h.db.Where("id = ?", id).First(&updatedRequest)
	c.JSON(http.StatusOK, updatedRequest)
}
