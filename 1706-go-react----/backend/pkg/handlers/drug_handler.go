package handlers

import (
	"hospital-pharmacy/pkg/models"
	"hospital-pharmacy/pkg/services"
	"hospital-pharmacy/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateDrug(c *gin.Context) {
	var drug models.Drug
	if err := c.ShouldBindJSON(&drug); err != nil {
		utils.BadRequestResponse(c, "请求参数错误")
		return
	}

	if err := services.CreateDrug(&drug); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, drug)
}

func UpdateDrug(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(c, "ID参数错误")
		return
	}

	var drug models.Drug
	if err := c.ShouldBindJSON(&drug); err != nil {
		utils.BadRequestResponse(c, "请求参数错误")
		return
	}
	drug.ID = uint(id)

	if err := services.UpdateDrug(&drug); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, drug)
}

func GetDrug(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(c, "ID参数错误")
		return
	}

	drug, err := services.GetDrug(uint(id))
	if err != nil {
		utils.NotFoundResponse(c, "药品不存在")
		return
	}

	utils.SuccessResponse(c, drug)
}

func ListDrugs(c *gin.Context) {
	search := c.Query("search")
	drugs, err := services.ListDrugs(search)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, drugs)
}

func DeleteDrug(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(c, "ID参数错误")
		return
	}

	if err := services.DeleteDrug(uint(id)); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, nil)
}

func CreateCategory(c *gin.Context) {
	var category models.DrugCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		utils.BadRequestResponse(c, "请求参数错误")
		return
	}

	if err := services.CreateCategory(&category); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, category)
}

func ListCategories(c *gin.Context) {
	categories, err := services.ListCategories()
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, categories)
}

func SetupDrugRoutes(r *gin.Engine) {
	drugGroup := r.Group("/api/drugs")
	{
		drugGroup.GET("", ListDrugs)
		drugGroup.GET("/:id", GetDrug)
		drugGroup.POST("", CreateDrug)
		drugGroup.PUT("/:id", UpdateDrug)
		drugGroup.DELETE("/:id", DeleteDrug)
	}

	catGroup := r.Group("/api/categories")
	{
		catGroup.GET("", ListCategories)
		catGroup.POST("", CreateCategory)
	}
}
