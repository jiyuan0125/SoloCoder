package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/performance-review/database"
	"github.com/performance-review/models"
	"github.com/performance-review/services"
)

type ReviewRequest struct {
	EmployeeID    int64   `json:"employee_id" binding:"required"`
	QuarterID     int64   `json:"quarter_id" binding:"required"`
	Quality       float64 `json:"quality" binding:"required"`
	Efficiency    float64 `json:"efficiency" binding:"required"`
	Collaboration float64 `json:"collaboration" binding:"required"`
	Innovation    float64 `json:"innovation" binding:"required"`
}

func CreateOrUpdateReview(c *gin.Context) {
	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !services.ValidateScore(req.Quality) || !services.ValidateScore(req.Efficiency) ||
		!services.ValidateScore(req.Collaboration) || !services.ValidateScore(req.Innovation) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评分范围 1-5"})
		return
	}

	if !services.EmployeeExists(req.EmployeeID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Employee not found"})
		return
	}

	if !services.QuarterExists(req.QuarterID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quarter not found"})
		return
	}

	totalScore := services.CalculateTotalScore(req.Quality, req.Efficiency, req.Collaboration, req.Innovation)
	level := services.CalculateLevel(totalScore)

	existing, err := services.GetDraftReview(req.EmployeeID, req.QuarterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if existing != nil {
		if existing.Status == "confirmed" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Review already confirmed"})
			return
		}

		_, err = database.DB.Exec(`
			UPDATE reviews SET quality = ?, efficiency = ?, collaboration = ?, innovation = ?,
			total_score = ?, level = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, req.Quality, req.Efficiency, req.Collaboration, req.Innovation, totalScore, level, existing.ID)
	} else {
		_, err = database.DB.Exec(`
			INSERT INTO reviews (employee_id, quarter_id, quality, efficiency, collaboration, innovation, total_score, level, status, final_level, need_improvement)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'draft', '', 0)
		`, req.EmployeeID, req.QuarterID, req.Quality, req.Efficiency, req.Collaboration, req.Innovation, totalScore, level)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	review, err := services.GetDraftReview(req.EmployeeID, req.QuarterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, review)
}

type ConfirmRequest struct {
	QuarterID int64 `json:"quarter_id" binding:"required"`
}

func ConfirmReviews(c *gin.Context) {
	var req ConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !services.QuarterExists(req.QuarterID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quarter not found"})
		return
	}

	_, err := database.DB.Exec(`
		UPDATE reviews SET status = 'confirmed', final_level = level, updated_at = CURRENT_TIMESTAMP
		WHERE quarter_id = ? AND status = 'draft'
	`, req.QuarterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = services.ApplySRestriction(req.QuarterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, err := database.DB.Query(`
		SELECT employee_id FROM reviews WHERE quarter_id = ? AND status = 'confirmed'
	`, req.QuarterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var employeeIDs []int64
	for rows.Next() {
		var empID int64
		if err := rows.Scan(&empID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		employeeIDs = append(employeeIDs, empID)
	}

	for _, empID := range employeeIDs {
		isConsecutiveD, _ := services.CheckConsecutiveD(empID, req.QuarterID)
		if isConsecutiveD {
			_, err = database.DB.Exec(`
				UPDATE reviews SET need_improvement = 1, updated_at = CURRENT_TIMESTAMP
				WHERE employee_id = ? AND quarter_id = ?
			`, empID, req.QuarterID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Reviews confirmed successfully",
		"quarter_id": req.QuarterID,
	})
}

func GetReview(c *gin.Context) {
	id := c.Param("id")

	var review models.Review
	err := database.DB.QueryRow(`
		SELECT id, employee_id, quarter_id, quality, efficiency, collaboration, innovation,
		       total_score, level, final_level, status, need_improvement, created_at, updated_at
		FROM reviews WHERE id = ?
	`, id).Scan(
		&review.ID, &review.EmployeeID, &review.QuarterID, &review.Quality, &review.Efficiency,
		&review.Collaboration, &review.Innovation, &review.TotalScore, &review.Level,
		&review.FinalLevel, &review.Status, &review.NeedImprovement, &review.CreatedAt, &review.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	c.JSON(http.StatusOK, review)
}

func ListReviews(c *gin.Context) {
	quarterID := c.Query("quarter_id")
	employeeID := c.Query("employee_id")
	teamID := c.Query("team_id")

	query := `
		SELECT r.id, r.employee_id, r.quarter_id, r.quality, r.efficiency, r.collaboration, r.innovation,
		       r.total_score, r.level, r.final_level, r.status, r.need_improvement, r.created_at, r.updated_at
		FROM reviews r
	`
	args := []interface{}{}
	conditions := []string{}

	if teamID != "" {
		query += ` JOIN employees e ON r.employee_id = e.id `
	}

	if quarterID != "" {
		conditions = append(conditions, "r.quarter_id = ?")
		args = append(args, quarterID)
	}
	if employeeID != "" {
		conditions = append(conditions, "r.employee_id = ?")
		args = append(args, employeeID)
	}
	if teamID != "" {
		conditions = append(conditions, "e.team_id = ?")
		args = append(args, teamID)
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

	query += " ORDER BY r.id DESC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var reviews []models.Review
	for rows.Next() {
		var review models.Review
		if err := rows.Scan(
			&review.ID, &review.EmployeeID, &review.QuarterID, &review.Quality, &review.Efficiency,
			&review.Collaboration, &review.Innovation, &review.TotalScore, &review.Level,
			&review.FinalLevel, &review.Status, &review.NeedImprovement, &review.CreatedAt, &review.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		reviews = append(reviews, review)
	}

	c.JSON(http.StatusOK, reviews)
}
