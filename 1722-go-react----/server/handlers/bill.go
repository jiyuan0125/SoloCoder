package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"smart-exam/database"
	"smart-exam/models"
	"smart-exam/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BillHandler struct {
	billing *services.BillingService
}

func NewBillHandler() *BillHandler {
	return &BillHandler{
		billing: services.NewBillingService(),
	}
}

func (h *BillHandler) Create(c *gin.Context) {
	var req struct {
		StudentID     string  `json:"student_id" binding:"required"`
		TotalAmount   float64 `json:"total_amount" binding:"required"`
		ItemCount     int     `json:"item_count"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bill := models.Bill{
		StudentID:       req.StudentID,
		TotalAmount:     req.TotalAmount,
		RemainingAmount: req.TotalAmount,
		Status:          models.BillStatusActive,
	}

	if err := database.DB.Create(&bill).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.ItemCount > 0 {
		itemAmount := req.TotalAmount / float64(req.ItemCount)
		for i := 0; i < req.ItemCount; i++ {
			item := models.BillItem{
				BillID:      bill.ID,
				Description: "Exam " + strconv.Itoa(i+1),
				Amount:      itemAmount,
				IsCompleted: false,
			}
			if err := database.DB.Create(&item).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	c.JSON(http.StatusCreated, gin.H{"data": bill})
}

func (h *BillHandler) List(c *gin.Context) {
	studentID := c.Query("student_id")

	var bills []models.Bill
	query := database.DB

	if studentID != "" {
		query = query.Where("student_id = ?", studentID)
	}

	if err := query.Order("created_at DESC").Find(&bills).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": bills})
}

func (h *BillHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var bill models.Bill
	if err := database.DB.First(&bill, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "bill not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var items []models.BillItem
	database.DB.Where("bill_id = ?", id).Order("created_at ASC").Find(&items)

	c.JSON(http.StatusOK, gin.H{
		"bill":  bill,
		"items": items,
	})
}

func (h *BillHandler) AdjustTotal(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		NewTotal float64 `json:"new_total" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.billing.AdjustBillAmount(uint(id), req.NewTotal); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var bill models.Bill
	database.DB.First(&bill, id)

	c.JSON(http.StatusOK, gin.H{"data": bill})
}
