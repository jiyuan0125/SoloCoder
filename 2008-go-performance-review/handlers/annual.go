package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/performance-review/database"
	"github.com/performance-review/models"
	"github.com/performance-review/services"
)

type GenerateAnnualRequest struct {
	Year int `json:"year" binding:"required"`
}

func GenerateAnnualReviews(c *gin.Context) {
	var req GenerateAnnualRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rows, err := database.DB.Query(`
		SELECT DISTINCT e.id FROM employees e
		JOIN reviews r ON e.id = r.employee_id
		JOIN quarters q ON r.quarter_id = q.id
		WHERE q.year = ? AND r.status = 'confirmed'
		GROUP BY e.id
		HAVING COUNT(DISTINCT q.quarter) = 4
	`, req.Year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []gin.H
	for rows.Next() {
		var empID int64
		if err := rows.Scan(&empID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		annual, err := services.CalculateAnnualReview(empID, req.Year)
		if err != nil {
			continue
		}

		var existingID int64
		err = database.DB.QueryRow(`
			SELECT id FROM annual_reviews WHERE employee_id = ? AND year = ?
		`, empID, req.Year).Scan(&existingID)

		if err == nil {
			_, err = database.DB.Exec(`
				UPDATE annual_reviews SET average_score = ?, level = ? WHERE id = ?
			`, annual.AverageScore, annual.Level, existingID)
		} else {
			_, err = database.DB.Exec(`
				INSERT INTO annual_reviews (employee_id, year, average_score, level)
				VALUES (?, ?, ?, ?)
			`, empID, req.Year, annual.AverageScore, annual.Level)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		results = append(results, gin.H{
			"employee_id":   empID,
			"average_score": annual.AverageScore,
			"level":         annual.Level,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Annual reviews generated",
		"year":    req.Year,
		"count":   len(results),
		"results": results,
	})
}

func ListAnnualReviews(c *gin.Context) {
	year := c.Query("year")
	employeeID := c.Query("employee_id")

	query := `
		SELECT id, employee_id, year, average_score, level, created_at
		FROM annual_reviews
	`
	args := []interface{}{}
	conditions := []string{}

	if year != "" {
		conditions = append(conditions, "year = ?")
		args = append(args, year)
	}
	if employeeID != "" {
		conditions = append(conditions, "employee_id = ?")
		args = append(args, employeeID)
	}

	if len(conditions) > 0 {
		query += " WHERE "
		for i, cond := range conditions {
			if i > 0 {
				query += " AND "
			}
			query += cond
		}
	}

	query += " ORDER BY year DESC, average_score DESC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var reviews []models.AnnualReview
	for rows.Next() {
		var review models.AnnualReview
		if err := rows.Scan(&review.ID, &review.EmployeeID, &review.Year, &review.AverageScore, &review.Level, &review.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		reviews = append(reviews, review)
	}

	c.JSON(http.StatusOK, reviews)
}

func GetEmployeeAnnualReview(c *gin.Context) {
	employeeIDStr := c.Param("employee_id")
	yearStr := c.Param("year")

	employeeID, err := strconv.ParseInt(employeeIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	annual, err := services.CalculateAnnualReview(employeeID, year)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, annual)
}
