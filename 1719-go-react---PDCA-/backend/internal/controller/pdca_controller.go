package controller

import (
	"net/http"
	"strconv"

	"medical-quality-system/internal/model"
	"medical-quality-system/internal/service"

	"github.com/gin-gonic/gin"
)

type PDCAController struct {
	service *service.PDCAService
}

func NewPDCAController() *PDCAController {
	return &PDCAController{
		service: service.NewPDCAService(),
	}
}

func (c *PDCAController) Create(ctx *gin.Context) {
	var pdca model.PDCA
	if err := ctx.ShouldBindJSON(&pdca); err != nil {
		ctx.Error(err)
		return
	}

	if err := c.service.Create(&pdca); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, pdca)
}

func (c *PDCAController) List(ctx *gin.Context) {
	pdcaList, err := c.service.List()
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, pdcaList)
}

func (c *PDCAController) Get(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	pdca, err := c.service.Get(uint(id))
	if err != nil {
		ctx.Error(err)
		return
	}

	phases, _ := c.service.GetPhaseDetails(uint(id))
	ctx.JSON(http.StatusOK, gin.H{
		"pdca":   pdca,
		"phases": phases,
	})
}

func (c *PDCAController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var pdca model.PDCA
	if err := ctx.ShouldBindJSON(&pdca); err != nil {
		ctx.Error(err)
		return
	}

	if err := c.service.Update(uint(id), &pdca); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func (c *PDCAController) Delete(ctx *gin.Context) {
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

func (c *PDCAController) CreatePhase(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var detail model.PDCAPhaseDetail
	if err := ctx.ShouldBindJSON(&detail); err != nil {
		ctx.Error(err)
		return
	}

	detail.PDCAID = uint(id)
	if err := c.service.CreatePhaseDetail(&detail); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, detail)
}

func (c *PDCAController) UpdatePhase(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	phase := model.PDCAPhase(ctx.Param("phase"))

	var detail model.PDCAPhaseDetail
	if err := ctx.ShouldBindJSON(&detail); err != nil {
		ctx.Error(err)
		return
	}

	if err := c.service.UpdatePhaseDetail(uint(id), phase, &detail); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "阶段更新成功"})
}

func (c *PDCAController) NextPhase(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := c.service.NextPhase(uint(id)); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "阶段推进成功"})
}

func (c *PDCAController) StartNextCycle(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	newPDCA, err := c.service.StartNextCycle(uint(id))
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, newPDCA)
}
