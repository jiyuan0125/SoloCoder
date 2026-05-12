package controller

import (
	"net/http"
	"strconv"

	"medical-quality-system/internal/model"
	"medical-quality-system/internal/service"

	"github.com/gin-gonic/gin"
)

type IndicatorController struct {
	service *service.IndicatorService
}

func NewIndicatorController() *IndicatorController {
	return &IndicatorController{
		service: service.NewIndicatorService(),
	}
}

func (c *IndicatorController) Create(ctx *gin.Context) {
	var indicator model.Indicator
	if err := ctx.ShouldBindJSON(&indicator); err != nil {
		ctx.Error(err)
		return
	}

	if err := c.service.Create(&indicator); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, indicator)
}

func (c *IndicatorController) List(ctx *gin.Context) {
	indicators, err := c.service.List()
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, indicators)
}

func (c *IndicatorController) Get(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	indicator, err := c.service.Get(uint(id))
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, indicator)
}

func (c *IndicatorController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var indicator model.Indicator
	if err := ctx.ShouldBindJSON(&indicator); err != nil {
		ctx.Error(err)
		return
	}

	if err := c.service.Update(uint(id), &indicator); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

type UpdateTargetRequest struct {
	NewTarget  float64 `json:"new_target" binding:"required"`
	Reason     string  `json:"reason" binding:"required"`
	ApprovedBy string  `json:"approved_by" binding:"required"`
}

func (c *IndicatorController) UpdateTarget(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req UpdateTargetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	if err := c.service.UpdateTarget(uint(id), req.NewTarget, req.Reason, req.ApprovedBy); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "目标值更新成功，下月生效"})
}

func (c *IndicatorController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := c.service.Delete(uint(id)); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func (c *IndicatorController) CreateData(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var data model.IndicatorData
	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.Error(err)
		return
	}

	data.IndicatorID = uint(id)
	if err := c.service.CreateData(&data); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, data)
}

func (c *IndicatorController) ListData(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	dataList, err := c.service.ListData(uint(id))
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, dataList)
}

func (c *IndicatorController) GetTrend(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	months := 12
	if m := ctx.Query("months"); m != "" {
		if n, err := strconv.Atoi(m); err == nil {
			months = n
		}
	}

	trend, err := c.service.GetTrend(uint(id), months)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, trend)
}
