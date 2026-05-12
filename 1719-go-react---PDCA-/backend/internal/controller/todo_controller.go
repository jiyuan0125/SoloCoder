package controller

import (
	"net/http"
	"strconv"

	"medical-quality-system/internal/model"
	"medical-quality-system/internal/service"

	"github.com/gin-gonic/gin"
)

type TodoController struct {
	service *service.TodoService
}

func NewTodoController() *TodoController {
	return &TodoController{
		service: service.NewTodoService(),
	}
}

func (c *TodoController) List(ctx *gin.Context) {
	status := model.TodoStatus(ctx.Query("status"))
	todoType := model.TodoType(ctx.Query("type"))

	todos, err := c.service.List(status, todoType)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, todos)
}

func (c *TodoController) Get(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	todo, err := c.service.Get(uint(id))
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, todo)
}

func (c *TodoController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var todo model.Todo
	if err := ctx.ShouldBindJSON(&todo); err != nil {
		ctx.Error(err)
		return
	}

	if err := c.service.Update(uint(id), &todo); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

type UpdateStatusRequest struct {
	Status model.TodoStatus `json:"status" binding:"required"`
}

func (c *TodoController) UpdateStatus(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req UpdateStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	if err := c.service.UpdateStatus(uint(id), req.Status); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "状态更新成功"})
}

func (c *TodoController) Delete(ctx *gin.Context) {
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
