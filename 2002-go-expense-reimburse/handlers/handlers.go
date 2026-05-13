package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"expense-reimburse/models"
	"expense-reimburse/services"

	"github.com/gin-gonic/gin"
)

type ApiResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func SubmitReimbursement(c *gin.Context) {
	var req models.SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	rm, err := services.SubmitReimbursement(&req)
	if err != nil {
		errStr := err.Error()
		if strings.HasPrefix(errStr, "409:") {
			c.JSON(http.StatusConflict, ApiResponse{
				Code:    http.StatusConflict,
				Message: strings.TrimPrefix(errStr, "409:"),
			})
			return
		}
		c.JSON(http.StatusBadRequest, ApiResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, ApiResponse{
		Code:    http.StatusCreated,
		Message: "提交成功",
		Data:    rm,
	})
}

func UpdateReimbursement(c *gin.Context) {
	var req models.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	rm, err := services.UpdateReimbursement(&req)
	if err != nil {
		errStr := err.Error()
		if strings.HasPrefix(errStr, "409:") {
			c.JSON(http.StatusConflict, ApiResponse{
				Code:    http.StatusConflict,
				Message: strings.TrimPrefix(errStr, "409:"),
			})
			return
		}
		if strings.HasPrefix(errStr, "400:") {
			c.JSON(http.StatusBadRequest, ApiResponse{
				Code:    http.StatusBadRequest,
				Message: strings.TrimPrefix(errStr, "400:"),
			})
			return
		}
		c.JSON(http.StatusBadRequest, ApiResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ApiResponse{
		Code:    http.StatusOK,
		Message: "修改成功",
		Data:    rm,
	})
}

func ApproveReimbursement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的报销单ID",
		})
		return
	}

	var req models.ApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	rm, err := services.ApproveReimbursement(id, &req)
	if err != nil {
		errStr := err.Error()
		if errStr == "已过期请重新提交" {
			c.JSON(http.StatusBadRequest, ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "已过期请重新提交",
			})
			return
		}
		c.JSON(http.StatusBadRequest, ApiResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ApiResponse{
		Code:    http.StatusOK,
		Message: "操作成功",
		Data:    rm,
	})
}

func GetReimbursement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的报销单ID",
		})
		return
	}

	rm, err := services.GetReimbursementByID(id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			c.JSON(http.StatusNotFound, ApiResponse{
				Code:    http.StatusNotFound,
				Message: "报销单不存在",
			})
			return
		}
		c.JSON(http.StatusBadRequest, ApiResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	if rm.Status == models.StatusExpired {
		c.JSON(http.StatusOK, ApiResponse{
			Code:    http.StatusOK,
			Message: "已过期请重新提交",
			Data:    rm,
		})
		return
	}

	c.JSON(http.StatusOK, ApiResponse{
		Code:    http.StatusOK,
		Message: "查询成功",
		Data:    rm,
	})
}

func ListReimbursements(c *gin.Context) {
	employeeID := c.Query("employee_id")

	list, err := services.GetReimbursements(employeeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ApiResponse{
		Code:    http.StatusOK,
		Message: "查询成功",
		Data:    list,
	})
}

func GetStatusHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的报销单ID",
		})
		return
	}

	history, err := services.GetStatusHistory(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ApiResponse{
		Code:    http.StatusOK,
		Message: "查询成功",
		Data:    history,
	})
}
