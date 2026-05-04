package handler

import (
	"coupon-service/model"
	"coupon-service/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BatchHandler struct {
	batchService *service.BatchService
}

func NewBatchHandler() *BatchHandler {
	return &BatchHandler{
		batchService: service.NewBatchService(),
	}
}

func (h *BatchHandler) CreateBatch(c *gin.Context) {
	var req model.CreateBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	batch, err := h.batchService.CreateBatch(&req)
	if err != nil {
		switch err {
		case service.ErrBatchNameEmpty,
			service.ErrDiscountAmountZero,
			service.ErrInvalidDateRange,
			service.ErrTotalQuantityZero,
			service.ErrLimitPerUserZero,
			service.ErrBatchNameDuplicate:
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "创建批次失败: " + err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    batch,
	})
}

func (h *BatchHandler) UpdateBatch(c *gin.Context) {
	batchIDStr := c.Param("id")
	batchID, err := strconv.ParseInt(batchIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的批次ID",
		})
		return
	}

	var req model.UpdateBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	batch, err := h.batchService.UpdateBatch(batchID, &req)
	if err != nil {
		switch err {
		case service.ErrBatchNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
		case service.ErrBatchAlreadyIssued,
			service.ErrBatchNameEmpty,
			service.ErrDiscountAmountZero,
			service.ErrInvalidDateRange,
			service.ErrLimitPerUserZero,
			service.ErrBatchNameDuplicate:
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "更新批次失败: " + err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    batch,
	})
}

func (h *BatchHandler) GetBatchByID(c *gin.Context) {
	batchIDStr := c.Param("id")
	batchID, err := strconv.ParseInt(batchIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的批次ID",
		})
		return
	}

	batch, err := h.batchService.GetBatchByID(batchID)
	if err != nil {
		if err == service.ErrBatchNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "获取批次信息失败: " + err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    batch,
	})
}

func (h *BatchHandler) GetBatchProgress(c *gin.Context) {
	batchIDStr := c.Param("id")
	batchID, err := strconv.ParseInt(batchIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的批次ID",
		})
		return
	}

	progress, err := h.batchService.GetBatchProgress(batchID)
	if err != nil {
		if err == service.ErrBatchNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "获取批次进度失败: " + err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    progress,
	})
}

func (h *BatchHandler) ListBatches(c *gin.Context) {
	batches, err := h.batchService.ListBatches()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取批次列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    batches,
	})
}
