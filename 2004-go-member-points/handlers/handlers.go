package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"member-points/services"
)

type CreateMemberRequest struct {
	Name string `json:"name" binding:"required"`
}

type ConsumptionRequest struct {
	MemberID int64 `json:"member_id" binding:"required"`
	Amount   int64 `json:"amount" binding:"required"`
}

type CreateProductRequest struct {
	Name        string `json:"name" binding:"required"`
	PointsCost  int64  `json:"points_cost" binding:"required"`
	Description string `json:"description"`
}

type ExchangeRequest struct {
	MemberID  int64 `json:"member_id" binding:"required"`
	ProductID int64 `json:"product_id" binding:"required"`
}

func CreateMember(c *gin.Context) {
	var req CreateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := services.CreateMember(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, member)
}

func GetMember(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member id"})
		return
	}

	member, err := services.GetMemberByID(id)
	if err != nil {
		if err.Error() == "member not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, member)
}

func RecordConsumption(c *gin.Context) {
	var req ConsumptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be positive"})
		return
	}

	consumption, member, err := services.RecordConsumption(req.MemberID, req.Amount)
	if err != nil {
		if err.Error() == "member not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
			return
		}
		if err.Error() == "amount must be positive" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"consumption": consumption,
		"member":      member,
	})
}

func CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := services.CreateProduct(req.Name, req.PointsCost, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, product)
}

func GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	product, err := services.GetProductByID(id)
	if err != nil {
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

func ListProducts(c *gin.Context) {
	products, err := services.ListProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, products)
}

func ExchangeProduct(c *gin.Context) {
	var req ExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	exchange, member, err := services.ExchangeProduct(req.MemberID, req.ProductID)
	if err != nil {
		var insufficientErr *services.InsufficientPointsError
		if errors.As(err, &insufficientErr) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":           "insufficient points",
				"current_points":  insufficientErr.CurrentPoints,
				"required_points": insufficientErr.RequiredPoints,
			})
			return
		}

		var alreadyExchanged *services.AlreadyExchangedError
		if errors.As(err, &alreadyExchanged) {
			c.JSON(http.StatusConflict, gin.H{"error": "product already exchanged by this member"})
			return
		}

		if err.Error() == "member not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
			return
		}

		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"exchange": exchange,
		"member":   member,
	})
}

func ProcessYearEnd(c *gin.Context) {
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
		return
	}

	records, err := services.ProcessYearEnd(year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"processed_count": len(records),
		"records":         records,
	})
}
