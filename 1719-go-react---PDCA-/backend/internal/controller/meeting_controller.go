package controller

import (
	"net/http"
	"strconv"

	"medical-quality-system/internal/model"
	"medical-quality-system/internal/service"

	"github.com/gin-gonic/gin"
)

type MeetingController struct {
	service *service.MeetingService
}

func NewMeetingController() *MeetingController {
	return &MeetingController{
		service: service.NewMeetingService(),
	}
}

type CreateMeetingRequest struct {
	Meeting     model.Meeting       `json:"meeting" binding:"required"`
	ActionItems []model.ActionItem  `json:"action_items"`
}

func (c *MeetingController) Create(ctx *gin.Context) {
	var req CreateMeetingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(err)
		return
	}

	if err := c.service.Create(&req.Meeting, req.ActionItems); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, req.Meeting)
}

func (c *MeetingController) List(ctx *gin.Context) {
	meetings, err := c.service.List()
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, meetings)
}

func (c *MeetingController) Get(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	meeting, err := c.service.Get(uint(id))
	if err != nil {
		ctx.Error(err)
		return
	}

	actionItems, _ := c.service.GetActionItems(uint(id))
	ctx.JSON(http.StatusOK, gin.H{
		"meeting":      meeting,
		"action_items": actionItems,
	})
}

func (c *MeetingController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var meeting model.Meeting
	if err := ctx.ShouldBindJSON(&meeting); err != nil {
		ctx.Error(err)
		return
	}

	if err := c.service.Update(uint(id), &meeting); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func (c *MeetingController) Delete(ctx *gin.Context) {
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
