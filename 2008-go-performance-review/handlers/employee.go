package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/performance-review/database"
	"github.com/performance-review/models"
)

func CreateEmployee(c *gin.Context) {
	var emp models.Employee
	if err := c.ShouldBindJSON(&emp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := database.DB.Exec(`
		INSERT INTO employees (name, team_id) VALUES (?, ?)
	`, emp.Name, emp.TeamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	emp.ID = id

	c.JSON(http.StatusCreated, emp)
}

func GetEmployee(c *gin.Context) {
	id := c.Param("id")

	var emp models.Employee
	err := database.DB.QueryRow(`
		SELECT id, name, team_id, created_at, updated_at FROM employees WHERE id = ?
	`, id).Scan(&emp.ID, &emp.Name, &emp.TeamID, &emp.CreatedAt, &emp.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, emp)
}

func ListEmployees(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT id, name, team_id, created_at, updated_at FROM employees
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var employees []models.Employee
	for rows.Next() {
		var emp models.Employee
		if err := rows.Scan(&emp.ID, &emp.Name, &emp.TeamID, &emp.CreatedAt, &emp.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		employees = append(employees, emp)
	}

	c.JSON(http.StatusOK, employees)
}

func UpdateEmployee(c *gin.Context) {
	id := c.Param("id")

	var emp models.Employee
	if err := c.ShouldBindJSON(&emp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := database.DB.Exec(`
		UPDATE employees SET name = ?, team_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, emp.Name, emp.TeamID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Employee updated successfully"})
}

func DeleteEmployee(c *gin.Context) {
	id := c.Param("id")

	_, err := database.DB.Exec("DELETE FROM employees WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Employee deleted successfully"})
}
