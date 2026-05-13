package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/performance-review/database"
	"github.com/performance-review/models"
)

type CreateQuarterRequest struct {
	Year    int `json:"year" binding:"required"`
	Quarter int `json:"quarter" binding:"required"`
}

func CreateQuarter(c *gin.Context) {
	var req CreateQuarterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Quarter < 1 || req.Quarter > 4 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quarter must be 1-4"})
		return
	}

	result, err := database.DB.Exec(`
		INSERT INTO quarters (year, quarter) VALUES (?, ?)
	`, req.Year, req.Quarter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()

	c.JSON(http.StatusCreated, gin.H{
		"id":      id,
		"year":    req.Year,
		"quarter": req.Quarter,
	})
}

func ListQuarters(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT id, year, quarter, created_at FROM quarters ORDER BY year DESC, quarter DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var quarters []models.Quarter
	for rows.Next() {
		var q models.Quarter
		if err := rows.Scan(&q.ID, &q.Year, &q.Quarter, &q.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		quarters = append(quarters, q)
	}

	c.JSON(http.StatusOK, quarters)
}

func GetQuarter(c *gin.Context) {
	id := c.Param("id")

	var q models.Quarter
	err := database.DB.QueryRow(`
		SELECT id, year, quarter, created_at FROM quarters WHERE id = ?
	`, id).Scan(&q.ID, &q.Year, &q.Quarter, &q.CreatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quarter not found"})
		return
	}

	c.JSON(http.StatusOK, q)
}
