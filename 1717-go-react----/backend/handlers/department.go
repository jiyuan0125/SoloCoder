package handlers

import (
	"net/http"

	"hospital-infection/database"
	"hospital-infection/models"

	"github.com/gin-gonic/gin"
)

type CreateDepartmentRequest struct {
	Name string `json:"Name" binding:"required"`
}

type UpdateDepartmentRequest struct {
	Name string `json:"Name" binding:"required"`
}

func CreateDepartment(c *gin.Context) {
	var req CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.Department
	database.DB.Where("name = ?", req.Name).First(&existing)
	if existing.ID > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "科室名称已存在"})
		return
	}

	dept := models.Department{
		Name: req.Name,
	}

	if err := database.DB.Create(&dept).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建科室失败"})
		return
	}

	c.JSON(http.StatusCreated, dept)
}

func GetDepartments(c *gin.Context) {
	var departments []models.Department
	if err := database.DB.Order("id asc").Find(&departments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取科室列表失败"})
		return
	}
	c.JSON(http.StatusOK, departments)
}

func GetDepartment(c *gin.Context) {
	id := c.Param("id")
	var dept models.Department
	if err := database.DB.First(&dept, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "科室不存在"})
		return
	}
	c.JSON(http.StatusOK, dept)
}

func UpdateDepartment(c *gin.Context) {
	id := c.Param("id")
	var dept models.Department
	if err := database.DB.First(&dept, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "科室不存在"})
		return
	}

	var req UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.Department
	database.DB.Where("name = ? AND id != ?", req.Name, id).First(&existing)
	if existing.ID > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "科室名称已存在"})
		return
	}

	dept.Name = req.Name
	if err := database.DB.Save(&dept).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新科室失败"})
		return
	}

	c.JSON(http.StatusOK, dept)
}

func DeleteDepartment(c *gin.Context) {
	id := c.Param("id")
	var dept models.Department
	if err := database.DB.First(&dept, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "科室不存在"})
		return
	}

	var caseCount int64
	database.DB.Model(&models.InfectionCase{}).Where("department_id = ?", id).Count(&caseCount)
	if caseCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该科室下存在感染病例，无法删除"})
		return
	}

	if err := database.DB.Delete(&dept).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除科室失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
