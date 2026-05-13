package handler

import (
	"net/http"
	"strings"
	"time"

	"coupon-system/internal/model"
	"coupon-system/internal/service"

	"github.com/gin-gonic/gin"
)

type createCouponReq struct {
	BatchID      string `json:"batch_id" binding:"required"`
	Denomination int64  `json:"denomination" binding:"required"`
	Threshold    int64  `json:"threshold" binding:"required"`
	TotalCount   int64  `json:"total_count" binding:"required"`
	LimitPerUser int64  `json:"limit_per_user" binding:"required"`
	StartTime    string `json:"start_time" binding:"required"`
	EndTime      string `json:"end_time" binding:"required"`
}

type claimCouponReq struct {
	UserID  string `json:"user_id" binding:"required"`
	BatchID string `json:"batch_id" binding:"required"`
}

type redeemCouponReq struct {
	UserCouponID int64  `json:"user_coupon_id" binding:"required"`
	OrderID      string `json:"order_id" binding:"required"`
	OrderAmount  int64  `json:"order_amount" binding:"required"`
}

type refundReq struct {
	OrderID      string `json:"order_id" binding:"required"`
	RefundAmount int64  `json:"refund_amount" binding:"required"`
}

func CreateCoupon(c *gin.Context) {
	var req createCouponReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_time 格式错误，应为 RFC3339 格式"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_time 格式错误，应为 RFC3339 格式"})
		return
	}

	coupon, err := service.CreateCoupon(&service.CreateCouponRequest{
		BatchID:      req.BatchID,
		Denomination: req.Denomination,
		Threshold:    req.Threshold,
		TotalCount:   req.TotalCount,
		LimitPerUser: req.LimitPerUser,
		StartTime:    startTime,
		EndTime:      endTime,
	})

	if err != nil {
		if strings.Contains(err.Error(), "面额") || 
		   strings.Contains(err.Error(), "门槛") || 
		   strings.Contains(err.Error(), "发行量") || 
		   strings.Contains(err.Error(), "限领") || 
		   strings.Contains(err.Error(), "时间") ||
		   strings.Contains(err.Error(), "已存在") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, coupon)
}

func ClaimCoupon(c *gin.Context) {
	var req claimCouponReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userCoupon, err := service.ClaimCoupon(&service.ClaimCouponRequest{
		UserID:  req.UserID,
		BatchID: req.BatchID,
	})

	if err != nil {
		if err.Error() == "已领完" || err.Error() == "已达领取上限" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "不存在") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userCoupon)
}

func RedeemCoupon(c *gin.Context) {
	var req redeemCouponReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record, err := service.RedeemCoupon(&service.RedeemCouponRequest{
		UserCouponID: req.UserCouponID,
		OrderID:      req.OrderID,
		OrderAmount:  req.OrderAmount,
	})

	if err != nil {
		if strings.Contains(err.Error(), "不在有效期内") || 
		   strings.Contains(err.Error(), "不满足使用门槛") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "已核销" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "不存在") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, record)
}

func PartialRefund(c *gin.Context) {
	var req refundReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refundDiscount, err := service.PartialRefund(&service.RefundRequest{
		OrderID:      req.OrderID,
		RefundAmount: req.RefundAmount,
	})

	if err != nil {
		if strings.Contains(err.Error(), "不存在") || strings.Contains(err.Error(), "无效") || strings.Contains(err.Error(), "已退款") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order_id":        req.OrderID,
		"refund_discount": refundDiscount,
		"message":         "退款成功",
	})
}

func GetStats(c *gin.Context) {
	batchID := c.Query("batch_id")
	if batchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 batch_id 参数"})
		return
	}

	stats, err := service.GetStatsByBatch(batchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func GetRedemptionRecords(c *gin.Context) {
	batchID := c.Query("batch_id")
	if batchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 batch_id 参数"})
		return
	}

	records, err := service.GetRedemptionRecords(batchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if records == nil {
		records = make([]*model.RedemptionRecord, 0)
	}

	c.JSON(http.StatusOK, records)
}
